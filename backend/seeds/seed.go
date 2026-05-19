// ============================================================
// FILE: seeds/seed.go
// WHAT IT IS:     Database seeding — creates default data on first run
// WHY IT EXISTS:  When the server first starts with a fresh database,
//                 there is no admin user to log in with and no content
//                 to display. This seed creates the default admin account
//                 and pre-fills all portfolio sections with Imtiaz's data
//                 so the portfolio is immediately usable.
// DEPENDS ON:     models, utils/hash.go, config, gorm
// IF REMOVED:     Fresh installs start with empty database — admin login fails
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package seeds

import (
	"fmt"
	"imtiaz-portfolio/config"
	"imtiaz-portfolio/internal/models"
	"imtiaz-portfolio/internal/utils"
	"log"

	"gorm.io/gorm"
)

/**
 * FUNCTION: Run
 * WHAT IT DOES:   Checks if seed data already exists and creates it
 *                 if not. Uses "upsert" logic — safe to call on every
 *                 server startup without creating duplicate data.
 *                 Seeds: admin user, profile, skills, experience, projects.
 * WHERE CALLED:   cmd/main.go → after DB migration on startup
 * PARAMETERS:     @param {*gorm.DB} db - database connection
 *                 @param {*config.Config} cfg - holds admin credentials
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func Run(db *gorm.DB, cfg *config.Config) {
	fmt.Println("🌱 Running database seeds...")

	seedAdmin(db, cfg)
	seedProfile(db)
	seedSkills(db)
	seedExperience(db)
	seedProjects(db)
	seedServices(db)
	seedArchitecture(db)

	fmt.Println("✅ Database seeding complete")
}

// seedAdmin creates the default admin account if it doesn't exist.
// Credentials come from the .env file so they can be changed before deploy.
func seedAdmin(db *gorm.DB, cfg *config.Config) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		return // Admin already exists — skip to avoid overwriting password changes
	}

	// Hash the password before storing — NEVER store plain text passwords
	hash, err := utils.HashPassword(cfg.AdminPassword)
	if err != nil {
		log.Fatal("Failed to hash admin password during seeding: ", err)
	}

	admin := models.User{
		Name:     "Chowdhury Md. Imtiazul Islam",
		Email:    cfg.AdminEmail,
		Password: hash,
		Role:     "admin",
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("⚠️  Failed to seed admin user: %v", err)
		return
	}
	fmt.Printf("   ✓ Admin created: %s\n", cfg.AdminEmail)
}

// seedProfile creates the initial portfolio profile with Imtiaz's information.
func seedProfile(db *gorm.DB) {
	var count int64
	db.Model(&models.Profile{}).Count(&count)
	if count > 0 {
		return
	}

	profile := models.Profile{
		FullName:        "Chowdhury Md. Imtiazul Islam",
		Title:           "Flutter Developer",
		Tagline:         "Flutter Developer building apps that live on App Store & Play Store",
		Bio:             "Flutter Developer with 2.5+ years of experience building and shipping production-ready cross-platform mobile apps for iOS and Android. Delivered 5+ live apps on the App Store and Google Play Store. Skilled in Clean Architecture, Riverpod/GetX/Bloc, REST APIs, Firebase, Socket.IO, WebRTC, and payment gateways (Stripe, RevenueCat, Amar Pay). Available immediately for remote or onsite roles.",
		ShortBio:        "Flutter Developer with 2.5+ years shipping live apps to App Store & Play Store.",
		Email:           "work.imtiazchowdhury@gmail.com",
		Phone:           "",
		WhatsApp:        "",
		Location:        "Dhaka, Bangladesh",
		GitHub:          "https://github.com/imtiazchowdhury",
		LinkedIn:        "https://linkedin.com/in/imtiazchowdhury",
		Twitter:         "https://twitter.com/imtiazchowdhury",
		Instagram:       "https://instagram.com/imtiazchowdhury",
		Availability:    "Open to Work",
		YearsExperience: "2.5+",
		AppsShipped:     "5+",
		MetaTitle:       "Chowdhury Md. Imtiazul Islam — Flutter Developer",
		MetaDescription: "Flutter Developer with 2.5+ years experience. 5+ apps live on App Store & Play Store. Available for remote or onsite roles.",
	}

	if err := db.Create(&profile).Error; err != nil {
		log.Printf("⚠️  Failed to seed profile: %v", err)
		return
	}
	fmt.Println("   ✓ Profile seeded")
}

// seedSkills creates all tech stack entries with percentages.
func seedSkills(db *gorm.DB) {
	var count int64
	db.Model(&models.Skill{}).Count(&count)
	if count > 0 {
		return
	}

	skills := []models.Skill{
		// ─── Core ────────────────────────────────────────────
		{Name: "Flutter", Category: "core", Percentage: 95, Description: "Cross-platform UI framework by Google for building beautiful native apps", SortOrder: 1},
		{Name: "Dart", Category: "core", Percentage: 90, Description: "The programming language used by Flutter — fast, strongly typed", SortOrder: 2},

		// ─── State Management ─────────────────────────────────
		{Name: "GetX", Category: "state_management", Percentage: 85, Description: "Lightweight state management, routing, and dependency injection for Flutter", SortOrder: 1},
		{Name: "BLoC", Category: "state_management", Percentage: 85, Description: "Business Logic Component — separates UI from business logic using streams", SortOrder: 2},
		{Name: "Riverpod", Category: "state_management", Percentage: 80, Description: "Modern state management — testable, composable, and type-safe", SortOrder: 3},

		// ─── Backend & APIs ───────────────────────────────────
		{Name: "Firebase", Category: "backend", Percentage: 88, Description: "Google's backend platform — Auth, Firestore, Storage, Cloud Messaging", SortOrder: 1},
		{Name: "REST API", Category: "backend", Percentage: 90, Description: "HTTP API integration — JSON parsing, authentication headers, error handling", SortOrder: 2},
		{Name: "Socket.IO", Category: "backend", Percentage: 80, Description: "Real-time bidirectional communication for chat and live features", SortOrder: 3},
		{Name: "WebRTC", Category: "backend", Percentage: 75, Description: "Peer-to-peer video/audio calling directly in Flutter apps", SortOrder: 4},

		// ─── Payments ─────────────────────────────────────────
		{Name: "Stripe", Category: "payments", Percentage: 80, Description: "Payment processing — card payments, subscriptions, webhooks", SortOrder: 1},
		{Name: "RevenueCat", Category: "payments", Percentage: 85, Description: "In-app purchases and subscriptions for iOS and Android", SortOrder: 2},
		{Name: "Amar Pay", Category: "payments", Percentage: 75, Description: "Bangladesh payment gateway integration for local apps", SortOrder: 3},

		// ─── Tools & DevOps ───────────────────────────────────
		{Name: "Git", Category: "tools", Percentage: 90, Description: "Version control — branching, merging, pull requests", SortOrder: 1},
		{Name: "GitHub", Category: "tools", Percentage: 88, Description: "Remote repository hosting, CI/CD with GitHub Actions", SortOrder: 2},
		{Name: "Figma", Category: "tools", Percentage: 75, Description: "Design tool — reading specs, extracting assets, pixel-perfect implementation", SortOrder: 3},
		{Name: "VS Code", Category: "tools", Percentage: 92, Description: "Primary IDE for Flutter development with extensions", SortOrder: 4},
		{Name: "Postman", Category: "tools", Percentage: 85, Description: "API testing and documentation before Flutter integration", SortOrder: 5},
	}

	if err := db.Create(&skills).Error; err != nil {
		log.Printf("⚠️  Failed to seed skills: %v", err)
		return
	}
	fmt.Printf("   ✓ %d skills seeded\n", len(skills))
}

// seedExperience creates the work history timeline entries.
func seedExperience(db *gorm.DB) {
	var count int64
	db.Model(&models.Experience{}).Count(&count)
	if count > 0 {
		return
	}

	experiences := []models.Experience{
		{
			Company:     "SperkTech",
			Role:        "Flutter Developer",
			StartDate:   "Jan 2024",
			EndDate:     "",
			IsCurrent:   true,
			Description: "Building and maintaining production Flutter applications for iOS and Android. Working with real-time features, payment integrations, and clean architecture patterns.",
			Achievements: models.StringArray{
				"Shipped 2 apps to App Store and Play Store from scratch",
				"Integrated Stripe payment gateway and RevenueCat subscriptions",
				"Implemented real-time chat using Socket.IO",
				"Reduced app crash rate by 40% through systematic error handling",
			},
			TechUsed:  models.StringArray{"Flutter", "Dart", "Firebase", "GetX", "Stripe", "Socket.IO", "REST API"},
			Location:  "Dhaka, Bangladesh",
			Type:      "full-time",
			SortOrder: 1,
		},
		{
			Company:     "SoftVence",
			Role:        "Flutter Developer",
			StartDate:   "Jun 2023",
			EndDate:     "Dec 2023",
			IsCurrent:   false,
			Description: "Developed cross-platform mobile applications with focus on UI/UX quality and performance optimization.",
			Achievements: models.StringArray{
				"Built 2 client-facing apps with 4.5+ App Store rating",
				"Implemented BLoC pattern for scalable state management",
				"Integrated WebRTC for video calling feature",
				"Delivered projects on time with zero post-release critical bugs",
			},
			TechUsed:  models.StringArray{"Flutter", "Dart", "BLoC", "Firebase", "WebRTC", "REST API"},
			Location:  "Dhaka, Bangladesh",
			Type:      "full-time",
			SortOrder: 2,
		},
		{
			Company:     "Daffodil International University",
			Role:        "Flutter Development Intern",
			StartDate:   "Jan 2023",
			EndDate:     "May 2023",
			IsCurrent:   false,
			Description: "Internship focused on Flutter mobile app development, learning industry practices and contributing to university projects.",
			Achievements: models.StringArray{
				"Built a university event management mobile app",
				"Learned clean architecture and best practices under mentorship",
				"Contributed to open-source Flutter packages",
			},
			TechUsed:  models.StringArray{"Flutter", "Dart", "Firebase", "GetX"},
			Location:  "Dhaka, Bangladesh",
			Type:      "internship",
			SortOrder: 3,
		},
	}

	if err := db.Create(&experiences).Error; err != nil {
		log.Printf("⚠️  Failed to seed experience: %v", err)
		return
	}
	fmt.Printf("   ✓ %d experience entries seeded\n", len(experiences))
}

// seedProjects creates the showcase app entries.
func seedProjects(db *gorm.DB) {
	var count int64
	db.Model(&models.Project{}).Count(&count)
	if count > 0 {
		return
	}

	projects := []models.Project{
		{
			Name:            "Project Finder",
			Slug:            "project-finder",
			Description:     "A platform connecting students and professionals with collaborative projects. Browse, join, or post projects across different domains.",
			LongDescription: "Project Finder is a cross-platform mobile app that bridges the gap between project creators and talented contributors. Whether you're a student looking for a real-world project or a developer seeking collaborators, Project Finder makes discovery effortless. Features real-time notifications, in-app chat, and a smart matching algorithm.",
			TechStack:       models.StringArray{"Flutter", "Firebase", "GetX", "REST API", "Socket.IO"},
			Features: models.StringArray{
				"Smart project matching based on skills and interests",
				"Real-time chat between project members",
				"Push notifications for new project opportunities",
				"Profile portfolio with skills showcase",
			},
			Screenshots:  models.StringArray{},
			AppStoreURL:  "https://apps.apple.com",
			PlayStoreURL: "https://play.google.com",
			Status:       "live",
			Featured:     true,
			SortOrder:    1,
		},
		{
			Name:            "Hiye Health App",
			Slug:            "hiye-health-app",
			Description:     "A comprehensive healthcare companion app for tracking health metrics, appointments, and connecting with doctors.",
			LongDescription: "Hiye is a health and wellness platform that puts your health data in your hands. Track vitals, schedule appointments, and get insights into your health trends. Integrated with wearable devices and supports real-time consultations via video calling.",
			TechStack:       models.StringArray{"Flutter", "Firebase", "BLoC", "WebRTC", "RevenueCat", "REST API"},
			Features: models.StringArray{
				"Health metrics tracking and trend visualization",
				"Video consultations with healthcare providers via WebRTC",
				"Appointment scheduling and reminder notifications",
				"Premium subscription management via RevenueCat",
			},
			Screenshots:  models.StringArray{},
			AppStoreURL:  "",
			PlayStoreURL: "https://play.google.com",
			Status:       "live",
			Featured:     true,
			SortOrder:    2,
		},
	}

	if err := db.Create(&projects).Error; err != nil {
		log.Printf("⚠️  Failed to seed projects: %v", err)
		return
	}
	fmt.Printf("   ✓ %d projects seeded\n", len(projects))
}

// seedServices creates the default "What I Offer" service cards.
func seedServices(db *gorm.DB) {
	var count int64
	db.Model(&models.Service{}).Count(&count)
	if count > 0 {
		return
	}

	services := []models.Service{
		{
			Icon:        "📱",
			Title:       "Mobile App Development",
			Description: "I build production-ready Flutter apps for iOS and Android from scratch to App Store launch. Clean code, clean architecture, real results.",
			SortOrder:   1,
			Active:      true,
		},
		{
			Icon:        "🔧",
			Title:       "App Maintenance & Updates",
			Description: "I maintain, debug, and improve existing Flutter apps. Performance optimization, dependency upgrades, new feature integration.",
			SortOrder:   2,
			Active:      true,
		},
		{
			Icon:        "🔌",
			Title:       "API & Service Integration",
			Description: "I integrate REST APIs, Firebase, payment gateways (Stripe, RevenueCat), Socket.IO, WebRTC, and third-party SDKs into Flutter apps.",
			SortOrder:   3,
			Active:      true,
		},
	}

	if err := db.Create(&services).Error; err != nil {
		log.Printf("⚠️  Failed to seed services: %v", err)
		return
	}
	fmt.Printf("   ✓ %d services seeded\n", len(services))
}

func seedArchitecture(db *gorm.DB) {
	var count int64
	db.Model(&models.Architecture{}).Count(&count)
	if count > 0 {
		return
	}

	items := []models.Architecture{
		{
			Number:      "01",
			Name:        "Clean Architecture",
			Diagram:     "UI Layer\n    ↓\nDomain Layer (Use Cases)\n    ↓\nData Layer (Repositories)\n    ↓\n  Database / API",
			Description: "Separates the app into layers with clear boundaries. Business logic in the domain layer never depends on UI or data frameworks — making it independently testable and framework-agnostic.",
			Projects:    models.StringArray{"Project Finder", "Hiye Health"},
			SortOrder:   1,
			Active:      true,
		},
		{
			Number:      "02",
			Name:        "BLoC Pattern",
			Diagram:     "    Events\n      ↓\n[BLoC / Cubit]\n      ↓\n    States\n      ↓\n    UI Widgets",
			Description: "Business Logic Component separates UI from business logic using streams. Events flow in, states flow out. Every state change is explicit, predictable, and testable.",
			Projects:    models.StringArray{"Hiye Health"},
			SortOrder:   2,
			Active:      true,
		},
		{
			Number:      "03",
			Name:        "GetX Pattern",
			Diagram:     "  Controllers\n  (logic + state)\n      ↓\n  GetX Bindings\n  (DI + routing)\n      ↓\n  Obx Widgets\n  (reactive UI)",
			Description: "Lightweight state management with built-in routing and dependency injection. Reactive variables (Rx) automatically update the UI without manual setState or streams.",
			Projects:    models.StringArray{"Project Finder"},
			SortOrder:   3,
			Active:      true,
		},
		{
			Number:      "04",
			Name:        "MVVM",
			Diagram:     "   View (Widget)\n       ↕\n  ViewModel\n  (ChangeNotifier)\n       ↕\n  Model / Repo",
			Description: "Model-View-ViewModel keeps UI code clean by moving all logic into the ViewModel. The View only observes the ViewModel — never makes decisions.",
			Projects:    models.StringArray{"SperkTech Apps"},
			SortOrder:   4,
			Active:      true,
		},
		{
			Number:      "05",
			Name:        "Repository Pattern",
			Diagram:     "UI / BLoC / GetX\n       ↓\n   Repository\n  (interface)\n    ↙      ↘\nRemoteDS  LocalDS\n(API)   (Hive/SQLite)",
			Description: "The Repository acts as the single source of truth, deciding whether to fetch fresh data from the API or serve cached local data — completely transparent to the UI layer.",
			Projects:    models.StringArray{"Project Finder", "Hiye Health"},
			SortOrder:   5,
			Active:      true,
		},
	}

	if err := db.Create(&items).Error; err != nil {
		log.Printf("⚠️  Failed to seed architecture: %v", err)
		return
	}
	fmt.Printf("   ✓ %d architecture items seeded\n", len(items))
}
