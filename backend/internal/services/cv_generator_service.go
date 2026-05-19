// ============================================================
// FILE: internal/services/cv_generator_service.go
// LAYOUT:   Exact match of uploaded CV — Times serif, A4 20mm margins,
//           navy section headings, two-column entries, bullet achievements.
// LAST UPDATED: 2026-05-19 — full rewrite from spec
// ============================================================

package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"imtiaz-portfolio/internal/models"

	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

// ─── Service ─────────────────────────────────────────────────────

type CVGeneratorService struct {
	db        *gorm.DB
	profile   *ProfileService
	uploadDir string
}

func NewCVGeneratorService(db *gorm.DB, profile *ProfileService, uploadDir string) *CVGeneratorService {
	return &CVGeneratorService{db: db, profile: profile, uploadDir: uploadDir}
}

// ─── Custom request types (used by the CV Builder form) ──────────

type CVExpEntry struct {
	Role           string   `json:"role"`
	Company        string   `json:"company"`
	CompanyWebsite string   `json:"company_website"`
	StartDate      string   `json:"start_date"`
	EndDate        string   `json:"end_date"`
	IsCurrent      bool     `json:"is_current"`
	Location       string   `json:"location"`
	Description    string   `json:"description"`
	Achievements   []string `json:"achievements"`
	TechUsed       []string `json:"tech_used"`
}

type CVProjEntry struct {
	Name      string   `json:"name"`
	Store     string   `json:"store"`
	StartDate string   `json:"start_date"`
	EndDate   string   `json:"end_date"`
	Features  []string `json:"features"`
}

type CVSkillRow struct {
	Category string `json:"category"`
	Names    string `json:"names"` // comma-separated
}

type CVEducationEntry struct {
	Degree      string `json:"degree"`
	Institution string `json:"institution"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Location    string `json:"location"`
	GPA         string `json:"gpa"`
}

type CVLanguageEntry struct {
	Language string `json:"language"`
	Level    string `json:"level"`
}

type CVReferenceEntry struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Department  string `json:"department"`
	Institution string `json:"institution"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

type CVGenerateRequest struct {
	FullName    string             `json:"full_name"`
	Title       string             `json:"title"`
	Email       string             `json:"email"`
	Phone       string             `json:"phone"`
	Location    string             `json:"location"`
	LinkedIn    string             `json:"linkedin"`
	GitHub      string             `json:"github"`
	WhatsApp    string             `json:"whatsapp"`
	Summary     string             `json:"summary"`
	Experiences []CVExpEntry       `json:"experiences"`
	Projects    []CVProjEntry      `json:"projects"`
	Skills      []CVSkillRow       `json:"skills"`
	Education   []CVEducationEntry `json:"education"`
	Languages   []CVLanguageEntry  `json:"languages"`
	References  []CVReferenceEntry `json:"references"`
}

// ─── Internal document model ─────────────────────────────────────
// Both Generate() and GenerateFromRequest() convert their input
// into this model, then pass it to renderCV().

type cvDoc struct {
	FullName    string
	Title       string
	Email       string
	Phone       string
	Location    string
	LinkedIn    string
	GitHub      string
	WhatsApp    string
	Summary     string
	Experiences []cvDocExp
	Projects    []cvDocProj
	Skills      []cvDocSkill
	Education   []cvDocEdu
	Languages   []cvDocLang
	References  []cvDocRef
}

type cvDocExp struct {
	Role           string
	Company        string
	CompanyWebsite string
	StartDate      string
	EndDate        string
	IsCurrent      bool
	Location       string
	Description    string
	Achievements   []string
}

type cvDocProj struct {
	Name      string
	Store     string
	StartDate string
	EndDate   string
	Features  []string
}

type cvDocSkill struct {
	Category string
	Names    []string
}

type cvDocEdu struct {
	Degree      string
	Institution string
	StartDate   string
	EndDate     string
	Location    string
	GPA         string
}

type cvDocLang struct {
	Language string
	Level    string
}

type cvDocRef struct {
	Name        string
	Title       string
	Department  string
	Institution string
	Email       string
	Phone       string
}

// ─── Generate from DB ─────────────────────────────────────────────

func (s *CVGeneratorService) Generate() (*models.CVFile, error) {
	profile, err := s.profile.Get()
	if err != nil || profile == nil {
		return nil, fmt.Errorf("profile not available")
	}

	var skills []models.Skill
	if err := s.db.Order("category ASC, sort_order ASC").Find(&skills).Error; err != nil {
		return nil, fmt.Errorf("load skills: %w", err)
	}

	var experiences []models.Experience
	if err := s.db.Order("sort_order ASC, id DESC").Find(&experiences).Error; err != nil {
		return nil, fmt.Errorf("load experience: %w", err)
	}

	var projects []models.Project
	s.db.Where("status = 'live'").Order("featured DESC, sort_order ASC").Find(&projects)

	doc := profileToDoc(profile, skills, experiences, projects)
	return s.savePDF(renderCV(doc), doc.FullName)
}

// ─── Generate from form request ──────────────────────────────────

func (s *CVGeneratorService) GenerateFromRequest(req *CVGenerateRequest) (*models.CVFile, error) {
	doc := requestToDoc(req)
	return s.savePDF(renderCV(doc), doc.FullName)
}

func (s *CVGeneratorService) savePDF(pdf *gofpdf.Fpdf, name string) (*models.CVFile, error) {
	cvDir := filepath.Join(s.uploadDir, "cv")
	if err := os.MkdirAll(cvDir, 0755); err != nil {
		return nil, fmt.Errorf("create cv dir: %w", err)
	}

	timestamp := time.Now().Unix()
	fileName := fmt.Sprintf("generated_%d.pdf", timestamp)
	fullPath := filepath.Join(cvDir, fileName)

	if err := pdf.OutputFileAndClose(fullPath); err != nil {
		return nil, fmt.Errorf("write pdf: %w", err)
	}

	info, _ := os.Stat(fullPath)
	var size int64
	if info != nil {
		size = info.Size()
	}

	publicPath := fmt.Sprintf("/uploads/cv/%s", fileName)
	displayName := fmt.Sprintf("%s — Resume.pdf", strings.TrimSpace(name))
	if strings.TrimSpace(name) == "" {
		displayName = fileName
	}

	return s.profile.AddCVHistory(publicPath, displayName, size, "generated", false)
}

// ─── Converters ───────────────────────────────────────────────────

func profileToDoc(p *models.Profile, skills []models.Skill, exps []models.Experience, projs []models.Project) cvDoc {
	doc := cvDoc{
		FullName: p.FullName,
		Title:    p.Title,
		Email:    p.Email,
		Phone:    p.Phone,
		Location: p.Location,
		LinkedIn: p.LinkedIn,
		GitHub:   p.GitHub,
		WhatsApp: p.WhatsApp,
		Summary:  strings.TrimSpace(p.Bio),
	}
	if doc.Summary == "" {
		doc.Summary = strings.TrimSpace(p.ShortBio)
	}

	sort.SliceStable(exps, func(i, j int) bool {
		if exps[i].SortOrder != exps[j].SortOrder {
			return exps[i].SortOrder < exps[j].SortOrder
		}
		return exps[i].ID > exps[j].ID
	})
	for _, e := range exps {
		doc.Experiences = append(doc.Experiences, cvDocExp{
			Role:         e.Role,
			Company:      e.Company,
			StartDate:    e.StartDate,
			EndDate:      e.EndDate,
			IsCurrent:    e.IsCurrent,
			Location:     e.Location,
			Description:  e.Description,
			Achievements: []string(e.Achievements),
		})
	}

	for _, proj := range projs {
		store := ""
		if proj.AppStoreURL != "" && proj.PlayStoreURL != "" {
			store = "App Store & Google Play"
		} else if proj.AppStoreURL != "" {
			store = "App Store"
		} else if proj.PlayStoreURL != "" {
			store = "Google Play"
		}
		doc.Projects = append(doc.Projects, cvDocProj{
			Name:     proj.Name,
			Store:    store,
			Features: []string(proj.Features),
		})
	}

	grouped := map[string][]string{}
	for _, sk := range skills {
		grouped[sk.Category] = append(grouped[sk.Category], sk.Name)
	}
	catOrder := []string{"languages", "framework", "core", "state_management", "backend", "payments", "tools", "architecture", "distribution"}
	seen := map[string]bool{}
	for _, cat := range catOrder {
		if names, ok := grouped[cat]; ok {
			doc.Skills = append(doc.Skills, cvDocSkill{Category: cat, Names: names})
			seen[cat] = true
		}
	}
	var extra []string
	for k := range grouped {
		if !seen[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	for _, k := range extra {
		doc.Skills = append(doc.Skills, cvDocSkill{Category: k, Names: grouped[k]})
	}

	return doc
}

func requestToDoc(req *CVGenerateRequest) cvDoc {
	doc := cvDoc{
		FullName: req.FullName,
		Title:    req.Title,
		Email:    req.Email,
		Phone:    req.Phone,
		Location: req.Location,
		LinkedIn: req.LinkedIn,
		GitHub:   req.GitHub,
		WhatsApp: req.WhatsApp,
		Summary:  req.Summary,
	}

	for _, e := range req.Experiences {
		doc.Experiences = append(doc.Experiences, cvDocExp{
			Role:           e.Role,
			Company:        e.Company,
			CompanyWebsite: e.CompanyWebsite,
			StartDate:      e.StartDate,
			EndDate:        e.EndDate,
			IsCurrent:      e.IsCurrent,
			Location:       e.Location,
			Description:    e.Description,
			Achievements:   e.Achievements,
		})
	}

	for _, proj := range req.Projects {
		doc.Projects = append(doc.Projects, cvDocProj{
			Name:      proj.Name,
			Store:     proj.Store,
			StartDate: proj.StartDate,
			EndDate:   proj.EndDate,
			Features:  proj.Features,
		})
	}

	for _, row := range req.Skills {
		if row.Names == "" {
			continue
		}
		var names []string
		for _, n := range strings.Split(row.Names, ",") {
			if t := strings.TrimSpace(n); t != "" {
				names = append(names, t)
			}
		}
		if len(names) > 0 {
			doc.Skills = append(doc.Skills, cvDocSkill{Category: row.Category, Names: names})
		}
	}

	for _, edu := range req.Education {
		doc.Education = append(doc.Education, cvDocEdu{
			Degree:      edu.Degree,
			Institution: edu.Institution,
			StartDate:   edu.StartDate,
			EndDate:     edu.EndDate,
			Location:    edu.Location,
			GPA:         edu.GPA,
		})
	}

	for _, lang := range req.Languages {
		if lang.Language != "" {
			doc.Languages = append(doc.Languages, cvDocLang{Language: lang.Language, Level: lang.Level})
		}
	}

	for _, ref := range req.References {
		if ref.Name != "" {
			doc.References = append(doc.References, cvDocRef{
				Name:        ref.Name,
				Title:       ref.Title,
				Department:  ref.Department,
				Institution: ref.Institution,
				Email:       ref.Email,
				Phone:       ref.Phone,
			})
		}
	}

	return doc
}

// ─── Main renderer ────────────────────────────────────────────────

func renderCV(doc cvDoc) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(cvMargin, cvMargin, cvMargin)
	pdf.SetAutoPageBreak(true, cvMargin)
	pdf.AddPage()

	renderCVHeader(pdf, doc)
	renderCVSummary(pdf, doc.Summary)
	renderCVExperience(pdf, doc.Experiences)
	renderCVProjects(pdf, doc.Projects)
	renderCVSkills(pdf, doc.Skills)
	renderCVEducation(pdf, doc.Education)
	renderCVLanguages(pdf, doc.Languages)
	renderCVReferences(pdf, doc.References)

	return pdf
}

// ─── Layout constants ─────────────────────────────────────────────

const (
	cvMargin   = 20.0
	cvPageW    = 210.0
	cvContentW = cvPageW - 2*cvMargin // 170mm

	// Column split for two-column entry headers (Role | Date, Company | Location)
	cvCol1W = cvContentW * 0.65 // 110.5mm — left (role, company)
	cvCol2W = cvContentW * 0.35 // 59.5mm  — right (date, location)

	// Skills table column widths
	cvSkillLabelW = 44.0                    // mm for category label
	cvSkillValueW = cvContentW - cvSkillLabelW // 126mm for values

	// Reference two-column
	cvRefColW = cvContentW / 2 // 85mm each

	cvNavy  = "1a3a6b"
	cvBlack = "000000"
	cvGray  = "555555"
	cvBlue  = "1a56db"
)

// ─── Section renderers ────────────────────────────────────────────

func renderCVHeader(pdf *gofpdf.Fpdf, doc cvDoc) {
	// Name — 22pt Bold centered
	cvColor(pdf, cvBlack)
	pdf.SetFont("Times", "B", 22)
	pdf.CellFormat(cvContentW, 8, sanitize(orFallback(doc.FullName, "Your Name")), "", 1, "C", false, 0, "")

	// Contact line 1: Location | Phone | Email (gray)
	line1 := joinNonEmpty(" | ", doc.Location, doc.Phone, doc.Email)
	if line1 != "" {
		cvColor(pdf, cvGray)
		pdf.SetFont("Times", "", 9.5)
		pdf.CellFormat(cvContentW, 4.5, sanitize(line1), "", 1, "C", false, 0, "")
	}

	// Contact line 2: LinkedIn | GitHub | WhatsApp (blue — labels only)
	var links []string
	if doc.LinkedIn != "" {
		links = append(links, "LinkedIn")
	}
	if doc.GitHub != "" {
		links = append(links, "GitHub")
	}
	if doc.WhatsApp != "" {
		links = append(links, "WhatsApp")
	}
	if len(links) > 0 {
		cvColor(pdf, cvBlue)
		pdf.SetFont("Times", "", 9.5)
		pdf.CellFormat(cvContentW, 4.5, strings.Join(links, " | "), "", 1, "C", false, 0, "")
	}

	pdf.Ln(3) // 3mm gap after header
}

func renderCVSummary(pdf *gofpdf.Fpdf, summary string) {
	if strings.TrimSpace(summary) == "" {
		return
	}
	cvPageCheck(pdf, 18)
	cvHeading(pdf, "PROFESSIONAL SUMMARY")
	cvColor(pdf, cvBlack)
	pdf.SetFont("Times", "", 10)
	pdf.MultiCell(cvContentW, 5.1, sanitize(summary), "", "J", false)
	pdf.Ln(4)
}

func renderCVExperience(pdf *gofpdf.Fpdf, entries []cvDocExp) {
	if len(entries) == 0 {
		return
	}
	cvPageCheck(pdf, 20)
	cvHeading(pdf, "WORK EXPERIENCE")
	for _, e := range entries {
		renderCVExpEntry(pdf, e)
	}
}

func renderCVExpEntry(pdf *gofpdf.Fpdf, e cvDocExp) {
	cvPageCheck(pdf, 18)

	dates := cvDateRange(e.StartDate, e.EndDate, e.IsCurrent)

	// Row 1: Role (bold 10.5pt, black, left) | Date (bold 10pt, black, right-aligned)
	y := pdf.GetY()
	cvColor(pdf, cvBlack)
	pdf.SetFont("Times", "B", 10.5)
	pdf.SetXY(cvMargin, y)
	pdf.CellFormat(cvCol1W, 5.5, sanitize(e.Role), "", 0, "L", false, 0, "")
	pdf.SetFont("Times", "B", 10)
	pdf.CellFormat(cvCol2W, 5.5, sanitize(dates), "", 1, "R", false, 0, "")

	// Row 2: Company | website (italic 10pt, gray, left) | Location (italic, right)
	compLine := e.Company
	if e.CompanyWebsite != "" {
		compLine += " | " + e.CompanyWebsite
	}
	y = pdf.GetY()
	cvColor(pdf, cvGray)
	pdf.SetFont("Times", "I", 10)
	pdf.SetXY(cvMargin, y)
	pdf.CellFormat(cvCol1W, 4.5, sanitize(compLine), "", 0, "L", false, 0, "")
	pdf.CellFormat(cvCol2W, 4.5, sanitize(e.Location), "", 1, "R", false, 0, "")

	pdf.Ln(1) // 1mm gap before bullets

	// Bullets: achievements only; fall back to description if no achievements
	cvColor(pdf, cvBlack)
	pdf.SetFont("Times", "", 10)
	bullets := e.Achievements
	if len(bullets) == 0 && strings.TrimSpace(e.Description) != "" {
		bullets = []string{e.Description}
	}
	for _, ach := range bullets {
		if t := strings.TrimSpace(ach); t != "" {
			cvBullet(pdf, t, 5)
		}
	}

	pdf.Ln(3) // 3mm gap after entry
}

func renderCVProjects(pdf *gofpdf.Fpdf, entries []cvDocProj) {
	if len(entries) == 0 {
		return
	}
	cvPageCheck(pdf, 18)
	cvHeading(pdf, "PROJECTS")
	for _, proj := range entries {
		renderCVProjEntry(pdf, proj)
	}
}

func renderCVProjEntry(pdf *gofpdf.Fpdf, proj cvDocProj) {
	cvPageCheck(pdf, 14)

	// Name — Store (bold 10.5pt, left) | Dates (bold 10pt, right)
	nameLabel := sanitize(proj.Name)
	if proj.Store != "" {
		nameLabel += " \x97 " + sanitize(proj.Store) // em-dash
	}
	dates := cvDateRange(proj.StartDate, proj.EndDate, false)

	y := pdf.GetY()
	cvColor(pdf, cvBlack)
	pdf.SetFont("Times", "B", 10.5)
	pdf.SetXY(cvMargin, y)
	pdf.CellFormat(cvCol1W, 5.5, nameLabel, "", 0, "L", false, 0, "")
	pdf.SetFont("Times", "B", 10)
	pdf.CellFormat(cvCol2W, 5.5, sanitize(dates), "", 1, "R", false, 0, "")

	// Bullet features
	cvColor(pdf, cvBlack)
	pdf.SetFont("Times", "", 10)
	for _, feat := range proj.Features {
		if t := strings.TrimSpace(feat); t != "" {
			cvBullet(pdf, t, 5)
		}
	}

	pdf.Ln(3)
}

func renderCVSkills(pdf *gofpdf.Fpdf, groups []cvDocSkill) {
	if len(groups) == 0 {
		return
	}
	cvPageCheck(pdf, 18)
	cvHeading(pdf, "TECHNICAL SKILLS")

	cvColor(pdf, cvBlack)
	for _, g := range groups {
		if len(g.Names) == 0 {
			continue
		}
		y := pdf.GetY()

		// Label column (bold, 44mm)
		pdf.SetFont("Times", "B", 10)
		pdf.SetXY(cvMargin, y)
		pdf.CellFormat(cvSkillLabelW, 4.5, sanitize(cvCatLabel(g.Category)+":"), "", 0, "L", false, 0, "")

		// Values column (regular, wraps)
		pdf.SetFont("Times", "", 10)
		pdf.SetXY(cvMargin+cvSkillLabelW, y)
		pdf.MultiCell(cvSkillValueW, 4.5, sanitize(strings.Join(g.Names, ", ")), "", "L", false)
	}

	pdf.Ln(4)
}

func renderCVEducation(pdf *gofpdf.Fpdf, entries []cvDocEdu) {
	if len(entries) == 0 {
		return
	}
	cvPageCheck(pdf, 18)
	cvHeading(pdf, "EDUCATION")

	for _, edu := range entries {
		if edu.Degree == "" {
			continue
		}
		cvPageCheck(pdf, 14)
		dates := cvDateRange(edu.StartDate, edu.EndDate, false)

		// Degree | Dates
		y := pdf.GetY()
		cvColor(pdf, cvBlack)
		pdf.SetFont("Times", "B", 10.5)
		pdf.SetXY(cvMargin, y)
		pdf.CellFormat(cvCol1W, 5.5, sanitize(edu.Degree), "", 0, "L", false, 0, "")
		pdf.SetFont("Times", "B", 10)
		pdf.CellFormat(cvCol2W, 5.5, sanitize(dates), "", 1, "R", false, 0, "")

		// Institution | Location
		y = pdf.GetY()
		cvColor(pdf, cvGray)
		pdf.SetFont("Times", "I", 10)
		pdf.SetXY(cvMargin, y)
		pdf.CellFormat(cvCol1W, 4.5, sanitize(edu.Institution), "", 0, "L", false, 0, "")
		pdf.CellFormat(cvCol2W, 4.5, sanitize(edu.Location), "", 1, "R", false, 0, "")

		// GPA as bullet
		if edu.GPA != "" {
			cvColor(pdf, cvBlack)
			pdf.SetFont("Times", "", 10)
			cvBullet(pdf, "CGPA: "+edu.GPA, 5)
		}

		pdf.Ln(3)
	}
}

func renderCVLanguages(pdf *gofpdf.Fpdf, entries []cvDocLang) {
	if len(entries) == 0 {
		return
	}
	cvPageCheck(pdf, 15)
	cvHeading(pdf, "LANGUAGES")

	for _, lang := range entries {
		if lang.Language == "" {
			continue
		}
		y := pdf.GetY()
		cvColor(pdf, cvBlack)
		pdf.SetFont("Times", "B", 10)
		pdf.SetXY(cvMargin, y)
		pdf.CellFormat(30, 5.1, sanitize(lang.Language+":"), "", 0, "L", false, 0, "")
		pdf.SetFont("Times", "", 10)
		pdf.CellFormat(cvContentW-30, 5.1, sanitize(lang.Level), "", 1, "L", false, 0, "")
	}

	pdf.Ln(4)
}

func renderCVReferences(pdf *gofpdf.Fpdf, entries []cvDocRef) {
	if len(entries) == 0 {
		return
	}
	cvPageCheck(pdf, 20)
	cvHeading(pdf, "REFERENCES")

	for i := 0; i < len(entries); i += 2 {
		left := entries[i]
		var right *cvDocRef
		if i+1 < len(entries) {
			right = &entries[i+1]
		}
		renderCVRefPair(pdf, left, right)
		pdf.Ln(3)
	}
}

func renderCVRefPair(pdf *gofpdf.Fpdf, left cvDocRef, right *cvDocRef) {
	col1X := cvMargin
	col2X := cvMargin + cvRefColW

	type refRow struct{ l, r string; h float64 }
	rows := []refRow{}

	// Name row (bold 10.5pt, 5.5mm)
	rName := ""
	if right != nil { rName = right.Name }
	rows = append(rows, refRow{left.Name, rName, 5.5})

	// Title row (regular 10pt, 4.5mm each)
	rTitle := ""
	if right != nil { rTitle = right.Title }
	rows = append(rows, refRow{left.Title, rTitle, 4.5})

	// Department
	rDept := ""
	if right != nil { rDept = right.Department }
	rows = append(rows, refRow{left.Department, rDept, 4.5})

	// Institution
	rInst := ""
	if right != nil { rInst = right.Institution }
	rows = append(rows, refRow{left.Institution, rInst, 4.5})

	// Email | Phone
	lContact := joinNonEmpty(" | ", left.Email, left.Phone)
	rContact := ""
	if right != nil { rContact = joinNonEmpty(" | ", right.Email, right.Phone) }
	rows = append(rows, refRow{lContact, rContact, 4.5})

	startY := pdf.GetY()
	y := startY

	for idx, row := range rows {
		if row.l == "" && row.r == "" {
			continue
		}
		cvColor(pdf, cvBlack)
		if idx == 0 {
			pdf.SetFont("Times", "B", 10.5)
		} else {
			pdf.SetFont("Times", "", 10)
		}
		pdf.SetXY(col1X, y)
		pdf.CellFormat(cvRefColW, row.h, sanitize(row.l), "", 0, "L", false, 0, "")
		if right != nil {
			pdf.SetXY(col2X, y)
			pdf.CellFormat(cvRefColW, row.h, sanitize(row.r), "", 0, "L", false, 0, "")
		}
		y += row.h
	}

	pdf.SetXY(cvMargin, y)
}

// ─── Layout helpers ───────────────────────────────────────────────

// cvHeading renders the section heading: ALL CAPS, Times Bold 11pt, navy,
// followed by a 0.4pt navy horizontal rule, then a 3mm gap.
func cvHeading(pdf *gofpdf.Fpdf, title string) {
	cvColor(pdf, cvNavy)
	pdf.SetFont("Times", "B", 11)
	pdf.CellFormat(cvContentW, 5.5, title, "", 1, "L", false, 0, "")

	// 0.4pt navy rule
	y := pdf.GetY()
	r, g, b := hexRGB(cvNavy)
	pdf.SetDrawColor(r, g, b)
	pdf.SetLineWidth(0.4)
	pdf.Line(cvMargin, y, cvMargin+cvContentW, y)
	pdf.SetDrawColor(0, 0, 0)
	pdf.Ln(3)
}

// cvBullet renders a bullet point with the given indent (mm from left margin).
// Uses cp1252 bullet byte \x95. Text is justified.
func cvBullet(pdf *gofpdf.Fpdf, text string, indent float64) {
	x := cvMargin + indent
	y := pdf.GetY()
	pdf.SetXY(x, y)
	pdf.CellFormat(3, 5, "\x95", "", 0, "L", false, 0, "")
	pdf.SetXY(x+4, y)
	pdf.MultiCell(cvContentW-indent-4, 5, sanitize(text), "", "J", false)
}

// cvPageCheck ensures at least minSpace mm is available before rendering;
// adds a new page if not, preventing orphaned headings and entries.
func cvPageCheck(pdf *gofpdf.Fpdf, minSpace float64) {
	_, pageH := pdf.GetPageSize()
	available := pageH - cvMargin - pdf.GetY()
	if available < minSpace {
		pdf.AddPage()
	}
}

// cvDateRange returns "Start \x96 End" (en-dash in cp1252) or "Start \x96 Present".
func cvDateRange(start, end string, isCurrent bool) string {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if isCurrent || strings.EqualFold(end, "present") || end == "" {
		end = "Present"
	}
	if start == "" {
		return end
	}
	return start + " \x96 " + end
}

// cvColor sets text color from a 6-char hex string.
func cvColor(pdf *gofpdf.Fpdf, hex string) {
	r, g, b := hexRGB(hex)
	pdf.SetTextColor(r, g, b)
}

// cvCatLabel maps internal category keys to display labels.
func cvCatLabel(cat string) string {
	switch cat {
	case "languages":
		return "Languages"
	case "framework":
		return "Framework"
	case "core":
		return "Core"
	case "state_management":
		return "State Management"
	case "backend":
		return "Backend & APIs"
	case "payments":
		return "Payments"
	case "tools":
		return "Tools & Platforms"
	case "architecture":
		return "Architecture"
	case "distribution":
		return "App Distribution"
	default:
		return titleCase(strings.ReplaceAll(cat, "_", " "))
	}
}

// ─── General helpers ──────────────────────────────────────────────

func hexRGB(hex string) (int, int, int) {
	if len(hex) != 6 {
		return 0, 0, 0
	}
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}

func joinNonEmpty(sep string, parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return strings.Join(out, sep)
}

func orFallback(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

// cp1252Map maps select Unicode code points to Windows-1252 byte equivalents
// so gofpdf's built-in font renders them correctly.
var cp1252Map = map[rune]byte{
	'–': 0x96, // en dash
	'—': 0x97, // em dash
	'‘': 0x91, // left single quote
	'’': 0x92, // right single quote
	'“': 0x93, // left double quote
	'”': 0x94, // right double quote
	'•': 0x95, // bullet
	'…': 0x85, // ellipsis
	'€': 0x80, // euro sign
	'™': 0x99, // trade mark
}

func sanitize(s string) string {
	var out strings.Builder
	out.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\r':
			// skip
		case r < 32 && r != '\n' && r != '\t':
			// skip control chars
		case r <= 0x7E:
			out.WriteByte(byte(r))
		case r >= 0xA0 && r <= 0xFF:
			out.WriteByte(byte(r))
		default:
			if b, ok := cp1252Map[r]; ok {
				out.WriteByte(b)
			}
			// skip unmappable chars
		}
	}
	return out.String()
}
