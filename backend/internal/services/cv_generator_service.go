// ============================================================
// FILE: internal/services/cv_generator_service.go
// WHAT IT IS:     Builds a clean classic resume PDF from the
//                 profile + skills + experience data already in
//                 the database. The result is saved to
//                 /uploads/cv/ and registered as a new history
//                 row (active by default).
// ENDPOINTS:      POST /api/v1/profile/cv/generate
// LAYOUT:         White background, black text, traditional
//                 resume formatting (recruiter-friendly).
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

type CVGeneratorService struct {
	db        *gorm.DB
	profile   *ProfileService
	uploadDir string
}

func NewCVGeneratorService(db *gorm.DB, profile *ProfileService, uploadDir string) *CVGeneratorService {
	return &CVGeneratorService{db: db, profile: profile, uploadDir: uploadDir}
}

// Generate builds a PDF resume from current DB content, saves it to disk, and
// appends a row to cv_files with source="generated" (which also activates it).
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

	pdf := buildResumePDF(profile, skills, experiences)

	// Ensure /uploads/cv exists
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
	displayName := fmt.Sprintf("%s — Resume.pdf", strings.TrimSpace(profile.FullName))
	if profile.FullName == "" {
		displayName = fileName
	}

	// Don't auto-activate — admin previews the generated CV in the dashboard
	// and explicitly clicks "Yes, set as active" to make it live.
	return s.profile.AddCVHistory(publicPath, displayName, size, "generated", false)
}

// ─── PDF layout ──────────────────────────────────────────────

const (
	pageMargin = 18.0 // mm
	pageWidth  = 210.0
	contentW   = pageWidth - 2*pageMargin

	colorHeading = "000000"
	colorAccent  = "1f1f1f"
	colorMuted   = "555555"
	colorRule    = "BBBBBB"
)

func buildResumePDF(p *models.Profile, skills []models.Skill, experiences []models.Experience) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pageMargin, pageMargin, pageMargin)
	pdf.SetAutoPageBreak(true, pageMargin)
	pdf.AddPage()

	// ─── Header: Name + Title ──────────────────────────────
	setRGB(pdf, colorHeading)
	pdf.SetFont("Helvetica", "B", 22)
	pdf.CellFormat(contentW, 9, sanitize(orFallback(p.FullName, "Your Name")), "", 1, "L", false, 0, "")

	if t := strings.TrimSpace(p.Title); t != "" {
		setRGB(pdf, colorAccent)
		pdf.SetFont("Helvetica", "", 12)
		pdf.CellFormat(contentW, 6, sanitize(t), "", 1, "L", false, 0, "")
	}

	// Contact bar — pipes between non-empty values
	contact := joinNonEmpty(" | ",
		p.Email, p.Phone, p.Location,
		stripScheme(p.LinkedIn), stripScheme(p.GitHub),
	)
	if contact != "" {
		setRGB(pdf, colorMuted)
		pdf.SetFont("Helvetica", "", 9)
		pdf.MultiCell(contentW, 4.5, sanitize(contact), "", "L", false)
	}
	pdf.Ln(2)
	hr(pdf)
	pdf.Ln(3)

	// ─── Summary ───────────────────────────────────────────
	bio := strings.TrimSpace(p.Bio)
	if bio == "" {
		bio = strings.TrimSpace(p.ShortBio)
	}
	if bio != "" {
		sectionTitle(pdf, "Summary")
		setRGB(pdf, colorHeading)
		pdf.SetFont("Helvetica", "", 10)
		pdf.MultiCell(contentW, 5, sanitize(bio), "", "L", false)
		pdf.Ln(3)
	}

	// ─── Experience ────────────────────────────────────────
	if len(experiences) > 0 {
		sectionTitle(pdf, "Experience")

		// Sort defensively — sort_order ascending, then most recent first
		sort.SliceStable(experiences, func(i, j int) bool {
			if experiences[i].SortOrder != experiences[j].SortOrder {
				return experiences[i].SortOrder < experiences[j].SortOrder
			}
			return experiences[i].ID > experiences[j].ID
		})

		for _, e := range experiences {
			// Role @ Company  (left)        Dates (right)
			setRGB(pdf, colorHeading)
			pdf.SetFont("Helvetica", "B", 11)
			roleLine := strings.TrimSpace(e.Role)
			if e.Company != "" {
				roleLine = fmt.Sprintf("%s — %s", roleLine, e.Company)
			}
			dates := dateRange(e.StartDate, e.EndDate, e.IsCurrent)
			pdf.CellFormat(contentW*0.7, 6, sanitize(roleLine), "", 0, "L", false, 0, "")
			setRGB(pdf, colorMuted)
			pdf.SetFont("Helvetica", "I", 9)
			pdf.CellFormat(contentW*0.3, 6, sanitize(dates), "", 1, "R", false, 0, "")

			if e.Location != "" {
				setRGB(pdf, colorMuted)
				pdf.SetFont("Helvetica", "", 9)
				pdf.CellFormat(contentW, 5, sanitize(e.Location), "", 1, "L", false, 0, "")
			}

			if d := strings.TrimSpace(e.Description); d != "" {
				setRGB(pdf, colorHeading)
				pdf.SetFont("Helvetica", "", 10)
				pdf.MultiCell(contentW, 5, sanitize(d), "", "L", false)
			}

			// Achievements as bullet list
			for _, ach := range e.Achievements {
				ach = strings.TrimSpace(ach)
				if ach == "" {
					continue
				}
				setRGB(pdf, colorHeading)
				pdf.SetFont("Helvetica", "", 10)
				bullet(pdf, ach)
			}

			// Tech stack
			if len(e.TechUsed) > 0 {
				setRGB(pdf, colorMuted)
				pdf.SetFont("Helvetica", "I", 9)
				pdf.MultiCell(contentW, 4.5, "Tech: "+sanitize(strings.Join(e.TechUsed, ", ")), "", "L", false)
			}
			pdf.Ln(2)
		}
	}

	// ─── Skills (grouped by category) ──────────────────────
	if len(skills) > 0 {
		sectionTitle(pdf, "Skills")
		grouped := groupSkills(skills)
		categoryOrder := []string{"core", "state_management", "backend", "payments", "tools"}
		seen := map[string]bool{}
		for _, cat := range categoryOrder {
			if list, ok := grouped[cat]; ok {
				renderSkillRow(pdf, prettyCategory(cat), list)
				seen[cat] = true
			}
		}
		// Render any categories not in the predefined order, alphabetically
		var extras []string
		for k := range grouped {
			if !seen[k] {
				extras = append(extras, k)
			}
		}
		sort.Strings(extras)
		for _, k := range extras {
			renderSkillRow(pdf, prettyCategory(k), grouped[k])
		}
	}

	// ─── Footer ────────────────────────────────────────────
	pdf.SetY(-15)
	setRGB(pdf, colorMuted)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.CellFormat(0, 5,
		fmt.Sprintf("Generated from portfolio data on %s", time.Now().Format("Jan 2, 2006")),
		"", 0, "C", false, 0, "")

	return pdf
}

// ─── PDF helpers ─────────────────────────────────────────────

func sectionTitle(pdf *gofpdf.Fpdf, title string) {
	setRGB(pdf, colorAccent)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(contentW, 6, strings.ToUpper(title), "", 1, "L", false, 0, "")
	hr(pdf)
	pdf.Ln(2)
}

func bullet(pdf *gofpdf.Fpdf, text string) {
	x := pdf.GetX()
	y := pdf.GetY()
	pdf.SetX(x + 4)
	pdf.CellFormat(3, 5, "•", "", 0, "L", false, 0, "")
	pdf.SetX(x + 7)
	pdf.MultiCell(contentW-7, 5, sanitize(text), "", "L", false)
	_ = y
}

func renderSkillRow(pdf *gofpdf.Fpdf, label string, list []models.Skill) {
	if len(list) == 0 {
		return
	}
	names := make([]string, 0, len(list))
	for _, sk := range list {
		names = append(names, sk.Name)
	}
	setRGB(pdf, colorHeading)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(contentW, 5, sanitize(label), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.MultiCell(contentW, 5, sanitize(strings.Join(names, " • ")), "", "L", false)
	pdf.Ln(1)
}

func hr(pdf *gofpdf.Fpdf) {
	x := pdf.GetX()
	y := pdf.GetY()
	r, g, b := hexToRGB(colorRule)
	pdf.SetDrawColor(r, g, b)
	pdf.Line(x, y, x+contentW, y)
}

func setRGB(pdf *gofpdf.Fpdf, hex string) {
	r, g, b := hexToRGB(hex)
	pdf.SetTextColor(r, g, b)
}

func hexToRGB(hex string) (int, int, int) {
	if len(hex) != 6 {
		return 0, 0, 0
	}
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

func groupSkills(skills []models.Skill) map[string][]models.Skill {
	out := map[string][]models.Skill{}
	for _, sk := range skills {
		out[sk.Category] = append(out[sk.Category], sk)
	}
	return out
}

func prettyCategory(cat string) string {
	switch cat {
	case "core":
		return "Core"
	case "state_management":
		return "State Management"
	case "backend":
		return "Backend"
	case "payments":
		return "Payments"
	case "tools":
		return "Tools"
	default:
		return strings.Title(strings.ReplaceAll(cat, "_", " "))
	}
}

func dateRange(start, end string, isCurrent bool) string {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if isCurrent || end == "" {
		end = "Present"
	}
	if start == "" {
		return end
	}
	return start + " — " + end
}

func joinNonEmpty(sep string, parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, sep)
}

func stripScheme(url string) string {
	u := strings.TrimSpace(url)
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	return u
}

func orFallback(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

// sanitize strips characters that the default cp1252 PDF font can't render
// (most emoji + many extended Unicode glyphs). Falls back to '?' for unknown
// runes so the layout stays intact.
func sanitize(s string) string {
	if s == "" {
		return ""
	}
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r == '\r':
			continue
		case r < 32 && r != '\n' && r != '\t':
			continue
		case r > 0xFF:
			out = append(out, '?')
		default:
			out = append(out, r)
		}
	}
	return string(out)
}
