# Imtiaz Portfolio — Flutter Developer Portfolio

Personal portfolio website and admin dashboard for **Chowdhury Md. Imtiazul Islam**, Flutter Developer.

## Tech Stack

| Layer    | Technology                              |
|----------|-----------------------------------------|
| Frontend | HTML5, CSS3, Vanilla JavaScript         |
| Backend  | Golang + Gin framework                  |
| ORM      | GORM                                    |
| Database | PostgreSQL                              |
| Auth     | JWT (golang-jwt/jwt)                    |
| Animation| Vanta.js (Three.js), CSS animations     |
| Slider   | Swiper.js                               |

## Architecture

```
imtiaz-portfolio/
├── frontend/          # Pure HTML/CSS/JS — no framework
│   ├── public/        # HTML pages (index, dashboard, login)
│   └── assets/        # CSS, JS, images
│       ├── css/       # Component-based CSS
│       └── js/        # Component-based JS (core + components + pages)
├── backend/           # Go server
│   ├── cmd/           # Entry point (main.go)
│   ├── config/        # DB + app config
│   ├── internal/      # Business logic
│   │   ├── models/    # GORM database models
│   │   ├── services/  # Business logic
│   │   ├── handlers/  # HTTP handlers
│   │   ├── middleware/# JWT auth, CORS, logging
│   │   └── routes/    # All route definitions
│   ├── migrations/    # Reference SQL files
│   └── seeds/         # Default data seeder
├── scripts/           # Setup, run, migrate scripts
└── docs/              # API + DB documentation
```

## Getting Started

### Prerequisites
- Go 1.21+
- PostgreSQL 14+
- A modern web browser

### Setup

**1. Clone and configure**
```bash
cd backend
cp .env.example .env
# Edit .env — set your PostgreSQL DB_PASSWORD
```

**2. Create the database**
```sql
CREATE DATABASE imtiaz_portfolio;
```

**3. Start the backend**
```bash
# Windows:
cd backend
go run ./cmd/main.go

# Linux/Mac:
bash scripts/run.sh
```

The server will:
- Auto-migrate all database tables
- Seed default admin + portfolio content
- Start listening on http://localhost:8080

**4. Open the portfolio**

Open `frontend/public/index.html` in a browser, or serve with VS Code Live Server.

**5. Admin dashboard**

Open `frontend/public/login.html`

Default credentials:
- Email: `admin@imtiaz.dev`
- Password: `Admin@1234`

> Change these in `.env` before deploying to production!

## API Documentation

Base URL: `http://localhost:8080/api/v1`

All responses follow this shape:
```json
{
  "success": true,
  "message": "Projects fetched successfully",
  "data": [...],
  "error": null
}
```

### Public Endpoints (no auth)

| Method | Endpoint           | Description            |
|--------|--------------------|------------------------|
| POST   | /auth/login        | Admin login → JWT      |
| GET    | /projects          | Get all projects       |
| GET    | /projects/:id      | Get one project        |
| GET    | /skills            | Get all skills         |
| GET    | /experience        | Get all experience     |
| GET    | /profile           | Get portfolio profile  |
| POST   | /messages          | Submit contact form    |

### Protected Endpoints (Bearer token required)

| Method | Endpoint           | Description                    |
|--------|--------------------|--------------------------------|
| GET    | /auth/me           | Current user info              |
| POST   | /projects          | Create project                 |
| PUT    | /projects/:id      | Update project                 |
| DELETE | /projects/:id      | Delete project                 |
| POST   | /skills            | Add skill                      |
| PUT    | /skills/:id        | Update skill                   |
| DELETE | /skills/:id        | Delete skill                   |
| POST   | /experience        | Add experience                 |
| PUT    | /experience/:id    | Update experience              |
| DELETE | /experience/:id    | Delete experience              |
| GET    | /messages          | Get all messages               |
| PUT    | /messages/:id      | Mark message as read           |
| DELETE | /messages/:id      | Delete message                 |
| GET    | /profile           | (public too)                   |
| PUT    | /profile           | Update profile                 |
| POST   | /upload            | Upload file (image/PDF)        |
| GET    | /stats             | Dashboard statistics           |

## Database Schema

| Table        | Purpose                                    |
|--------------|--------------------------------------------|
| users        | Admin account (1 row)                      |
| profiles     | Portfolio content (1 row)                  |
| projects     | App showcase entries                       |
| skills       | Tech stack with percentages                |
| experiences  | Work history timeline                      |
| testimonials | Client reviews (optional)                  |
| messages     | Contact form submissions                   |

## Frontend Pages

| File              | Purpose                                   |
|-------------------|-------------------------------------------|
| public/index.html | Portfolio — public facing                 |
| public/login.html | Admin login page                          |
| public/dashboard.html | Admin dashboard — manage content      |

## Deployment

### Backend (Render / Railway)
1. Push backend/ to a GitHub repo
2. Set environment variables from .env.example
3. Build command: `go build -o server ./cmd/main.go`
4. Start command: `./server`

### Frontend (Netlify / Vercel / GitHub Pages)
1. Update `API_BASE_URL` in `assets/js/core/api.js` to your deployed backend URL
2. Deploy the `frontend/` folder

## Changelog

| Version | Date       | Changes                          | Files                    | Reason          |
|---------|------------|----------------------------------|--------------------------|-----------------|
| 1.0.0   | 2026-05-07 | Initial creation — full project  | All files                | Initial build   |
