# FITLAB Gym Castelar

Sistema completo del gimnasio FITLAB en Castelar, Buenos Aires.

## Stack

- **Backend**: Go + Gin + SQLite (puro, sin CGO)
- **Frontend App**: Vue 3 + Vite + Pinia
- **Landing**: HTML5 + CSS3 + Vanilla JS

## Estructura

```
fitlab/
├── index.html              # Landing page institucional
├── styles.css            # Estilos landing
├── scripts.js          # JS landing (generador de rutinas)
├── fitlab-web/         # App Vue (SPA)
│   ├── src/
│   │   ├── views/      # Login, Register, Rutinas, Progreso
│   │   ├── stores/     # Pinia (auth, routines)
│   │   ├── router/     # Vue Router
│   │   └── api/        # Cliente API
│   └── dist/          # Build production
├── fitlab-api/        # Backend Go
│   ├── main.go
│   └── internal/
│       ├── handlers/   # Auth, Routine, Exercise
│       ├── middleware/  # Auth middleware
│       ├── models/      # Data models
│       └── database/    # SQLite
└── README.md
```

## Funcionalidades

### Landing Page
- Información del gimnasio
- Generador de rutinas (exporta a print/PDF)
- Planes de membresía
- Contacto y ubicación

### App (Backend)
- **Auth**: Registro/Login con sesiones, recuperación de contraseña
- **Admin Dashboard**: Login especial para administración
- **Rutinas**: CRUD completas
- **Ejercicios**: Con peso, series, repeticiones
- **Tracking**: Alumnos marcan ejercicios completados
- **Progreso**: Historial de sesiones

## Roles

- **Admin**: Gestión de usuarios (configurado via `ADMIN_EMAILS`)
- **Profesor**: Crea y asigna rutinas a alumnos
- **Alumno**: Ve sus rutinas, marca ejercicios completados

## API Endpoints

| Método | Path | Auth | Desc |
|--------|------|------|------|
| POST | `/api/auth/register` | - | Registro |
| POST | `/api/auth/login` | - | Login |
| POST | `/api/auth/logout` | ✓ | Logout |
| POST | `/api/auth/forgot-password` | - | Recuperar contraseña |
| GET  | `/api/auth/me` | ✓ | Usuario actual |
| GET  | `/api/routines` | ✓ | Listar rutinas |
| POST | `/api/routines` | Prof | Crear rutina |
| GET  | `/api/routines/:id` | ✓ | Ver rutina |
| POST | `/api/routines/:id/exercises` | Prof | Agregar ejercicio |
| POST | `/api/exercises/:id/logs` | ✓ | Marcar completado |
| GET  | `/api/progress` | ✓ | Ver progreso |

## Tech Stack Detallado

### Backend
- Go 1.21+
- Gin (web framework)
- gorilla/sessions (auth)
- modernc/sqlite (database sin CGO)
- bcrypt (passwords)

### Frontend
- Vue 3 (Composition API)
- Vue Router 4
- Pinia 2 (state)
- Vite 5

## Deployment

### Backend (Fly.io)
```bash
cd fitlab-api
fly deploy
fly secrets set SMTP_HOST=smtp.gmail.com SMTP_PORT=587 SMTP_USER=tu@email.com SMTP_PASS=app-password SMTP_FROM="Fitlab <tu@email.com>"
```

### Frontend (Vercel)
```bash
cd fitlab-web
vercel --prod
```

---

💪 Entrená en FITLAB