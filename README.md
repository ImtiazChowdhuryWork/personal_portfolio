# Chowdhury Md. Imtiazul Islam — Flutter Developer Portfolio

A full-stack personal portfolio website inspired by the **Drake theme** (wpriverthemes.com/HTML/drake), rebuilt from scratch for a Flutter Developer. Includes a public-facing portfolio and a private admin dashboard for managing all content dynamically.

---

## Table of Contents

0. [**Where We Left Off**](#0-where-we-left-off--2026-05-10) — read first when resuming work
1. [Project Overview](#1-project-overview)
2. [Full Tech Stack](#2-full-tech-stack)
3. [Architecture Overview](#3-architecture-overview)
4. [Complete Folder Structure](#4-complete-folder-structure)
5. [Getting Started](#5-getting-started)
6. [Credentials](#6-credentials)
7. [How to Run](#7-how-to-run)
8. [How Everything Works (Data Flow)](#8-how-everything-works-data-flow)
9. [All API Endpoints](#9-all-api-endpoints)
10. [Database Schema](#10-database-schema)
11. [Frontend Pages](#11-frontend-pages)
12. [Component Library](#12-component-library)
13. [Design System](#13-design-system)
14. [What Has Been Done So Far](#14-what-has-been-done-so-far)
15. [Known Issues & Pending Fixes](#15-known-issues--pending-fixes)
16. [Roadmap](#16-roadmap)
17. [Deployment Guide](#17-deployment-guide)
18. [Changelog](#18-changelog)

---

## 0. Where We Left Off · 2026-05-10

> Live snapshot at the end of the most recent session — read this first when resuming work.

**Current version:** `v3.45.3` · DB schema is at 10 tables (no pending migrations).

### Last two sessions (2026-05-09 → 2026-05-10) focused entirely on the dashboard's content-editing UX

**CV / Resume system** — built end-to-end:
- New `cv_files` table + version history with active picker, hide/show toggle, download counter, in-dashboard PDF preview (PDF.js-rendered canvases), Compare-two-CVs modal with side-by-side panes + chip strip + "Make Active" button per pane.
- Auto-generator (`gofpdf`-based) builds a basic resume from profile/skills/experience. **Needs significant polish — see § 15 → CV / Resume.**

**Profile tab redesigned** — split layout, sticky identity card with two photo uploads, three status cards (Open to Work / Available for Freelance / 🌴 On Vacation), Phosphor-iconed form sections, live identity preview, Bio character counter, phone field with `intl-tel-input` country code picker.

**Footer & Copyright** — split into its own `📜 Footer` sidebar tab with live preview, 5 quick-fill presets, insert chips, `{year}`/`{name}` placeholders, optional "Built with…" line, collapsible per-place overrides for sidebar vs. footer.

**Hero customization on the portfolio** — every previously hardcoded text spot is now editable from the dashboard:
- `nickname` + `hero_subtitle` (replace hardcoded "Imtiaz" + greeting pill)
- `hero_heading_line1/2/3` + `hero_heading_highlights` + advanced override (replace the big 3-line hero title)
- `title` (sidebar designation, hero subtitle, footer tagline)
- `copyright_text` + sidebar/footer overrides + built-with line
- `availability` now drives **per-status visibility/wording** of all hire-related buttons + a status banner above the contact form, and tailors the contact-form success message ("On Vacation" no longer falsely promises 24-hour reply).

### Suggested next areas (in priority order)

1. **Dashboard CRUD modals for Projects, Skills, Experience** — this is the biggest gap remaining (§15 → Dashboard). Today the Add buttons just toast a placeholder.
2. **CV Generate polish** — pick from the punch list in §15 → CV / Resume (section controls, preview-before-save, design pass, richer content).
3. **Tier 2 status extras** — `vacation_return_date` field + "Back on {date}" banner messaging (we did Tier 1 of the Status overhaul; Tier 2 was deferred).
4. **Testimonials section on the public portfolio** — table exists in DB, no UI.
5. **Project detail modal** — clicking "View Details" on an app slide.

---

## 1. Project Overview

This is a **single-page portfolio website** for Chowdhury Md. Imtiazul Islam, a Flutter Developer based in Dhaka, Bangladesh.

**What it does:**
- Shows portfolio sections: Hero, About, Tech Stack, Architecture, Apps Showcase, Experience, GitHub Stats, Services, Contact
- All content (skills, projects, experience, profile) is stored in a PostgreSQL database and served via a REST API
- The admin can log in to a dashboard and edit everything — no code changes needed
- The Go backend serves both the API and the frontend, so one command starts everything

**Design inspiration:** Drake HTML theme — dark `#1f1f1f` background, floating left sidebar card, right-side icon navigation, light font-weight headings, pill-shaped buttons, GSAP-style scroll animations.

**Live URLs (local development):**

| URL | What it opens |
|-----|---------------|
| `http://localhost:8080` | Portfolio (public) |
| `http://localhost:8080/login` | Admin login |
| `http://localhost:8080/dashboard` | Admin dashboard |
| `http://localhost:8080/api/v1/...` | REST API |
| `http://localhost:8080/health` | Health check |

---

## 2. Full Tech Stack

### Backend
| Technology | Version | Purpose |
|-----------|---------|---------|
| Go (Golang) | 1.21+ | Server language |
| Gin | v1.12 | HTTP web framework |
| GORM | v1.31 | ORM for PostgreSQL |
| PostgreSQL | 17 | Primary database |
| golang-jwt/jwt | v5 | JWT authentication |
| bcrypt | — | Password hashing |
| godotenv | v1.5 | Environment variable loading |

### Frontend
| Technology | Version | Purpose |
|-----------|---------|---------|
| HTML5 | — | Page structure |
| CSS3 | — | Styling (component-based, no framework) |
| Vanilla JavaScript | ES6+ | All interactivity |
| Swiper.js | v11 | Apps Showcase slider |
| Phosphor Icons | v2.1.1 | Icon set (sidebar + right nav) |
| Google Fonts (Inter) | 300–900 | Typography |
| IntersectionObserver API | native | Scroll animations |

### Infrastructure
| Tool | Purpose |
|------|---------|
| PostgreSQL 17 | Database (installed on Windows) |
| Go modules | Dependency management |
| `.env` file | Environment configuration |

---

## 3. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        BROWSER                                  │
│                                                                 │
│  http://localhost:8080  →  Portfolio page (index.html)          │
│  http://localhost:8080/login  →  Login page                     │
│  http://localhost:8080/dashboard  →  Admin dashboard            │
└───────────────────────────┬─────────────────────────────────────┘
                            │  HTTP
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    GO SERVER (port 8080)                         │
│                                                                 │
│  Static routes:   / → index.html                                │
│                   /login → login.html                           │
│                   /dashboard → dashboard.html                   │
│                   /assets/* → CSS, JS, images                   │
│                   /uploads/* → uploaded files                   │
│                                                                 │
│  Middleware:  Logger → CORS → [Auth for protected routes]       │
│                                                                 │
│  API routes:  /api/v1/*  →  Handlers → Services → GORM → DB    │
└───────────────────────────┬─────────────────────────────────────┘
                            │  GORM (SQL)
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                  POSTGRESQL DATABASE                             │
│                                                                 │
│  Tables: users, profiles, projects, skills, experiences,        │
│          testimonials, messages                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Backend Layer Pattern

```
HTTP Request
    ↓
Middleware (logger, CORS, auth)
    ↓
Handler (reads request, validates input, calls service)
    ↓
Service (business logic, database operations via GORM)
    ↓
Model (GORM struct → PostgreSQL table)
    ↓
HTTP Response (standard JSON: { success, message, data, error })
```

### Frontend Layer Pattern

```
index.html (DOM structure)
    ↓
main.js (orchestrator — calls all section init functions)
    ↓
Section JS files (hero.js, about.js, techstack.js, etc.)
    ↓
API calls via core/api.js (all requests go through one place)
    ↓
Store (core/store.js — shared state between components)
    ↓
Components (AppShowcase.js, PhoneMockup.js, SkillBar.js, etc.)
    ↓
DOM updated with rendered HTML
    ↓
ScrollAnimations.init() — observes all .scroll-animation elements
```

---

## 4. Complete Folder Structure

```
personal_portfolio/
│
├── README.md                          ← This file
│
├── frontend/
│   ├── public/
│   │   ├── index.html                 ← Portfolio (public-facing, served at /)
│   │   ├── login.html                 ← Admin login page (/login)
│   │   └── dashboard.html             ← Admin dashboard (/dashboard)
│   │
│   └── assets/
│       ├── css/
│       │   ├── variables.css          ← All CSS custom properties (colors, fonts, spacing)
│       │   ├── base.css               ← CSS reset, typography, scroll animation classes
│       │   ├── main.css               ← Master import file — imports all CSS in order
│       │   ├── components/
│       │   │   ├── sidebar.css        ← Left floating card sidebar
│       │   │   ├── navbar.css         ← Right-side icon nav + mobile hamburger
│       │   │   ├── button.css         ← Button variants (primary, outline, ghost)
│       │   │   ├── card.css           ← Card base + stat card variants
│       │   │   ├── badge.css          ← Badge/tag elements
│       │   │   ├── avatar.css         ← Avatar/profile photo styles
│       │   │   ├── modal.css          ← Modal overlay and dialog
│       │   │   ├── form.css           ← Input, textarea, select, validation states
│       │   │   ├── toast.css          ← Toast notification popups
│       │   │   ├── loader.css         ← Page loader + inline spinner
│       │   │   ├── skill-bar.css      ← Animated skill percentage bars
│       │   │   ├── timeline.css       ← Experience timeline with left border line
│       │   │   ├── phone-mockup.css   ← 3D phone frame for app screenshots
│       │   │   ├── app-slider.css     ← Full-width apps showcase slider
│       │   │   ├── tech-stack.css     ← Two-column tech stack layout
│       │   │   ├── architecture-card.css ← Architecture expertise cards
│       │   │   ├── github-stats.css   ← GitHub stats grid
│       │   │   ├── stat-card.css      ← Hero section stat counters
│       │   │   ├── table.css          ← Dashboard data tables
│       │   │   └── portfolio-card.css ← Dashboard project cards
│       │   └── pages/
│       │       └── portfolio.css      ← Page layout: sidebar offset, hero, sections
│       │
│       ├── js/
│       │   ├── core/
│       │   │   ├── api.js             ← ALL HTTP requests go through here (base URL, JWT, 401 redirect)
│       │   │   ├── store.js           ← Reactive shared state (set/get/subscribe)
│       │   │   ├── utils.js           ← Helpers: animateCounter, debounce, isInViewport, parseArray
│       │   │   ├── auth.js            ← Login/logout, token storage, requireAuth guard
│       │   │   └── validator.js       ← Form validation (isEmail, isEmpty, showErrors)
│       │   │
│       │   ├── components/
│       │   │   ├── ScrollAnimations.js ← IntersectionObserver scroll animations (.scroll-animation)
│       │   │   ├── Toast.js           ← Toast notifications (success/error/info)
│       │   │   ├── Sidebar.js         ← Active nav link tracking on scroll
│       │   │   ├── Navbar.js          ← Mobile hamburger open/close
│       │   │   ├── SkillBar.js        ← Renders animated skill bars from data
│       │   │   ├── PhoneMockup.js     ← 3D phone frame with cycling screenshots
│       │   │   └── AppShowcase.js     ← Full-width app slider with thumbnails
│       │   │
│       │   ├── pages/portfolio/
│       │   │   ├── hero.js            ← Stat counter animations
│       │   │   ├── about.js           ← Loads profile from API, updates DOM
│       │   │   ├── techstack.js       ← Loads skills from API, groups by category
│       │   │   ├── architecture.js    ← Static architecture cards (Clean Arch, BLoC, etc.)
│       │   │   ├── apps.js            ← Loads projects from API, renders AppShowcase
│       │   │   ├── experience.js      ← Loads experience from API, renders timeline
│       │   │   ├── github.js          ← Static GitHub stats + readme-stats embeds
│       │   │   ├── services.js        ← Static service cards (3 cards)
│       │   │   └── contact.js         ← Contact form submission handler
│       │   │
│       │   └── main.js                ← Orchestrator: initializes everything in correct order
│       │
│       └── images/
│           ├── profile/               ← Profile photo (placeholder.jpg) + CV (cv.pdf) go here
│           ├── apps/                  ← App screenshots go here
│           ├── logos/                 ← Tech logos/icons go here
│           └── mockups/               ← Phone mockup images go here
│
├── backend/
│   ├── cmd/
│   │   └── main.go                    ← Entry point: loads config, connects DB, runs server
│   │
│   ├── config/
│   │   ├── config.go                  ← Loads all env vars into Config struct
│   │   └── database.go                ← Opens PostgreSQL connection with GORM
│   │
│   ├── internal/
│   │   ├── models/                    ← GORM structs = database tables
│   │   │   ├── user.go                ← Admin user (id, name, email, password, role)
│   │   │   ├── project.go             ← App/project entries for showcase
│   │   │   ├── skill.go               ← Tech stack skills with percentages
│   │   │   ├── experience.go          ← Work history timeline entries
│   │   │   ├── testimonial.go         ← Client reviews (optional)
│   │   │   ├── message.go             ← Contact form submissions
│   │   │   └── profile.go             ← Portfolio profile/bio data
│   │   │
│   │   ├── services/                  ← Business logic + database operations
│   │   │   ├── auth_service.go        ← Login: verify email/password, generate JWT
│   │   │   ├── project_service.go     ← CRUD for projects
│   │   │   ├── skill_service.go       ← CRUD for skills
│   │   │   ├── experience_service.go  ← CRUD for experience
│   │   │   ├── message_service.go     ← CRUD + mark-read for messages
│   │   │   ├── profile_service.go     ← Get/update single profile record
│   │   │   └── upload_service.go      ← Validates + saves uploaded files
│   │   │
│   │   ├── handlers/                  ← HTTP layer: read request → call service → send response
│   │   │   ├── auth_handler.go        ← POST /auth/login, POST /auth/logout, GET /auth/me
│   │   │   ├── project_handler.go     ← CRUD /projects
│   │   │   ├── skill_handler.go       ← CRUD /skills
│   │   │   ├── experience_handler.go  ← CRUD /experience
│   │   │   ├── message_handler.go     ← GET/POST/PUT/DELETE /messages
│   │   │   ├── profile_handler.go     ← GET/PUT /profile
│   │   │   ├── upload_handler.go      ← POST /upload
│   │   │   └── stats_handler.go       ← GET /stats (dashboard counts)
│   │   │
│   │   ├── middleware/
│   │   │   ├── auth_middleware.go     ← Validates Bearer JWT token on protected routes
│   │   │   ├── cors_middleware.go     ← Sets CORS headers from ALLOWED_ORIGINS
│   │   │   └── logger_middleware.go   ← Logs every request: method, path, status, latency
│   │   │
│   │   ├── routes/
│   │   │   └── routes.go              ← All route registrations (public + protected + static files)
│   │   │
│   │   └── utils/
│   │       ├── jwt.go                 ← GenerateToken, ValidateToken (HMAC-SHA256)
│   │       ├── hash.go                ← HashPassword (bcrypt cost 12), CheckPassword
│   │       ├── response.go            ← Standard API response helpers (Success, BadRequest, etc.)
│   │       └── validator.go           ← IsValidEmail, IsEmpty, Sanitize, IsValidPercentage
│   │
│   ├── migrations/
│   │   └── 001_create_users.sql       ← Reference SQL (GORM AutoMigrate handles actual creation)
│   │
│   ├── seeds/
│   │   └── seed.go                    ← Seeds admin user + Imtiaz's profile, skills, experience, projects
│   │
│   ├── uploads/                       ← Uploaded files saved here (images, CV PDF)
│   ├── .env                           ← Environment variables (DO NOT commit to git)
│   ├── .env.example                   ← Template showing all required variables
│   ├── go.mod                         ← Go module dependencies
│   └── go.sum                         ← Dependency checksums
│
├── docs/
│   ├── api.md                         ← API reference
│   ├── database.md                    ← Schema reference
│   └── deployment.md                  ← Deployment steps
│
└── scripts/
    ├── setup.sh                       ← Full project setup script
    ├── run.sh                         ← Start the server
    └── migrate.sh                     ← Run SQL migration files manually
```

---

## 5. Getting Started

### Prerequisites

- **Go** 1.21 or higher — [download](https://go.dev/dl/)
- **PostgreSQL** — installed via Homebrew on Mac (office), or at `C:\Program Files\PostgreSQL\17` on Windows (home)
- A modern web browser (Chrome recommended)

### First-time Setup

**Step 1 — Start PostgreSQL**

**Mac (office):**
```bash
brew services start postgresql@17
```

**Windows (home):**
```powershell
Start-Service "postgresql-x64-17"
```

**Step 2 — Create the database**

The database was already created. If you ever need to recreate it:

**Mac (office):**
```bash
createdb -U imtiazchowdhury imtiaz_portfolio
```

**Windows (home):**
```powershell
$env:PGPASSWORD = "postgres123"
& "C:\Program Files\PostgreSQL\17\bin\createdb.exe" -U postgres -h 127.0.0.1 imtiaz_portfolio
```

**Step 3 — Check your `.env` file**

Located at `backend/.env`. Values differ by machine — see section 6.

**Mac (office):**
```env
SERVER_PORT=8080
APP_ENV=development
DB_HOST=localhost
DB_PORT=5432
DB_USER=imtiazchowdhury
DB_PASSWORD=
DB_NAME=imtiaz_portfolio
DB_SSLMODE=disable
JWT_SECRET=imtiaz-portfolio-secret-key-2024-change-in-production
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE_MB=10
ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:5500,http://localhost:5500
ADMIN_EMAIL=admin@imtiaz.dev
ADMIN_PASSWORD=Admin@1234
```

**Windows (home):**
```env
SERVER_PORT=8080
APP_ENV=development
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=imtiaz_portfolio
DB_SSLMODE=disable
JWT_SECRET=imtiaz-portfolio-secret-key-2024-change-in-production
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE_MB=10
ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:5500,http://localhost:5500
ADMIN_EMAIL=admin@imtiaz.dev
ADMIN_PASSWORD=Admin@1234
```

---

## 6. Credentials

### PostgreSQL Database

**Mac (office):**
| Field | Value |
|-------|-------|
| Host | localhost |
| Port | 5432 |
| Username | imtiazchowdhury |
| Password | *(none)* |
| Database | imtiaz_portfolio |

**Windows (home):**
| Field | Value |
|-------|-------|
| Host | localhost |
| Port | 5432 |
| Username | postgres |
| Password | postgres123 |
| Database | imtiaz_portfolio |

### Admin Dashboard Login
| Field | Value |
|-------|-------|
| URL | http://localhost:8080/login |
| Email | admin@imtiaz.dev |
| Password | Admin@1234 |

> **Important:** Change the admin password after first login in production. Update `ADMIN_PASSWORD` in `.env` and reset via the dashboard.

---

## 7. How to Run

### Start the server (one command does everything)

**Mac (office):**
```bash
cd ~/Development/Projects/Portfolio/personal_portfolio/backend
go run ./cmd/main.go
```

**Windows (home):**
```powershell
cd I:\website\personal_portfolio\backend
go run ./cmd/main.go
```

**What happens automatically on startup:**
1. Loads `.env` configuration
2. Connects to PostgreSQL (`imtiaz_portfolio` database)
3. Runs GORM AutoMigrate — creates/updates all 7 tables
4. Runs seed — creates admin account + Imtiaz's profile, 17 skills, 3 experience entries, 2 projects (only if database is empty)
5. Creates `/uploads` directory if missing
6. Registers all routes (static files + API)
7. Server starts listening on `http://localhost:8080`

### Open the site

Open Chrome and navigate to:
- **Portfolio:** `http://localhost:8080`
- **Admin login:** `http://localhost:8080/login`
- **Dashboard:** `http://localhost:8080/dashboard`

### Stop the server

Press `Ctrl+C` in the terminal. The server shuts down gracefully (waits for in-flight requests to finish).

### If port 8080 is already in use

**Mac (office):**
```bash
lsof -ti :8080 | xargs kill -9
```

**Windows (home):**
```powershell
$p = Get-NetTCPConnection -LocalPort 8080 | Where-Object State -eq "Listen" | Select-Object -First 1 OwningProcess
Stop-Process -Id $p.OwningProcess -Force
```

---

## 8. How Everything Works (Data Flow)

### Portfolio page loads

```
1. Browser requests http://localhost:8080
2. Go server responds with frontend/public/index.html
3. Browser loads /assets/css/main.css (all component CSS)
4. Browser loads all JS files in order:
   - CDN: swiper.min.js
   - core/: api.js → store.js → utils.js → auth.js → validator.js
   - components/: ScrollAnimations.js → Toast.js → Sidebar.js → etc.
   - pages/portfolio/: hero.js → about.js → techstack.js → etc.
   - main.js (orchestrator, runs last)
5. DOMContentLoaded fires → main.js runs:
   a. Sidebar.init() — activates scroll-based nav highlighting
   b. Navbar.init() — wires hamburger menu
   c. HeroSection.init() — starts stat counter animations
   d. ArchitectureSection.init() — renders static arch cards
   e. ServicesSection.init() — renders static service cards
   f. GithubSection.init() — renders static GitHub stats
   g. ContactSection.init() — wires contact form
   h. await Promise.allSettled([
        AboutSection.init()      → GET /api/v1/profile → updates hero, sidebar, about
        TechStackSection.init()  → GET /api/v1/skills → groups by category, renders bars
        AppsSection.init()       → GET /api/v1/projects → renders Swiper slider
        ExperienceSection.init() → GET /api/v1/experience → renders timeline
      ])
   i. ScrollAnimations.init() — observes all .scroll-animation elements
   j. Page loader hidden
```

### Admin logs in

```
1. Admin visits http://localhost:8080/login
2. Enters email + password → form submits
3. POST /api/v1/auth/login → AuthHandler.Login
4. AuthService.Login: finds user by email, bcrypt.Compare(password, hash)
5. If valid: generates JWT (HMAC-SHA256, 24h expiry)
6. Response: { token: "eyJ...", user: { id, name, email, role } }
7. auth.js stores token in localStorage as 'portfolio_token'
8. Redirected to /dashboard
9. dashboard.html loads → Auth.requireAuth() → GET /api/v1/auth/me
10. If token valid: dashboard loads stats and content
11. If token expired: cleared from localStorage, redirected to /login
```

### JWT Authentication flow (protected routes)

```
Frontend request to protected endpoint:
  Headers: { Authorization: "Bearer eyJ..." }
              ↓
  AuthMiddleware reads header
  ValidateToken(tokenString, JWT_SECRET)
  Checks: signature valid? expiry not passed?
              ↓
  If valid: stores userID, email, role in request context
  Handler reads them: c.GetUint("userID")
              ↓
  If invalid: 401 Unauthorized → frontend clears token → redirect to /login
```

### Contact form submission

```
Visitor fills form (name, email, type, message)
    ↓
Validator.validate() runs client-side checks
    ↓
POST /api/v1/messages (public endpoint, no auth needed)
    ↓
MessageHandler.Create validates + saves to 'messages' table
Records visitor's IP address
    ↓
Response: 201 Created → Toast shows "Message sent!"
    ↓
Admin sees unread badge in dashboard sidebar
Admin can view, mark as read, or delete in Messages section
```

### Scroll animations

```
After all content renders:
    ↓
ScrollAnimations.init() runs
    ↓
querySelectorAll('.scroll-animation') finds all animated elements
    ↓
IntersectionObserver watches each one with rootMargin: '0px 0px -5% 0px'
    ↓
When element enters viewport:
  .is-visible class added
  CSS transition: opacity 0 → 1, transform translateY(80px) → 0
  Duration: 1.2s, cubic-bezier(0.16, 1, 0.3, 1) [= GSAP power4.out]
    ↓
When element leaves viewport:
  .is-visible removed → element fades back out (same as Drake)
```

---

## 9. All API Endpoints

**Base URL:** `http://localhost:8080/api/v1`

**Standard response shape (every endpoint):**
```json
{
  "success": true,
  "message": "Human-readable description",
  "data": { },
  "error": null
}
```

### Authentication

| Method | Endpoint | Auth | Body | Description |
|--------|----------|------|------|-------------|
| POST | /auth/login | No | `{ email, password }` | Returns JWT token |
| POST | /auth/logout | No | — | Client-side logout confirmation |
| GET | /auth/me | ✅ | — | Returns current user info |

**Login request:**
```json
{ "email": "admin@imtiaz.dev", "password": "Admin@1234" }
```
**Login response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": { "id": 1, "name": "Chowdhury Md. Imtiazul Islam", "email": "admin@imtiaz.dev", "role": "admin" }
  }
}
```

### Projects

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /projects | No | All projects (sorted by featured, then sort_order) |
| GET | /projects/:id | No | Single project by ID |
| POST | /projects | ✅ | Create new project |
| PUT | /projects/:id | ✅ | Update project |
| DELETE | /projects/:id | ✅ | Soft-delete project |

### Skills

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /skills | No | All skills (sorted by category, then sort_order) |
| POST | /skills | ✅ | Create skill |
| PUT | /skills/:id | ✅ | Update skill |
| DELETE | /skills/:id | ✅ | Delete skill |

### Experience

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /experience | No | All experience entries |
| POST | /experience | ✅ | Create experience entry |
| PUT | /experience/:id | ✅ | Update experience entry |
| DELETE | /experience/:id | ✅ | Delete experience entry |

### Messages

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /messages | No | Submit contact form (public) |
| GET | /messages | ✅ | Get all messages (dashboard) |
| PUT | /messages/:id | ✅ | Mark message as read |
| DELETE | /messages/:id | ✅ | Delete message |

### Profile

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /profile | No | Get portfolio profile |
| PUT | /profile | ✅ | Update profile (preserves untouched fields) |
| PUT | /profile/photo | ✅ | Update sidebar profile photo only |
| PUT | /profile/about-photo | ✅ | Update About-section photo only |
| PUT | /profile/cv | ✅ | Update active CV path. Also appends a row to `cv_files` so every saved CV is kept in version history. Body: `{ cv_file, file_name, file_size }` |
| PUT | /profile/cv/visibility | ✅ | Toggle the public "Download CV" button on/off. Body: `{ cv_visible: bool }` |
| POST | /profile/cv/generate | ✅ | Build a fresh PDF from profile + skills + experience, save to `/uploads/cv/`, append to history, set as active |
| GET | /profile/cv/history | ✅ | List every saved CV (newest first). Each row carries `is_active` |
| POST | /profile/cv/history/:id/activate | ✅ | Set this CV as the active one (mirrors its path onto `profiles.cv_file`) |
| DELETE | /profile/cv/history/:id | ✅ | Delete one CV from history + remove the file from disk. Refuses to delete the active CV |
| POST | /profile/cv/download | ❌ | **Public** — increments `profiles.cv_download_count`. Fired by the portfolio when a visitor clicks "Download CV" |
| PUT | /profile/social | ✅ | Update social links only (GitHub, LinkedIn, Twitter, Instagram, WhatsApp). Empty values clear the field. |
| PUT | /profile/github | ✅ | Update GitHub Stats fields (username + 4 manual overrides) |
| PUT | /profile/mail | ✅ | Update SMTP credentials only. Records each new password in `mail_password_histories`. |
| PUT | /profile/footer | ✅ | Update the four footer fields only (`copyright_text`, `sidebar_copyright_text`, `footer_copyright_text`, `footer_built_with`). Empty values are written through (allowing clear). |
| POST | /profile/mail/verify | ✅ | Test arbitrary `{ smtp_user, smtp_pass }` against Gmail without sending mail. Returns `{ valid, error }`. |
| GET | /profile/mail/history | ✅ | List every saved Gmail App Password (newest first). Each row carries `is_active` and `is_hidden` flags. |
| DELETE | /profile/mail/history/:id | ✅ | Delete a single password-history row |
| POST | /profile/mail/accounts/:id/hide | ✅ | Hide an account from the dashboard's Available Accounts panel without touching its history |
| POST | /profile/mail/accounts/:id/unhide | ✅ | Restore a hidden account |

### Upload & Stats

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /upload | ✅ | Upload file (image or PDF). Form field: `file`. Optional: `folder` |
| GET | /stats | ✅ | Dashboard counts: projects, skills, messages, unread |
| GET | /health | No | Server health check |

---

## 10. Database Schema

### users
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL | Primary key |
| name | VARCHAR(100) | Display name |
| email | VARCHAR(255) | Unique, used for login |
| password | VARCHAR(255) | bcrypt hash (cost 12) |
| role | VARCHAR(50) | Always "admin" currently |
| avatar | VARCHAR(500) | Path to avatar image |
| created_at | TIMESTAMPTZ | Auto-set |
| updated_at | TIMESTAMPTZ | Auto-updated |
| deleted_at | TIMESTAMPTZ | Soft delete (GORM) |

### profiles
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL | Always 1 (single profile) |
| full_name | VARCHAR(200) | |
| title | VARCHAR(200) | |
| tagline | VARCHAR(500) | |
| bio | TEXT | Long bio for About section |
| short_bio | VARCHAR(500) | Short bio for hero |
| nickname | VARCHAR(100) | Greeting name shown in the hero pill ("Say Hi from {nickname}…"). Empty = first word of `full_name` |
| hero_subtitle | VARCHAR(300) | Full override for the hero pill text. Supports `{name}`, `{title}`, `{nickname}` placeholders. Empty = template "Say Hi from {nickname}, {title}" |
| hero_heading_line1 | VARCHAR(200) | First line of the big hero heading |
| hero_heading_line2 | VARCHAR(200) | Second line of the big hero heading |
| hero_heading_line3 | VARCHAR(200) | Third line of the big hero heading |
| hero_heading_highlights | VARCHAR(500) | Comma-separated words/phrases to colour brand-blue inside the three lines |
| hero_heading_override | TEXT | Free-form heading override — newlines = line breaks, `*text*` = highlight. Takes precedence over the line + highlight fields when set |
| email | VARCHAR(255) | Contact email |
| phone | VARCHAR(50) | |
| whats_app | VARCHAR(50) | WhatsApp number |
| location | VARCHAR(200) | |
| git_hub | VARCHAR(500) | GitHub profile URL |
| linked_in | VARCHAR(500) | LinkedIn profile URL |
| twitter | VARCHAR(500) | |
| instagram | VARCHAR(500) | |
| profile_photo | VARCHAR(500) | Path to sidebar profile photo |
| about_photo | VARCHAR(500) | Path to About-section photo (separate image) |
| cv_file | VARCHAR(500) | Path to currently active CV PDF (mirrored from `cv_files`) |
| cv_visible | BOOLEAN | Default `true`. When false, hides the public "Download CV" button without deleting the file |
| cv_download_count | BIGINT | Incremented every time a visitor clicks the public "Download CV" button |
| availability | VARCHAR(100) | "Open to Work" etc. |
| years_experience | VARCHAR(20) | "2.5+" |
| apps_shipped | VARCHAR(20) | "5+" |
| reply_emails | TEXT | Comma-separated list shown in the dashboard inbox's "From" picker |
| copyright_text | VARCHAR(200) | Default copyright shown in both sidebar + footer when no override. Supports `{year}` / `{name}` placeholders. Empty = auto "© {year} {full_name}. All Rights Reserved." |
| sidebar_copyright_text | VARCHAR(200) | Sidebar-only copyright override |
| footer_copyright_text | VARCHAR(200) | Footer-only copyright override |
| footer_built_with | VARCHAR(200) | Optional second line under the footer copyright (e.g. "Built with Flutter spirit 💙"). Empty hides the line |
| smtp_user | VARCHAR(255) | Sending Gmail address (overrides `.env` if non-empty) |
| smtp_pass | VARCHAR(255) | Gmail App Password (overrides `.env` if non-empty) |
| git_hub_username | VARCHAR(100) | Drives live GitHub API fetch on the public portfolio |
| git_hub_repos | VARCHAR(20) | Manual override for the Repositories card. Blank = auto. |
| git_hub_commits | VARCHAR(20) | Manual-only value for Total Commits (no cheap API). Blank = static fallback. |
| git_hub_top_language | VARCHAR(50) | Manual override for Top Language. Blank = auto from recent repos. |
| git_hub_years_active | VARCHAR(20) | Manual override for Years Active. Blank = computed from account creation date. |
| meta_title | VARCHAR(200) | SEO title |
| meta_description | VARCHAR(500) | SEO description |

### projects
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL | |
| name | VARCHAR(200) | App name |
| slug | VARCHAR(200) | Unique URL slug |
| description | TEXT | Short description |
| long_description | TEXT | Full description |
| tech_stack | TEXT | JSON array: `["Flutter","Firebase"]` |
| features | TEXT | JSON array of bullet points |
| screenshots | TEXT | JSON array of image paths/URLs |
| thumbnail | VARCHAR(500) | Main preview image |
| app_store_url | VARCHAR(500) | iOS App Store link |
| play_store_url | VARCHAR(500) | Google Play link |
| github_url | VARCHAR(500) | Source repo (if public) |
| status | VARCHAR(50) | "live" / "development" / "archived" |
| featured | BOOLEAN | Featured apps appear first |
| sort_order | BIGINT | Display order |

### skills
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL | |
| name | VARCHAR(100) | e.g. "Flutter" |
| category | VARCHAR(100) | "core" / "state_management" / "backend" / "payments" / "tools" |
| percentage | BIGINT | 0–100, validated by DB constraint |
| description | TEXT | Shown on hover |
| icon | VARCHAR(500) | Path to logo image |
| sort_order | BIGINT | Order within category |

### experiences
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL | |
| company | VARCHAR(200) | Employer name |
| role | VARCHAR(200) | Job title |
| start_date | VARCHAR(50) | e.g. "Jan 2024" |
| end_date | VARCHAR(50) | Empty string = "Present" |
| is_current | BOOLEAN | Marks active job |
| description | TEXT | Role description |
| achievements | TEXT | JSON array of bullet points |
| tech_used | TEXT | JSON array of technology names |
| company_logo | VARCHAR(500) | Logo image path |
| location | VARCHAR(200) | |
| type | VARCHAR(50) | "full-time" / "internship" / "freelance" |
| sort_order | BIGINT | Lower = shown first |

### messages
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL | |
| name | VARCHAR(200) | Sender name |
| email | VARCHAR(255) | Sender email |
| type | VARCHAR(100) | "Job Opportunity" / "Freelance" / "Other" |
| content | TEXT | Message body |
| is_read | BOOLEAN | Read status (dashboard badge) |
| is_replied | BOOLEAN | Replied status |
| ip_address | VARCHAR(50) | Sender IP for spam filtering |

### testimonials
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL | |
| author_name | VARCHAR(200) | |
| author_title | VARCHAR(200) | Job title of reviewer |
| author_avatar | VARCHAR(500) | Photo path |
| content | TEXT | Review text |
| rating | BIGINT | 1–5 stars |
| is_published | BOOLEAN | Only published ones show publicly |
| sort_order | BIGINT | |

### message_replies
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL | |
| message_id | BIGINT | FK to `messages.id` |
| from_email | VARCHAR(255) | Address the reply was sent from |
| subject | VARCHAR(500) | |
| body | TEXT | |
| attachments | TEXT | JSON array of uploaded file URLs |
| sent_at | TIMESTAMPTZ | |

### mail_password_histories
Append-only audit log of every Gmail App Password ever saved through the dashboard.

| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL | |
| gmail_address | VARCHAR(255) | The sending Gmail address paired with this password |
| app_password | VARCHAR(255) | **Stored in plaintext at the admin's explicit request** — treat as sensitive |
| created_at | TIMESTAMPTZ | When the password was saved |

The `is_active` and `is_hidden` flags returned by `GET /profile/mail/history` are computed at read time, not stored.

### cv_files
Append-only version history of every CV ever uploaded or auto-generated. The currently active CV's path is mirrored onto `profiles.cv_file` so the public portfolio doesn't need a join.

| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL | |
| file_path | VARCHAR(500) | Public path under `/uploads/cv/` |
| file_name | VARCHAR(255) | Display name shown in the dashboard history list |
| file_size | BIGINT | Bytes (0 if not recorded) |
| source | VARCHAR(20) | `"uploaded"` or `"generated"` |
| created_at | TIMESTAMPTZ | When the CV was saved |

The `is_active` flag returned by `GET /profile/cv/history` is computed at read time, not stored.

### hidden_mail_accounts
Lightweight list of Gmail addresses the admin has chosen to hide from the dashboard's Available Accounts panel. History rows for the email are preserved, only the deduped view filters them out.

| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL | |
| email | VARCHAR(255) | Unique — one row per hidden Gmail address |
| hidden_at | TIMESTAMPTZ | When the account was hidden |

---

## 11. Frontend Pages

### `index.html` — Portfolio (public)

**Purpose:** The main portfolio page visitors see.

**Sections in order:**
1. **Hero** — Greeting pill, large light heading, rotating badge, bio, CTA buttons, stat counters (2.5+ years, 5+ apps, 10+ technologies)
2. **About** — Profile photo + bio text + info cards (location, availability, experience, focus)
3. **Tech Stack** — Skills loaded from API, grouped: Core | State Management | Backend | Payments | Tools. Animated percentage bars.
4. **Architecture** — Static cards: Clean Architecture, BLoC, GetX, MVVM, Repository Pattern — each with ASCII diagram
5. **Apps Showcase** — Swiper slider. Each slide: app name, description, tech tags, features, App Store/Play Store badges. Phone mockup with cycling screenshots.
6. **Experience Timeline** — Work history from API: SperkTech (current), SoftVence, DIU internship
7. **GitHub Stats** — Static stat cards + github-readme-stats image embeds
8. **What I Offer** — 3 static service cards: App Development, Maintenance, API Integration
9. **Contact** — Two cards (Job / Freelance) + contact form (saved to database) + CV download
10. **Footer** — Name, tagline, social icons, copyright

**APIs called by this page:**
- `GET /api/v1/profile` → About section, hero bio, sidebar info
- `GET /api/v1/skills` → Tech Stack section
- `GET /api/v1/projects` → Apps Showcase
- `GET /api/v1/experience` → Experience Timeline

### `login.html` — Admin Login

**Purpose:** Authentication gate for the dashboard.

- If already logged in (valid token in localStorage) → auto-redirects to `/dashboard`
- Form validates email format before submitting
- On success → stores JWT in localStorage → redirects to `/dashboard`
- On failure → shows error message inline

### `dashboard.html` — Admin Dashboard

**Purpose:** Content management for all portfolio data.

**Sections (sidebar tabs):**
- **📊 Overview** — Stat cards: total projects, skills, messages, unread count
- **📱 Projects** — Table of all apps with delete button
- **⚡ Skills** — Table of all skills with delete button
- **💼 Experience** — Table of work history with delete button
- **✉ Messages** — WhatsApp-style inbox: list + chat view, replies with attachments
- **👤 Profile** — Split layout: sticky left identity card (two photo uploads with click-to-upload + hover overlay + Cropper.js modal, live name/title/email/phone/location preview, status pill); right side organised into Basic Info, **Hero Heading** (3 lines + highlight words + advanced override + live preview + Restore default), **Hero Pill** (nickname + subtitle override), Status (3 visual cards: Open to Work / Available for Freelance / 🌴 On Vacation), and About You (Short Bio + Long Bio with character count). Phone field uses `intl-tel-input` country picker.
- **📄 CV / Resume** — Hero card for active CV, Upload + auto-Generate cards, stats row (downloads / versions / last updated), iOS-style visibility toggle, in-dashboard PDF preview (PDF.js canvas rendering for exact-fit height), version history (featured active card + grid of past tiles with stylized PDF-cover artwork, three-dot menu, click-twice-confirm delete). **⫴ Compare** modal: side-by-side panes with dropdown + chip strip for swapping any saved CV into either pane + per-pane "Make Active" button.
- **🔗 Social Links** — GitHub, LinkedIn, Twitter, Instagram, WhatsApp. WhatsApp uses an `intl-tel-input` country code picker. Saving an empty value clears that link on the public portfolio.
- **📧 Mail Settings** — Sending Gmail address + 16-character App Password (4×4 boxes with show/hide). Auto-tests credentials against Gmail when the tab opens. Below the form:
  - **Available Accounts** — One card per unique Gmail ever saved (deduped from history). Each card shows the password as 4 boxes, an auto-test status, and buttons: Show / Copy / Use this account / Hide. Hidden accounts can be revealed via the "Show N hidden" toggle and restored.
  - **App Password History** — Append-only audit table of every save. Each row auto-tests in parallel and shows ✅ Valid / ❌ Invalid with the SMTP error on hover.
- **🐙 GitHub Stats** — Username + 4 manual override inputs (Repositories, Total Commits, Top Language, Years Active). Blank override = use auto-fetched value from GitHub API on the public portfolio.
- **📜 Footer & Copyright** — Live preview of sidebar + footer copyright lines, main copyright field with `{year}`/`{name}` placeholders, 5 quick-fill presets, insert chips for placeholders + ©/❤️/✨/💙/🚀/·, optional "Built with" second line, collapsible per-place override section. Saves via dedicated `PUT /profile/footer`.

**Auth:** Every page load calls `GET /api/v1/auth/me`. If it returns 401, the page redirects to `/login` immediately.

---

## 12. Component Library

### ScrollAnimations.js
Add `class="scroll-animation"` to any element. Optional `data-animation` attribute:
- `fade_from_bottom` (default) — slides up from 80px below
- `fade_from_top` — slides down from 80px above
- `fade_from_left` — slides in from 80px left
- `fade_from_right` — slides in from 80px right
- `fade_in` — fades in with no movement

Reverse animation plays when element scrolls out of view.

### Toast.js
```javascript
Toast.show('Message sent!', 'success');      // green
Toast.show('Something failed', 'error');     // red
Toast.show('Check this out', 'info');        // blue
Toast.show('Be careful', 'warning');         // orange
```
Auto-dismisses after 4 seconds. Manual close button on each toast.

### AppShowcase.js
```javascript
AppShowcase.render({
  container: document.querySelector('#apps-showcase'),
  apps: projectsArray,   // from API
  autoPlay: true,
  interval: 6000
});
```

### PhoneMockup.js
```javascript
PhoneMockup.render({
  container: document.querySelector('.phone-slot'),
  screenshots: ['url1.jpg', 'url2.jpg'],
  interval: 3000
});
```

### SkillBar.js
```javascript
SkillBar.render({
  container: document.querySelector('#skills-container'),
  skills: skillsArray  // from API
});
```

### API.js (core)
```javascript
// All requests go through here — auto-attaches JWT, handles 401
const res = await API.get('/projects');
const projects = res.data;

const created = await API.post('/messages', { name, email, content });
await API.put('/profile', updatedProfile);
await API.delete('/projects/1');
const res = await API.upload('/upload', formData);
```

### Store.js (core)
```javascript
Store.set('projects', data);          // save + notify subscribers
const projects = Store.get('projects'); // read current value
Store.subscribe('projects', (data) => {  // react to changes
  AppShowcase.render({ apps: data });
});
```

---

## 13. Design System

### Colors
| Variable | Value | Used for |
|----------|-------|---------|
| `--color-primary` | `#54c5f8` | Flutter blue — buttons, highlights, accents |
| `--color-secondary` | `#e040fb` | Purple — used sparingly |
| `--color-bg` | `#0a0a0f` | CSS variable (body overridden to `#1f1f1f`) |
| Body background | `#1f1f1f` | Drake's exact background color |
| Border color | `#565656` | Drake's exact border color |
| Text color | `#999999` | Drake's muted text color |
| Heading color | `#ffffff` | All headings |

### Typography
- **Font:** Inter (300, 400, 500, 600, 700, 900)
- **Hero heading:** 78px, font-weight 300 (light) — Drake's signature style
- **Section headings:** 48px, font-weight 300
- **Body text:** 16px, color `#999999`

### Spacing
All sections: `padding: 90px 0` (matching Drake's section spacing)

### Animations
- Duration: 1.2s
- Easing: `cubic-bezier(0.16, 1, 0.3, 1)` = GSAP `power4.out`
- Elements: `.scroll-animation` + `.is-visible`
- Scroll trigger: element enters 5% above viewport bottom

### Button Style (Drake)
```css
background: var(--color-primary);
color: #000;
padding: 13px 48px;
border-radius: 30px;
text-transform: uppercase;
border: 2px solid var(--color-primary);
```

---

## 14. What Has Been Done So Far

### Backend ✅
- [x] Complete Go server with Gin framework
- [x] PostgreSQL database connection with GORM
- [x] 9 database models with proper GORM tags
- [x] Auto-migration on startup
- [x] Full seeder (admin + Imtiaz's real data)
- [x] JWT authentication (login, token validation, 401 handling)
- [x] bcrypt password hashing (cost 12)
- [x] 30+ API endpoints (public + protected)
- [x] JWT auth middleware
- [x] CORS middleware
- [x] Request logger middleware
- [x] File upload endpoint (images, PDF, up to 10MB)
- [x] Standard API response format across all endpoints
- [x] Graceful server shutdown
- [x] Go backend serves frontend (no separate dev server needed)
- [x] Gmail SMTP email reply system with attachment support
- [x] SMTP credentials stored in database (configurable from dashboard)
- [x] Message reply history stored per message (`message_replies` table)
- [x] Dedicated endpoints for cv_file and profile_photo partial updates
- [x] `Cache-Control: no-store` on profile endpoint (always fresh photo/CV)
- [x] Profile update preserves existing fields when partial data sent
- [x] Dedicated `PUT /profile/social` endpoint — empty values clear social links (preserve logic intentionally bypassed)
- [x] Dedicated `PUT /profile/mail` endpoint that auto-records every save into `mail_password_histories`
- [x] `mail_password_histories` table — append-only audit log of every Gmail App Password ever saved
- [x] `POST /profile/mail/verify` — connects to Gmail SMTP and runs AUTH PLAIN against arbitrary credentials without sending mail
- [x] `hidden_mail_accounts` table + hide/unhide endpoints — soft-hide accounts from the dashboard while keeping audit history intact
- [x] `PUT /profile/github` endpoint with 5 GitHub-related profile fields (username + 4 stat overrides)
- [x] `cv_files` table — append-only version history of every CV (uploaded or generated)
- [x] CV management endpoints: `PUT /profile/cv` (now appends history + activates), `PUT /profile/cv/visibility`, `POST /profile/cv/generate`, `GET /profile/cv/history`, `POST /profile/cv/history/:id/activate`, `DELETE /profile/cv/history/:id`
- [x] Public `POST /profile/cv/download` — increments `profiles.cv_download_count` (no auth)
- [x] CV auto-generator service (`gofpdf`) — basic working classic-resume layout from profile + skills + experience (needs design polish)
- [x] CV history backfill — `GetCVHistory` synthesises a `cv_files` row when `profile.cv_file` exists but has no matching history row (handles legacy uploads from before the table existed)
- [x] `PUT /profile/footer` endpoint — pointer-typed body for the four footer fields (`copyright_text`, `sidebar_copyright_text`, `footer_copyright_text`, `footer_built_with`); empty values clear, nil values preserve
- [x] Profile model expanded with: `nickname`, `hero_subtitle`, `hero_heading_line1/2/3`, `hero_heading_highlights`, `hero_heading_override`, `copyright_text`, `sidebar_copyright_text`, `footer_copyright_text`, `footer_built_with`

### Frontend ✅
- [x] Drake-inspired dark theme (`#1f1f1f` background)
- [x] Floating left sidebar card (Drake exact: 30px radius, border, padding 50px)
- [x] Right-side icon navigation (Phosphor Icons, pill card, tooltips on hover)
- [x] Mobile hamburger menu with fullscreen dropdown
- [x] Hero section: light heading (font-weight 300, 78px), rotating badge, bio, pill CTA buttons, 72px stat counters
- [x] About section: two-column layout, profile photo, info cards
- [x] Tech Stack section: API-driven, grouped by category, animated skill bars
- [x] Architecture section: Clean Arch, BLoC, GetX, MVVM, Repository Pattern with ASCII diagrams
- [x] Apps Showcase: Swiper slider, phone mockup with cycling screenshots, thumbnail strip
- [x] Experience Timeline: API-driven, company/role/dates/achievements/tech tags
- [x] GitHub Stats section: stat cards + github-readme-stats image embeds
- [x] Services section: 3 service cards
- [x] Contact section: two intent cards (Job/Freelance) + form that saves to DB
- [x] Footer with social links
- [x] Admin login page
- [x] Admin dashboard (overview, projects, skills, experience, messages, profile)
- [x] WhatsApp-style inbox — split panel, conversation list + chat view
- [x] Message reply with full chat history (all replies shown as bubbles)
- [x] Reply compose panel with From email selector, subject, body, attachments
- [x] File attachments in replies — WhatsApp-style image grid + doc cards
- [x] In-dashboard attachment lightbox with image carousel (← → navigation)
- [x] Attachment files uploaded to server, stored, viewable after reload
- [x] Gmail App Password field (4×4-digit input, show/hide toggle)
- [x] CV upload from dashboard (PDF, persists and drives Download CV buttons)
- [x] Profile photo upload with Cropper.js crop modal (drag/zoom/resize)
- [x] Separate sidebar photo and about section photo — two independent images
- [x] HEIC/HEIF support — iPhone photos auto-converted to JPEG before upload
- [x] Live crop preview showing how photo looks in Sidebar + About section
- [x] Sidebar "Hire Me" button hides on hero section, shows when scrolled past
- [x] Sidebar card shrinks/expands when Hire Me button shows/hides (animated)
- [x] Contact form success message auto-dismisses after 6 seconds
- [x] Scroll animation flickering fixed (once: true, GPU compositing hints)
- [x] Dashboard login session fix (stale token cleared before new login)
- [x] Scroll animations on ALL elements (IntersectionObserver + CSS)
- [x] No alternating section backgrounds (all `#1f1f1f` like Drake)
- [x] No background animation in hero (plain dark like Drake)
- [x] Phosphor Icons for sidebar social links and right navigation
- [x] All API calls centralized through `api.js`
- [x] Reactive store (`store.js`)
- [x] Toast notifications
- [x] Form validation
- [x] Page loader
- [x] `Ctrl+Shift+R` to see changes instantly (no build step needed)
- [x] Dashboard sidebar split — Profile / Social Links / Mail Settings / GitHub Stats are now separate tabs (was one giant Profile form)
- [x] Social Links tab with `intl-tel-input` country code picker for the WhatsApp field
- [x] Public portfolio's sidebar/footer/WhatsApp icons all driven by profile data via `data-social="…"` attributes — empty values hide the icon
- [x] Mail Settings auto-tests saved credentials against Gmail when the tab opens
- [x] Available Accounts panel — deduplicated cards per unique Gmail with parallel auto-test, ↺ Use this account, 🚫 Hide, ↺ Restore, "Show N hidden" toggle
- [x] App Password History panel — auto-tests every entry in parallel; Show / Copy / Use this / Delete-row controls
- [x] Click-twice confirmation pattern (no native `confirm()` — robust against Firefox dialog suppression)
- [x] GitHub Stats section combines live API auto-fetch with per-card manual overrides; falls back gracefully on rate limit / network error
- [x] CV / Resume tab — split out of Profile into its own sidebar tab
- [x] CV version history (`cv_files` table) — every upload/generate kept; admin picks which is active; active CV cannot be deleted
- [x] CV visibility toggle — hide the public Download CV button without deleting the file (`profiles.cv_visible`)
- [x] CV download counter (`profiles.cv_download_count`) — public `POST /profile/cv/download` increments, dashboard displays
- [x] In-dashboard PDF preview iframe of the currently active CV
- [x] CV auto-generator (basic) — `gofpdf` builds a clean classic resume from profile + skills + experience; lands in history but does NOT auto-activate (admin confirms via inline "Yes, set as active" banner). **Needs polish — see Known Issues.**
- [x] CV Compare modal — fullscreen side-by-side comparison of any two saved CVs; per-pane dropdown + chip strip for quick swapping; "Make Active" button per pane; backfill of legacy active-CV into history so Compare always sees it
- [x] Profile section UI overhaul — split layout, sticky identity card with two photo uploads + live preview of name/title/email/phone/location/availability, three icon-decorated form sections, three-card visual availability selector
- [x] Phone field uses `intl-tel-input` country code picker (BD default) — stored as full E.164 with leading `+`
- [x] Hero pill (`Say Hi from {nickname}, {title}`) editable from dashboard with `nickname` + `hero_subtitle` override fields
- [x] Hero heading (`I build and ship / beautiful apps to / App Store & Play Store.`) editable with three line inputs + comma-separated highlight words + advanced free-form override; live preview in the dashboard, "Default heading" reference card, and "↺ Restore default" button
- [x] Title field (`profile.title`) bound to sidebar designation, hero subtitle pill, and footer tagline (was hardcoded everywhere)
- [x] Copyright system — editable `copyright_text` with `{year}` / `{name}` placeholders, sidebar/footer per-place overrides, optional "Built with" footer line, 5 quick-fill presets, insert chips
- [x] Footer & Copyright moved to its own `📜 Footer` sidebar tab with dedicated `PUT /profile/footer` endpoint
- [x] Status options drive the portfolio — sidebar Hire Me button text/state, hero CTA visibility, contact-card visibility, contact-form type filter, and a status banner above the contact form all change per Open to Work / Available for Freelance / On Vacation
- [x] Contact form success message + toast tailored to availability (vacation no longer promises 24-hour reply)

### Database ✅
- [x] PostgreSQL 17 installed and running as Windows service `postgresql-x64-17`
- [x] Database `imtiaz_portfolio` created
- [x] All 10 tables created via AutoMigrate (added `cv_files`)
- [x] Seeded: 1 admin, 1 profile, 17 skills, 3 experience entries, 2 projects

---

## 15. Known Issues & Pending Fixes

### Content
- [ ] **Profile photo** — `frontend/assets/images/profile/placeholder.jpg` is a placeholder. Replace with real photo.
- [x] ~~**CV PDF**~~ — handled by the new CV / Resume tab (upload + version history + auto-generate).
- [ ] **App screenshots** — Projects have empty `screenshots` arrays. Add real app screenshots.
- [ ] **App Store / Play Store links** — Currently `https://apps.apple.com` / `https://play.google.com` (generic). Update to real app listings.
- [x] ~~**Social links**~~ — now driven by the dashboard's Social Links tab; icons hide automatically when a URL is blank.
- [x] ~~**WhatsApp number**~~ — now driven by the dashboard. Contact button hides when no number is saved.
- [x] ~~**GitHub username**~~ — now driven by the dashboard's GitHub Stats tab. The `imtiazchowdhury` placeholder in `github.js` only acts as a fallback when no username is configured.

### Dashboard
- [ ] **Add modals** — "Add Project", "Add Skill", "Add Experience" buttons show placeholder toast. Real forms need to be built. **← Highest priority gap remaining.**
- [ ] **Edit functionality** — Dashboard can only delete, not edit existing entries inline.
- [ ] **Image upload UI** — Upload endpoint exists but no UI in dashboard to use it (for project screenshots, company logos, etc.).
- [x] ~~**Profile photo upload**~~ — fully done via the Profile tab's identity card (sidebar + about photos with click-to-upload + Cropper.js modal).

### Design / UI
- [ ] **Tech stack content** — When backend is unreachable, fallback message shows. Could use static hardcoded data as fallback.
- [ ] **Mobile responsiveness** — Some sections need polish on very small screens (< 400px).
- [ ] **Experience timeline** has a duplicate `class` attribute on the timeline container div.
- [ ] **Testimonials section** — Table exists in DB, no portfolio section rendered yet.

### CV / Resume
- [ ] **Generate from data — needs significant work.** Current `gofpdf`-based output is a minimal classic layout (name, title, contact bar, summary, experience, skills). Pending:
  - Better typography / spacing / hierarchy — current layout is utilitarian, not designed
  - Optional dark/branded variant matching the portfolio
  - Section ordering / inclusion controls (skip GitHub stats, skip availability, etc.)
  - Inline preview before save (so admin doesn't have to generate-then-discard)
  - Per-section page-break control to avoid orphaned headings
  - Pull projects/testimonials into the resume too (currently only profile + skills + experience)
  - Photo + iconography support (gofpdf can embed images, just not wired up yet)

---

## 16. Roadmap

### Phase 2 — Content & Polish
- Add dashboard modals for creating/editing projects, skills, experience
- ~~Profile photo and CV upload from dashboard~~ ✅ Done
- Real app screenshots in phone mockup
- Testimonials section on portfolio

### Phase 3 — Features
- ~~Email notification / reply when contact form is submitted~~ ✅ Done (Gmail SMTP)
- ~~Dashboard-driven social links and WhatsApp~~ ✅ Done
- ~~Multi-account Gmail App Password management with audit log~~ ✅ Done
- ~~SMTP credential auto-verification against Gmail~~ ✅ Done
- ~~GitHub stats with live API + manual override~~ ✅ Done
- Project detail modal/page (clicking "View Details" on app slide)
- Dark/light theme toggle
- ~~CV version history with active picker, visibility toggle, download counter, in-dashboard preview~~ ✅ Done
- ~~CV side-by-side compare modal~~ ✅ Done
- ~~CV auto-generation from database content~~ ⚠️ Partially done (see Known Issues → CV / Resume — minimal `gofpdf` layout works end-to-end, but needs design polish, section controls, preview-before-save, and richer content)
- ~~Editable hero pill, hero heading, copyright with placeholders~~ ✅ Done
- ~~Status options drive portfolio buttons (Tier 1)~~ ✅ Done
- Status Tier 2 — `vacation_return_date` field + "Back on {date}" banner messaging (deferred from this session)
- About section photo layout (planned)

### Phase 4 — Production
- Deploy backend to Railway or Render
- Deploy database to Railway PostgreSQL
- Custom domain setup
- SSL/HTTPS configuration
- Change JWT_SECRET to a strong random value
- Change admin password from `Admin@1234`

---

## 17. Deployment Guide

### Backend deployment (Railway)

1. Push `backend/` to a GitHub repository
2. Create new project on Railway → Deploy from GitHub
3. Set all environment variables from `.env` (especially `DB_PASSWORD`, `JWT_SECRET`, `APP_ENV=production`)
4. Railway auto-detects Go and runs `go build -o server ./cmd/main.go && ./server`
5. PostgreSQL addon available directly in Railway

### Environment variables for production
```env
APP_ENV=production
SERVER_PORT=8080
DB_HOST=[railway-postgres-host]
DB_PORT=5432
DB_USER=[railway-postgres-user]
DB_PASSWORD=[strong-random-password]
DB_NAME=imtiaz_portfolio
DB_SSLMODE=require
JWT_SECRET=[64-char-random-string]
JWT_EXPIRY=24h
ALLOWED_ORIGINS=https://yourdomain.com
ADMIN_EMAIL=admin@imtiaz.dev
ADMIN_PASSWORD=[strong-password-change-this]
```

### What changes for production frontend
In `frontend/assets/js/core/api.js`, `BASE_URL` is `/api/v1` (relative). Since the Go server serves both frontend and API, no change needed — it works identically in production.

---

## 18. Changelog

| Version | Date | Changes | Files Affected |
|---------|------|---------|----------------|
| 1.0.0 | 2026-05-07 | Initial build — complete backend + frontend | All files |
| 1.1.0 | 2026-05-07 | Drake-style sidebar (floating card, exact CSS) | sidebar.css, index.html |
| 1.2.0 | 2026-05-07 | Hero section rebuilt (font-weight 300, 78px, rotating badge, 72px stats) | portfolio.css, index.html |
| 1.3.0 | 2026-05-07 | Go backend serves frontend (no Live Server needed) | routes.go, api.js |
| 1.4.0 | 2026-05-07 | Removed Vanta.js animation (plain dark background like Drake) | index.html, hero.js |
| 1.5.0 | 2026-05-07 | Removed alternating section background colors | portfolio.css, app-slider.css |
| 1.6.0 | 2026-05-07 | Scroll animations — IntersectionObserver + CSS transitions | base.css, ScrollAnimations.js, main.js, all section JS |
| 1.7.0 | 2026-05-07 | Scroll animations applied to ALL elements (hero, contact, footer, dynamic content) | index.html, architecture.js, github.js, techstack.js, experience.js |
| 1.8.0 | 2026-05-07 | Right-side navigation rebuilt with Phosphor Icons | navbar.css, index.html |
| 1.9.0 | 2026-05-07 | README updated for Mac/Windows dual-environment setup (DB credentials, paths, commands) | README.md, backend/.env |
| 2.0.0 | 2026-05-07 | Sidebar corner icon — rotating sparkle breaks border at top-left, Drake effect | sidebar.css, index.html |
| 2.1.0 | 2026-05-07 | Scroll animation flicker fix — once:true, GPU compositing, threshold 0.1 | ScrollAnimations.js, base.css |
| 2.2.0 | 2026-05-07 | Sidebar Hire Me button hides on hero, shows on scroll with card height animation | Sidebar.js, sidebar.css |
| 2.3.0 | 2026-05-07 | Dashboard login stale-token fix — clear token before login, proper 401 handling | auth.js, api.js |
| 2.4.0 | 2026-05-07 | WhatsApp-style inbox — split panel, conversation list, chat bubbles, reply compose | dashboard.html |
| 2.5.0 | 2026-05-07 | Gmail SMTP reply system — App Password stored in DB, send replies from dashboard | email_service.go, profile model |
| 2.6.0 | 2026-05-07 | Reply history — every reply stored in message_replies table, shown as chat bubbles | message_reply.go, dashboard.html |
| 2.7.0 | 2026-05-07 | File attachments in replies — WhatsApp image grid, doc cards, lightbox viewer | dashboard.html, message_handler.go |
| 2.8.0 | 2026-05-07 | CV upload from dashboard — PDF upload, Download CV buttons shown dynamically | profile_handler.go, dashboard.html, about.js |
| 2.9.0 | 2026-05-07 | Profile photo upload with Cropper.js — drag/zoom crop modal, live previews | dashboard.html, about.js |
| 3.0.0 | 2026-05-07 | Profile update safety — preserve existing fields, no-store cache on profile API | profile_service.go, profile_handler.go |
| 3.1.0 | 2026-05-07 | Separate sidebar photo and about section photo — two independent upload fields | profile.go, profile_handler.go, dashboard.html, about.js |
| 3.2.0 | 2026-05-07 | HEIC/HEIF support — auto-converts iPhone photos to JPEG before upload via heic2any | dashboard.html, upload_service.go |
| 3.3.0 | 2026-05-08 | Social Links sidebar tab — split GitHub/LinkedIn/Twitter/Instagram/WhatsApp out of Profile; new `PUT /profile/social` endpoint that allows clearing fields | dashboard.html, profile_handler.go, profile_service.go, routes.go |
| 3.4.0 | 2026-05-08 | WhatsApp country code picker via `intl-tel-input` — dark-themed flag dropdown, stores E.164 without leading `+` | dashboard.html |
| 3.5.0 | 2026-05-08 | Public portfolio social icons + WhatsApp button now driven by profile data via `data-social="…"` attributes; empty values hide the icon | index.html, about.js |
| 3.6.0 | 2026-05-08 | Mail Settings sidebar tab — split SMTP credentials out of Profile; new `PUT /profile/mail` endpoint | dashboard.html, profile_handler.go, routes.go |
| 3.7.0 | 2026-05-08 | App Password History audit log — `mail_password_histories` table, history panel with Show/Hide/Copy/Use this/Delete row | mail_password_history.go, profile_handler.go, profile_service.go, dashboard.html |
| 3.8.0 | 2026-05-08 | SMTP credential auto-verification — `POST /profile/mail/verify` runs Gmail AUTH PLAIN without sending mail; auto-tests on tab open and after every save | email_service.go, profile_handler.go, dashboard.html |
| 3.9.0 | 2026-05-08 | Available Accounts panel — deduped per-Gmail cards with parallel auto-test, ↺ Use this account, password as 4 boxes | dashboard.html |
| 3.10.0 | 2026-05-09 | Hide / Restore for Gmail accounts — `hidden_mail_accounts` table replaces destructive delete; audit log preserved | hidden_mail_account.go, profile_handler.go, profile_service.go, dashboard.html |
| 3.11.0 | 2026-05-09 | Click-twice confirmation pattern across destructive dashboard actions — robust against Firefox dialog suppression | dashboard.html |
| 3.12.0 | 2026-05-09 | GitHub Stats sidebar tab + dynamic public section — live `api.github.com` fetch (Repositories, Top Language, Years Active) with per-card manual overrides; Total Commits is manual-only | profile.go, profile_handler.go, profile_service.go, routes.go, dashboard.html, github.js |
| 3.13.0 | 2026-05-09 | README brought current — new endpoints, new tables, new dashboard tabs, Known Issues / Roadmap reconciled | README.md |
| 3.14.0 | 2026-05-09 | Brand icons added next to each Social Links input in the dashboard (GitHub, LinkedIn, Twitter/X, Instagram, WhatsApp) via Phosphor Icons | dashboard.html |
| 3.15.0 | 2026-05-09 | CV / Resume split out of Profile into its own dashboard sidebar tab (consistent with Social Links / Mail Settings / GitHub Stats split pattern) | dashboard.html |
| 3.16.0 | 2026-05-09 | CV version history — `cv_files` table + endpoints (history list, activate, delete). Every upload appends a row; active CV is mirrored onto `profiles.cv_file` | cv_file.go, profile_handler.go, profile_service.go, routes.go, main.go |
| 3.17.0 | 2026-05-09 | CV visibility toggle + download counter — `cv_visible` and `cv_download_count` columns on `profiles`; public `POST /profile/cv/download` increments counter; portfolio respects the toggle | profile.go, profile_handler.go, profile_service.go, routes.go, about.js |
| 3.18.0 | 2026-05-09 | CV auto-generator — `gofpdf`-based clean classic resume PDF built from profile + skills + experience, saved to `/uploads/cv/`, registered in history as `source="generated"` | cv_generator_service.go, go.mod |
| 3.19.0 | 2026-05-09 | Dashboard CV tab rebuilt — in-page PDF preview iframe, Generate button, visibility toggle, download counter, version history list with View / Make Active / Delete (click-twice confirm) | dashboard.html |
| 3.20.0 | 2026-05-09 | Generate doesn't auto-activate — generated CV lands in history but admin must confirm via inline "Yes, set as active" banner. Uploads still auto-activate. `AddCVHistory` gained an `activate bool` parameter | profile_service.go, profile_handler.go, cv_generator_service.go, dashboard.html |
| 3.21.0 | 2026-05-09 | CV tab full UI polish — hero card with brand-blue glow, iOS-style toggle switch, 3-up stats row (downloads / versions / last updated), action cards for Upload + Generate (Phosphor icons), preview frame with "Active" corner pill, history as hoverable cards with source-colored stripes and pulsing active glow | dashboard.html |
| 3.22.0 | 2026-05-09 | Preview redesigned — toolbar (filename + source/size pills + expand / open-in-tab / download icon buttons), 420px default height with Expand toggle to 820px. Version History redesigned — featured "Currently Active" card pinned at top + responsive grid of past-version tiles with stylized PDF-cover artwork, three-dot action menu (Make active / Download / Delete with click-twice confirm), animated popup, click-outside dismiss | dashboard.html |
| 3.23.0 | 2026-05-09 | CV tab split layout — Preview now occupies 80% horizontal width with Version History as a 20% scrollable side column (max-height 760px). History tiles in side mode reflow to single-column horizontal-cover layout; featured card stacks vertically; View link hidden in favor of the three-dot menu's Download. Falls back to stacked layout below 1100px viewport | dashboard.html |
| 3.24.0 | 2026-05-09 | Preview iframe auto-fits to full PDF size via PDF.js (3.11.174 from cdnjs) — measures page count + per-page aspect ratio, computes total fit-to-width height, sets iframe height so the entire CV displays without internal scrolling. Uses `view=FitH` URL fragment so the browser PDF viewer matches the calculation. Re-fits on window resize (debounced 250ms). Expand toggle removed (no longer needed) | dashboard.html |
| 3.25.0 | 2026-05-09 | Tightened CV preview height calculation — was overshooting the actual PDF size, leaving empty space below. Now subtracts ~8px horizontal browser padding from rendered page width, drops top/bottom padding (`toolbar=0` strips the chrome), and reduces inter-page gap to 8px | dashboard.html |
| 3.26.0 | 2026-05-09 | CV preview switched from `<iframe>` to PDF.js canvas rendering — each page rendered as a stacked `<canvas>`, container height equals the exact sum of canvas heights so there's zero whitespace below. Renders at 2× scale for crisp display on high-DPI screens; debounced resize re-render keeps it sharp after width changes. In-flight render token cancels stale renders when the active CV changes mid-render. "Open in new tab" still uses the native PDF viewer for full interaction | dashboard.html |
| 3.27.0 | 2026-05-09 | Removed the hero card's "View PDF" button — redundant with the inline preview right below it. Toolbar's Open-in-tab + Download icons remain | dashboard.html |
| 3.28.0 | 2026-05-09 | Preview button on every history card — click any past CV to load it into the preview without making it active. New "Previewing past version" amber pill appears in the toolbar when the previewed CV isn't the active one; the previewed tile gets a primary-colored outline ring. Featured card's old "View" link replaced with a "Preview" button (handy for jumping back to the active CV) | dashboard.html |
| 3.29.0 | 2026-05-09 | "↩ Back to active" button in the preview toolbar — appears in green next to the "Previewing past version" pill whenever a non-active CV is loaded; one click jumps the preview back to the currently active version | dashboard.html |
| 3.30.0 | 2026-05-09 | Back-to-active button reliability fixes — cache the active row in `_cvActiveRow` so the click never depends on history-render timing; reordered `loadCVTab` to render history before preview so highlight ring is correctly applied on first paint; moved the button's inline styles to a real `.cv-back-to-active-btn` class with hover state | dashboard.html |
| 3.31.0 | 2026-05-09 | CV side-by-side comparison — new `⫴ Compare` icon button in the History panel header opens a fullscreen modal with two scrollable PDF panes, each with its own dropdown to pick any saved CV. Defaults to active CV vs. next-newest. PDF.js renders each pane with independent cancellation tokens. Esc or backdrop click closes. Refuses if fewer than 2 CVs are saved | dashboard.html |
| 3.32.0 | 2026-05-09 | History backfill for legacy active CVs — `GetCVHistory` now inserts a synthetic `cv_files` row when `profile.cv_file` exists but isn't tracked in history (typical after pre-`cv_files` uploads). Fixes Compare/History invisible legacy CVs and ensures every active CV is delete-protected and addressable | profile_service.go |
| 3.33.0 | 2026-05-09 | Compare modal — bottom strip of all-CV chips with `← Left` / `Right →` buttons so any saved CV can be swapped into either pane in one click. Chips currently shown in either pane glow with a primary border and the matching button is disabled. Strip + panes stay in sync whether the user uses dropdowns or chip buttons | dashboard.html |
| 3.34.0 | 2026-05-09 | "Make Active" button on each compare pane — green button below the dropdown that promotes the pane's CV to active. Optimistic local update keeps the modal in sync without a re-fetch; background `loadCVTab()` refreshes the underlying dashboard. Disabled with "Already Active" label when the shown CV is already active | dashboard.html |
| 3.35.0 | 2026-05-09 | Profile section UI overhaul — split layout with sticky left identity card (live preview of name/title/email/phone/location/availability + click-to-upload sidebar & about photos with hover overlay), and right side organized into 3 panels (Basic Info / Status / About You) with Phosphor icons on every label. Availability is now a 3-card visual selector (Open/Freelance/Busy). Bio shows live character count. Added Phone field (already in DB model, was previously unedited) | dashboard.html |
| 3.36.0 | 2026-05-09 | Phone field now uses an `intl-tel-input` country code picker (same library as WhatsApp) — flag dropdown + separate dial code, defaults to BD with US/GB/IN/AE preferred. Stored as full E.164 with leading "+". Identity card preview shows the full international number, including after country switches | dashboard.html |
| 3.37.0 | 2026-05-09 | "Busy" availability replaced with "On Vacation" — palm-tree icon, amber badge color (more apt than the prior coffee/red). CSS class renamed from `is-busy` to `is-away` | dashboard.html |
| 3.38.0 | 2026-05-09 | Profile `title` field is now wired to the public portfolio in three places (was hardcoded everywhere): sidebar designation, hero subtitle pill ("Say Hi from Imtiaz, {title}"), and footer tagline ("{title} · {location}"). New IDs `sidebar-designation`, `hero-subtitle-title`, `footer-tagline` + `updateFooter()` helper | index.html, about.js |
| 3.39.0 | 2026-05-09 | Editable copyright with auto-fallback — new `copyright_text` field on `profiles`, new "Footer" panel in the dashboard's Profile tab. Sidebar + footer copyright lines now show the custom text when provided, otherwise auto-fill `© {currentYear} {full_name}. All Rights Reserved.` | profile.go, dashboard.html, index.html, about.js |
| 3.40.0 | 2026-05-09 | Footer panel customization — three new profile fields (`sidebar_copyright_text`, `footer_copyright_text`, `footer_built_with`); `{year}` / `{name}` placeholders that resolve on the public site; live preview block, 5 quick-fill presets (Standard / Minimal / Made with ❤️ / Crafted by / Reverse), insert chips for placeholders + ©/❤️/✨/💙/🚀/·, collapsible per-place override section. Footer now has an optional second "Built with…" line driven by the new field | profile.go, dashboard.html, index.html, about.js, portfolio.css |
| 3.41.0 | 2026-05-09 | Footer panel split into its own `📜 Footer` sidebar tab — new dedicated `PUT /profile/footer` endpoint with pointer-typed body (empty = clear, nil = preserve), four footer fields added to the regular profile preserve list so the main Profile save no longer touches them | profile_handler.go, profile_service.go, routes.go, dashboard.html, README.md |
| 3.42.0 | 2026-05-09 | Hero pill is now editable — two new profile fields: `nickname` (replaces hardcoded "Imtiaz" in the greeting) and `hero_subtitle` (full override of the pill text with `{name}`/`{title}`/`{nickname}` placeholders). New "Hero Pill" section in the Profile tab. Public portfolio resolves nickname → first word of full_name when blank | profile.go, dashboard.html, index.html, about.js |
| 3.43.0 | 2026-05-09 | Hero heading is now editable — five new profile fields (`hero_heading_line1/2/3`, `hero_heading_highlights`, `hero_heading_override`). Dashboard Profile tab gains a "Hero Heading" section with a live preview, three line inputs, a comma-separated highlight-words field, and a collapsible advanced override (newlines + `*asterisks*` for highlights). Override auto-opens when a value is saved. Same renderer used by the dashboard preview and the public portfolio so they stay perfectly in sync | profile.go, dashboard.html, index.html, about.js |
| 3.43.1 | 2026-05-09 | Fix hero-heading preview not showing the brand-blue highlight color — public CSS rule was scoped to `.hero-title .highlight`, which never matched in the dashboard preview wrapper. Added a dashboard-wide `.dash-section .highlight` rule mirroring the portfolio styling | dashboard.html |
| 3.43.2 | 2026-05-09 | Hero-heading renderer marker hardening — replaced `\x00`-based HOPEN/HCLOSE markers with ASCII-safe `__HL_OPEN_KP__` / `__HL_CLOSE_KP__` (some HTML parsers strip null bytes from string fragments before innerHTML assignment, leaving raw marker tokens visible and breaking highlight rendering). Same fix in `dashboard.html` and `about.js` | dashboard.html, about.js |
| 3.44.0 | 2026-05-10 | Hero Heading section UX polish — clearer per-field hints (with concrete examples + warning that Highlight Words only colours text already in the lines), a "Default heading" reference card showing the original styled output, and a "↺ Restore default" button in the section header that one-click fills the original values | dashboard.html |
| 3.45.0 | 2026-05-10 | Status options now drive the portfolio's hire-related buttons. New `applyStatusBehavior(p)` in `about.js`: **Open to Work** keeps everything visible with green "Open to opportunities" banner. **Available for Freelance** hides job-only buttons (sidebar relabels to "Start a Project", hero job CTA + Job contact card hidden, form type filtered to Freelance/Other), blue banner. **On Vacation** hides all hero CTAs + both contact cards, sidebar shows disabled amber "Currently Away", form type filtered to Other only, amber banner with vacation message; contact form stays accessible per requirement | index.html, portfolio.css, about.js |
| 3.45.1 | 2026-05-10 | Contact form success message + toast now tailored to availability — On Vacation no longer falsely promises "reply within 24 hours"; instead shows "🌴 Message received! I'm currently on vacation, but I'll reply when I'm back." Reads `Store.get('profile')` so it picks up whatever about.js loaded | contact.js |
| 3.45.2 | 2026-05-10 | Moved the contact form success div from above the Send button to below it — was appearing in an unnatural spot mid-form; now sits where users expect to see confirmation after clicking Send | index.html |
| 3.45.3 | 2026-05-10 | Removed hardcoded "I'll reply within 24 hours" fallback text from the contact-success div — the div is now empty in HTML and contact.js fills it with the status-aware message at submit time. Eliminates the apparent stale-text bug when a cached old contact.js was being served | index.html |
| 3.46.0 | 2026-05-10 | README session summary — new `§ 0 · Where We Left Off` section at the top with version, recent focus areas, and prioritized suggested next steps. Section 11 dashboard tabs list updated to reflect the new Profile / CV / Footer panel structure. Section 14 backend + frontend lists updated with the session's additions. Section 15 Known Issues + Section 16 Roadmap reconciled (e.g. Profile photo upload now ✅, Status Tier 2 added as deferred work) | README.md |

> **Rule:** Every future change must add a row to this table before the session ends.
