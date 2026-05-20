package services

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"imtiaz-portfolio/internal/models"
)

// ScraperConfig holds all external-API credentials and search terms.
type ScraperConfig struct {
	Keywords        []string
	AdzunaAppID     string
	AdzunaAppKey    string
	RapidAPIKey     string
	DisabledSources map[string]bool // source names that should be skipped
}

// JobScraperService fetches job listings from multiple external sources.
type JobScraperService struct {
	db         *gorm.DB
	jobSvc     *JobService
	profileSvc *ProfileService
	emailSvc   *EmailService
	cfg        ScraperConfig
	client     *http.Client
}

// scraperEntry is a named source + its fetch function.
type scraperEntry struct {
	name string
	fn   func(context.Context) ([]models.JobPost, error)
}

// NewJobScraperService wires up the scraper with credentials and search config.
func NewJobScraperService(db *gorm.DB, jobSvc *JobService, profileSvc *ProfileService, emailSvc *EmailService, cfg ScraperConfig) *JobScraperService {
	return &JobScraperService{
		db:         db,
		jobSvc:     jobSvc,
		profileSvc: profileSvc,
		emailSvc:   emailSvc,
		cfg:        cfg,
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

// SourceList returns every known source and whether it is configured (API key set).
func (s *JobScraperService) SourceList() []map[string]interface{} {
	s.reloadFromDB() // keep status accurate after dashboard key changes
	return []map[string]interface{}{
		{"name": "remoteok",      "configured": true},
		{"name": "remotive",      "configured": true},
		{"name": "weworkremotely","configured": true},
		{"name": "arbeitnow",     "configured": true},
		{"name": "jobicy",        "configured": true},
		{"name": "himalayas",     "configured": true},
		{"name": "themuse",       "configured": true},
		{"name": "remoteco",      "configured": true},
		{"name": "jobspresso",    "configured": true},
		{"name": "authenticjobs", "configured": true},
		{"name": "adzuna",        "configured": s.cfg.AdzunaAppID != "" && s.cfg.AdzunaAppKey != ""},
		{"name": "jsearch",       "configured": s.cfg.RapidAPIKey != ""},
	}
}

// ScrapeAll runs all scrapers concurrently, then saves + logs results per source.
func (s *JobScraperService) ScrapeAll(ctx context.Context) {
	s.reloadFromDB()

	// Build full scraper list
	all := []scraperEntry{
		{"remoteok",       s.scrapeRemoteOK},
		{"remotive",       s.scrapeRemotive},
		{"weworkremotely", s.scrapeWWR},
		{"arbeitnow",      s.scrapeArbeitnow},
		{"jobicy",         s.scrapeJobicy},
		{"himalayas",      s.scrapeHimalayas},
		{"themuse",        s.scrapeTheMuse},
		{"remoteco",       func(ctx context.Context) ([]models.JobPost, error) {
			return s.scrapeRSSFeed(ctx, "remoteco", "https://remote.co/feed")
		}},
		{"jobspresso",     func(ctx context.Context) ([]models.JobPost, error) {
			return s.scrapeRSSFeed(ctx, "jobspresso", "https://jobspresso.co/feed/?post_type=job_listing")
		}},
		{"authenticjobs",  func(ctx context.Context) ([]models.JobPost, error) {
			return s.scrapeRSSFeed(ctx, "authenticjobs", "https://authenticjobs.com/rss/custom.php?remote=1")
		}},
	}
	if s.cfg.AdzunaAppID != "" && s.cfg.AdzunaAppKey != "" {
		all = append(all, scraperEntry{"adzuna", s.scrapeAdzuna})
	}
	if s.cfg.RapidAPIKey != "" {
		all = append(all, scraperEntry{"jsearch", s.scrapeJSearch})
	}

	// Filter out disabled sources
	var scrapers []scraperEntry
	for _, e := range all {
		if !s.cfg.DisabledSources[e.name] {
			scrapers = append(scrapers, e)
		}
	}

	// Run all scrapers concurrently, collect results
	type result struct {
		source   string
		jobs     []models.JobPost
		err      error
		duration time.Duration
	}
	results := make([]result, len(scrapers))
	var wg sync.WaitGroup
	for i, sc := range scrapers {
		wg.Add(1)
		go func(i int, e scraperEntry) {
			defer wg.Done()
			start := time.Now()
			jobs, err := e.fn(ctx)
			results[i] = result{e.name, jobs, err, time.Since(start)}
		}(i, sc)
	}
	wg.Wait()

	// Save + log each source sequentially to avoid DB race conditions
	totalAdded := 0
	for _, r := range results {
		added := 0
		errMsg := ""
		if r.err != nil {
			errMsg = r.err.Error()
			log.Printf("[scraper] %s failed: %v", r.source, r.err)
		} else {
			log.Printf("[scraper] %s: %d jobs fetched", r.source, len(r.jobs))
			added, _ = s.jobSvc.SaveJobs(r.jobs)
		}
		s.db.Create(&models.JobScrapeLog{
			Source:     r.source,
			ScrapedAt:  time.Now(),
			JobsFound:  len(r.jobs),
			JobsAdded:  added,
			ErrorMsg:   errMsg,
			DurationMs: r.duration.Milliseconds(),
		})
		totalAdded += added
	}

	s.sendAlertIfNeeded(totalAdded)
}

// sendAlertIfNeeded emails the admin when enough new jobs were found.
func (s *JobScraperService) sendAlertIfNeeded(newJobs int) {
	if s.emailSvc == nil || s.profileSvc == nil {
		return
	}
	settings, err := s.jobSvc.GetScraperSettings()
	if err != nil || settings == nil || !settings.AlertEnabled {
		return
	}
	threshold := settings.AlertMinNewJobs
	if threshold <= 0 {
		threshold = 5
	}
	if newJobs < threshold {
		return
	}
	profile, err := s.profileSvc.Get()
	if err != nil || profile == nil || profile.Email == "" {
		return
	}
	subject := fmt.Sprintf("🔔 %d new matching jobs found", newJobs)
	body := fmt.Sprintf(
		"Hi %s,\n\nYour job scraper just found %d new jobs matching your keywords.\n\nLog in to your dashboard to review them.",
		profile.FullName, newJobs,
	)
	if err := s.emailSvc.Send(profile.FullName, profile.Email, "", subject, body, nil); err != nil {
		log.Printf("[scraper] alert email failed: %v", err)
	}
}

// reloadFromDB refreshes s.cfg from the scraper_settings table and the
// job_keywords table before each scrape run.  Called at the top of ScrapeAll
// and SourceList so the dashboard is always the source of truth.
func (s *JobScraperService) reloadFromDB() {
	settings, err := s.jobSvc.GetScraperSettings()
	if err == nil && settings != nil {
		s.cfg.AdzunaAppID  = settings.AdzunaAppID
		s.cfg.AdzunaAppKey = settings.AdzunaAppKey
		s.cfg.RapidAPIKey  = settings.RapidAPIKey

		// Parse disabled sources into a lookup map
		disabled := make(map[string]bool)
		for _, src := range strings.Split(settings.DisabledSources, ",") {
			src = strings.TrimSpace(src)
			if src != "" {
				disabled[src] = true
			}
		}
		s.cfg.DisabledSources = disabled
	}

	if kws, err := s.jobSvc.GetKeywords(); err == nil && len(kws) > 0 {
		s.cfg.Keywords = kws
	}
}

// ScrapeIntervalHours reads the configured interval from DB (default 6).
func (s *JobScraperService) ScrapeIntervalHours() int {
	settings, err := s.jobSvc.GetScraperSettings()
	if err != nil || settings == nil || settings.ScrapeIntervalHours <= 0 {
		return 6
	}
	return settings.ScrapeIntervalHours
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// keyword returns the primary search keyword (first in list, or empty string).
func (s *JobScraperService) keyword() string {
	if len(s.cfg.Keywords) > 0 {
		return s.cfg.Keywords[0]
	}
	return ""
}

// keywordsJoined joins all keywords with a space for full-text search params.
func (s *JobScraperService) keywordsJoined() string {
	return strings.Join(s.cfg.Keywords, " ")
}

func (s *JobScraperService) get(ctx context.Context, rawURL string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; personal-dashboard/1.0)")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, rawURL)
	}
	return resp, nil
}

// ─── RemoteOK ────────────────────────────────────────────────────────────────

type remoteOKJob struct {
	Slug        string   `json:"slug"`
	Epoch       int64    `json:"epoch"`
	Date        string   `json:"date"`
	Company     string   `json:"company"`
	Position    string   `json:"position"`
	Tags        []string `json:"tags"`
	Location    string   `json:"location"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
}

func (s *JobScraperService) scrapeRemoteOK(ctx context.Context) ([]models.JobPost, error) {
	resp, err := s.get(ctx, "https://remoteok.com/api", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	now := time.Now()
	var jobs []models.JobPost
	for i, item := range raw {
		if i == 0 {
			continue // first element is metadata
		}
		var j remoteOKJob
		if err := json.Unmarshal(item, &j); err != nil || j.Position == "" {
			continue
		}
		jobURL := j.URL
		if jobURL == "" && j.Slug != "" {
			jobURL = "https://remoteok.com/remote-jobs/" + j.Slug
		}
		if jobURL == "" {
			continue
		}
		postedAt := now
		if j.Date != "" {
			if t, err := time.Parse(time.RFC3339, j.Date); err == nil {
				postedAt = t
			}
		} else if j.Epoch != 0 {
			postedAt = time.Unix(j.Epoch, 0)
		}
		jobs = append(jobs, models.JobPost{
			Source: "remoteok", Title: j.Position, Company: j.Company,
			Location: j.Location, JobType: jobTypeFromTags(j.Tags),
			URL: jobURL, Description: stripHTML(j.Description),
			PostedAt: postedAt, ScrapedAt: now,
		})
	}
	return jobs, nil
}

// ─── Remotive ────────────────────────────────────────────────────────────────

func (s *JobScraperService) scrapeRemotive(ctx context.Context) ([]models.JobPost, error) {
	resp, err := s.get(ctx, "https://remotive.com/api/remote-jobs", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Jobs []struct {
			URL             string `json:"url"`
			Title           string `json:"title"`
			CompanyName     string `json:"company_name"`
			Location        string `json:"candidate_required_location"`
			JobType         string `json:"job_type"`
			PublicationDate string `json:"publication_date"`
			Description     string `json:"description"`
		} `json:"jobs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	now := time.Now()
	var jobs []models.JobPost
	for _, j := range payload.Jobs {
		if j.URL == "" || j.Title == "" {
			continue
		}
		postedAt := now
		if t, err := time.Parse(time.RFC3339, j.PublicationDate); err == nil {
			postedAt = t
		}
		jobs = append(jobs, models.JobPost{
			Source: "remotive", Title: j.Title, Company: j.CompanyName,
			Location: j.Location, JobType: normalizeJobType(j.JobType),
			URL: j.URL, Description: stripHTML(j.Description),
			PostedAt: postedAt, ScrapedAt: now,
		})
	}
	return jobs, nil
}

// ─── We Work Remotely (RSS) ──────────────────────────────────────────────────

func (s *JobScraperService) scrapeWWR(ctx context.Context) ([]models.JobPost, error) {
	resp, err := s.get(ctx, "https://weworkremotely.com/remote-jobs.rss",
		map[string]string{"Accept": "application/rss+xml, application/xml, text/xml"})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var feed struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Items []rssItem `xml:"item"`
		} `xml:"channel"`
	}
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, err
	}

	now := time.Now()
	var jobs []models.JobPost
	for _, item := range feed.Channel.Items {
		jobURL := item.Link
		if jobURL == "" {
			jobURL = item.GUID
		}
		if jobURL == "" {
			continue
		}
		title, company := parseWWRTitle(item.Title)
		jobs = append(jobs, models.JobPost{
			Source: "weworkremotely", Title: title, Company: company,
			Location: item.Region, JobType: "full-time",
			URL: jobURL, PostedAt: parseRSSDate(item.PubDate, now), ScrapedAt: now,
		})
	}
	return jobs, nil
}

// ─── Arbeitnow ───────────────────────────────────────────────────────────────

func (s *JobScraperService) scrapeArbeitnow(ctx context.Context) ([]models.JobPost, error) {
	resp, err := s.get(ctx, "https://www.arbeitnow.com/api/job-board-api?page=1",
		map[string]string{"Accept": "application/json"})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Data []struct {
			URL         string   `json:"url"`
			Title       string   `json:"title"`
			CompanyName string   `json:"company_name"`
			Location    string   `json:"location"`
			JobTypes    []string `json:"job_types"`
			Description string   `json:"description"`
			CreatedAt   int64    `json:"created_at"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	now := time.Now()
	var jobs []models.JobPost
	for _, j := range payload.Data {
		if j.URL == "" || j.Title == "" {
			continue
		}
		jobType := "full-time"
		if len(j.JobTypes) > 0 {
			jobType = normalizeJobType(strings.ToLower(j.JobTypes[0]))
		}
		postedAt := now
		if j.CreatedAt != 0 {
			postedAt = time.Unix(j.CreatedAt, 0)
		}
		jobs = append(jobs, models.JobPost{
			Source: "arbeitnow", Title: j.Title, Company: j.CompanyName,
			Location: j.Location, JobType: jobType,
			URL: j.URL, Description: stripHTML(j.Description),
			PostedAt: postedAt, ScrapedAt: now,
		})
	}
	return jobs, nil
}

// ─── Jobicy ──────────────────────────────────────────────────────────────────

func (s *JobScraperService) scrapeJobicy(ctx context.Context) ([]models.JobPost, error) {
	u := "https://jobicy.com/api/v2/remote-jobs?count=50&geo=worldwide"
	if kw := s.keyword(); kw != "" {
		u += "&tag=" + url.QueryEscape(kw)
	}
	resp, err := s.get(ctx, u, map[string]string{"Accept": "application/json"})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Jobs []struct {
			URL            string   `json:"url"`
			JobTitle       string   `json:"jobTitle"`
			CompanyName    string   `json:"companyName"`
			JobGeo         string   `json:"jobGeo"`
			JobType        []string `json:"jobType"`
			JobDescription string   `json:"jobDescription"`
			PubDate        string   `json:"pubDate"`
		} `json:"jobs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	now := time.Now()
	var jobs []models.JobPost
	for _, j := range payload.Jobs {
		if j.URL == "" || j.JobTitle == "" {
			continue
		}
		jobType := "full-time"
		if len(j.JobType) > 0 {
			jobType = normalizeJobType(strings.ToLower(j.JobType[0]))
		}
		postedAt := now
		// Jobicy uses "2006-01-02 15:04:05" format
		if t, err := time.Parse("2006-01-02 15:04:05", j.PubDate); err == nil {
			postedAt = t
		}
		jobs = append(jobs, models.JobPost{
			Source: "jobicy", Title: j.JobTitle, Company: j.CompanyName,
			Location: j.JobGeo, JobType: jobType,
			URL: j.URL, Description: stripHTML(j.JobDescription),
			PostedAt: postedAt, ScrapedAt: now,
		})
	}
	return jobs, nil
}

// ─── Himalayas ───────────────────────────────────────────────────────────────

func (s *JobScraperService) scrapeHimalayas(ctx context.Context) ([]models.JobPost, error) {
	u := "https://himalayas.app/jobs/api?limit=100"
	if kw := s.keywordsJoined(); kw != "" {
		u += "&q=" + url.QueryEscape(kw)
	}
	resp, err := s.get(ctx, u, map[string]string{"Accept": "application/json"})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Jobs []struct {
			Title          string   `json:"title"`
			ApplicationURL string   `json:"applicationUrl"`
			URL            string   `json:"url"`
			JobTypes       []string `json:"jobTypes"`
			Location       string   `json:"location"`
			PublishedAt    string   `json:"publishedAt"`
			Description    string   `json:"description"`
			Company        struct {
				Name string `json:"name"`
			} `json:"company"`
		} `json:"jobs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	now := time.Now()
	var jobs []models.JobPost
	for _, j := range payload.Jobs {
		jobURL := j.ApplicationURL
		if jobURL == "" {
			jobURL = j.URL
		}
		if jobURL == "" || j.Title == "" {
			continue
		}
		jobType := "full-time"
		if len(j.JobTypes) > 0 {
			jobType = normalizeJobType(strings.ToLower(j.JobTypes[0]))
		}
		postedAt := now
		if t, err := time.Parse(time.RFC3339, j.PublishedAt); err == nil {
			postedAt = t
		}
		jobs = append(jobs, models.JobPost{
			Source: "himalayas", Title: j.Title, Company: j.Company.Name,
			Location: j.Location, JobType: jobType,
			URL: jobURL, Description: stripHTML(j.Description),
			PostedAt: postedAt, ScrapedAt: now,
		})
	}
	return jobs, nil
}

// ─── The Muse ────────────────────────────────────────────────────────────────

func (s *JobScraperService) scrapeTheMuse(ctx context.Context) ([]models.JobPost, error) {
	resp, err := s.get(ctx,
		"https://www.themuse.com/api/public/v2/jobs?category=Software%20Engineer&level=Mid%20Level&level=Senior%20Level&page=0",
		map[string]string{"Accept": "application/json"})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Results []struct {
			ID              int    `json:"id"`
			Name            string `json:"name"`
			PublicationDate string `json:"publication_date"`
			Company         struct {
				Name string `json:"name"`
			} `json:"company"`
			Locations []struct {
				Name string `json:"name"`
			} `json:"locations"`
			Refs struct {
				LandingPage string `json:"landing_page"`
			} `json:"refs"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	now := time.Now()
	var jobs []models.JobPost
	for _, j := range payload.Results {
		if j.Refs.LandingPage == "" || j.Name == "" {
			continue
		}
		location := "Remote"
		if len(j.Locations) > 0 {
			location = j.Locations[0].Name
		}
		postedAt := now
		if t, err := time.Parse(time.RFC3339, j.PublicationDate); err == nil {
			postedAt = t
		}
		jobs = append(jobs, models.JobPost{
			Source: "themuse", Title: j.Name, Company: j.Company.Name,
			Location: location, JobType: "full-time",
			URL: j.Refs.LandingPage, PostedAt: postedAt, ScrapedAt: now,
		})
	}
	return jobs, nil
}

// ─── Generic RSS (Remote.co, Jobspresso, Authentic Jobs) ─────────────────────

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	GUID    string `xml:"guid"`
	PubDate string `xml:"pubDate"`
	Region  string `xml:"region"` // WWR-specific; empty for other feeds
}

func (s *JobScraperService) scrapeRSSFeed(ctx context.Context, source, feedURL string) ([]models.JobPost, error) {
	resp, err := s.get(ctx, feedURL,
		map[string]string{"Accept": "application/rss+xml, application/xml, text/xml"})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var feed struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Items []rssItem `xml:"item"`
		} `xml:"channel"`
	}
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, err
	}

	now := time.Now()
	var jobs []models.JobPost
	for _, item := range feed.Channel.Items {
		jobURL := item.Link
		if jobURL == "" {
			jobURL = item.GUID
		}
		if jobURL == "" || item.Title == "" {
			continue
		}
		jobs = append(jobs, models.JobPost{
			Source:    source,
			Title:     item.Title,
			URL:       jobURL,
			JobType:   "full-time",
			PostedAt:  parseRSSDate(item.PubDate, now),
			ScrapedAt: now,
		})
	}
	return jobs, nil
}

// ─── Adzuna (free API key required) ──────────────────────────────────────────

func (s *JobScraperService) scrapeAdzuna(ctx context.Context) ([]models.JobPost, error) {
	what := url.QueryEscape(s.keywordsJoined())
	if what == "" {
		what = "remote+developer"
	}
	// Search US + GB to maximise coverage
	var all []models.JobPost
	for _, country := range []string{"us", "gb"} {
		u := fmt.Sprintf(
			"https://api.adzuna.com/v1/api/jobs/%s/search/1?app_id=%s&app_key=%s&results_per_page=20&what=%s&content-type=application/json",
			country, s.cfg.AdzunaAppID, s.cfg.AdzunaAppKey, what,
		)
		resp, err := s.get(ctx, u, map[string]string{"Accept": "application/json"})
		if err != nil {
			log.Printf("[scraper] adzuna/%s: %v", country, err)
			continue
		}

		var payload struct {
			Results []struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Created     string `json:"created"`
				Description string `json:"description"`
				RedirectURL string `json:"redirect_url"`
				Company     struct {
					DisplayName string `json:"display_name"`
				} `json:"company"`
				Location struct {
					DisplayName string `json:"display_name"`
				} `json:"location"`
				ContractType string `json:"contract_type"`
			} `json:"results"`
		}
		err = json.NewDecoder(resp.Body).Decode(&payload)
		resp.Body.Close()
		if err != nil {
			continue
		}

		now := time.Now()
		for _, j := range payload.Results {
			if j.RedirectURL == "" || j.Title == "" {
				continue
			}
			postedAt := now
			if t, err := time.Parse(time.RFC3339, j.Created); err == nil {
				postedAt = t
			}
			all = append(all, models.JobPost{
				Source: "adzuna", Title: j.Title, Company: j.Company.DisplayName,
				Location: j.Location.DisplayName, JobType: normalizeAdzunaContract(j.ContractType),
				URL: j.RedirectURL, Description: stripHTML(j.Description),
				PostedAt: postedAt, ScrapedAt: now,
			})
		}
	}
	return all, nil
}

func normalizeAdzunaContract(c string) string {
	switch strings.ToLower(c) {
	case "permanent":
		return "full-time"
	case "contract":
		return "contract"
	case "part_time":
		return "part-time"
	default:
		return "full-time"
	}
}

// ─── JSearch via RapidAPI (LinkedIn + Indeed + Glassdoor + ZipRecruiter) ─────

func (s *JobScraperService) scrapeJSearch(ctx context.Context) ([]models.JobPost, error) {
	q := s.keywordsJoined()
	if q == "" {
		q = "developer"
	}
	q += " remote"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://jsearch.p.rapidapi.com/search?query="+url.QueryEscape(q)+"&num_pages=1&date_posted=week", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; personal-dashboard/1.0)")
	req.Header.Set("X-RapidAPI-Key", s.cfg.RapidAPIKey)
	req.Header.Set("X-RapidAPI-Host", "jsearch.p.rapidapi.com")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jsearch HTTP %d", resp.StatusCode)
	}

	var payload struct {
		Data []struct {
			JobTitle                 string `json:"job_title"`
			EmployerName             string `json:"employer_name"`
			JobCity                  string `json:"job_city"`
			JobCountry               string `json:"job_country"`
			JobEmploymentType        string `json:"job_employment_type"`
			JobApplyLink             string `json:"job_apply_link"`
			JobDescription           string `json:"job_description"`
			JobPostedAtDatetimeUTC   string `json:"job_posted_at_datetime_utc"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	now := time.Now()
	var jobs []models.JobPost
	for _, j := range payload.Data {
		if j.JobApplyLink == "" || j.JobTitle == "" {
			continue
		}
		location := strings.TrimSpace(j.JobCity + ", " + j.JobCountry)
		location = strings.Trim(location, ", ")
		postedAt := now
		if t, err := time.Parse(time.RFC3339, j.JobPostedAtDatetimeUTC); err == nil {
			postedAt = t
		}
		jobs = append(jobs, models.JobPost{
			Source: "jsearch", Title: j.JobTitle, Company: j.EmployerName,
			Location: location, JobType: normalizeJobType(strings.ToLower(j.JobEmploymentType)),
			URL: j.JobApplyLink, Description: stripHTML(j.JobDescription),
			PostedAt: postedAt, ScrapedAt: now,
		})
	}
	return jobs, nil
}

// ─── Shared helpers ───────────────────────────────────────────────────────────

// parseWWRTitle splits "Category: Job Title at Company" into title and company.
func parseWWRTitle(raw string) (title, company string) {
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, ": "); idx != -1 {
		raw = raw[idx+2:]
	}
	if idx := strings.LastIndex(raw, " at "); idx != -1 {
		title = strings.TrimSpace(raw[:idx])
		company = strings.TrimSpace(raw[idx+4:])
	} else {
		title = raw
	}
	return
}

var rssDateFormats = []string{
	time.RFC1123Z,
	time.RFC1123,
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05 MST",
}

func parseRSSDate(s string, fallback time.Time) time.Time {
	s = strings.TrimSpace(s)
	for _, layout := range rssDateFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return fallback
}

func jobTypeFromTags(tags []string) string {
	for _, t := range tags {
		switch strings.ToLower(t) {
		case "contract", "contractor":
			return "contract"
		case "part-time", "part_time":
			return "part-time"
		case "freelance":
			return "freelance"
		}
	}
	return "full-time"
}

func normalizeJobType(t string) string {
	switch t {
	case "full_time", "fulltime", "full-time", "permanent":
		return "full-time"
	case "part_time", "parttime", "part-time":
		return "part-time"
	case "contract", "contractor":
		return "contract"
	case "freelance":
		return "freelance"
	default:
		return t
	}
}

func stripHTML(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
			b.WriteRune(' ')
		case !inTag:
			b.WriteRune(r)
		}
	}
	result := strings.Join(strings.Fields(b.String()), " ")
	if len(result) > 500 {
		return result[:500] + "..."
	}
	return result
}
