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
- **Auth**: Registro/Login con sesiones
- **Rutinas**: CRUD completas
- **Ejercicios**: Con peso, series, repeticiones
- **Tracking**: Alumnos marcan ejercicios completados
- **Progreso**: Historial de sesiones

## Inicio Rápido

### Backend

```bash
cd fitlab-api
go run .
# Server en http://localhost:8080
```

### Frontend

```bash
cd fitlab-web
npm install
npm run dev
# App en http://localhost:5173
```

### Production

```bash
# Backend
cd fitlab-api
go build -o fitlab-api
./fitlab-api

# Frontend
cd fitlab-web
npm run build
# Servir dist/ con nginx/caddy
```

## Variables de Entorno

| Variable                      | Default                                | Descripción                     |
|-------------------------------|---------------------------------------|--------------------------------|
| `ADDR`                        | `:8080`                                | Dirección del server            |
| `DATABASE_PATH`               | `./fitlab.db`                          | Path a SQLite                   |
| `SESSION_SECRET`              | *(auto-generado)*                       | Clave sesiones (se genera si no se setea) |
| `PROFESSOR_REGISTRATION_CODE` | *(requerido en prod)*               | Código para registar profesores |
| `ALLOWED_ORIGINS`            | `localhost:5173,localhost:8080`        | Origins permitidos para CORS    |
| `DEV_MODE`                  | `true`                                | `true`=dev, `false`=prod   |

### Desarrollo (DEV_MODE=true)
- CORS permite `localhost:5173` y `localhost:8080`
- Session secret se auto-genera cada inicio
- Código de profesor usa fallback `"PROF2024"`

### Producción (DEV_MODE=false)
- Configurar `ALLOWED_ORIGINS` con el dominio real
- `SESSION_SECRET` debe estar configurado
- `PROFESSOR_REGISTRATION_CODE` debe estar configurado

## Roles

- **Profesor**: Crea y asigna rutinas a alumnos
- **Alumno**: Ve sus rutinas, marca ejercicios completados

## API Endpoints

| Método | Path | Auth | Desc |
|--------|------|------|------|
| POST | `/api/auth/register` | - | Registro |
| POST | `/api/auth/login` | - | Login |
| POST | `/api/auth/logout` | ✓ | Logout |
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
- mattn/go-sqlite3 (database)
- bcrypt (passwords)

### Frontend
- Vue 3 (Composition API)
- Vue Router 4
- Pinia 2 (state)
- Vite 5

---

💪 Entrená en FITLAB