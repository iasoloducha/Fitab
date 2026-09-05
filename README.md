# FITLAB Gym Castelar

Sistema completo de gestión y seguimiento para el gimnasio FITLAB en Castelar, Buenos Aires. Permite a profesores crear y asignar rutinas, a alumnos marcar sus ejercicios completados, y a administradores gestionar usuarios — todo en una plataforma web moderna.

## Arquitectura

```
┌─────────────────┐     ┌──────────────┐     ┌──────────────┐
│   Landing       │     │  fitlab-web  │     │  fitlab-api  │
│   (Vanilla)     │     │  Vue SPA     │     │  Go + Gin    │
│   institucional │     │  (Vercel)    │     │  (Fly.io)    │
└─────────────────┘     └──────────────┘     └──────┬───────┘
                                                    │ SQLite
                                            ┌───────┴────────┐
                                            │   fitlab.db    │
                                            │  (modernc/sqlite)│
                                            └────────────────┘
```

Tres componentes independientes, desplegables por separado:

- **Landing** — página institucional del gimnasio con generador de rutinas
- **`fitlab-web`** — aplicación Vue 3 (SPA) para alumnos, profesores y admins
- **`fitlab-api`** — API REST en Go + Gin, con SQLite empotrado (sin CGO)

## Estructura del repo

```
fitlab/
├── index.html           # Landing institucional
├── styles.css           # Estilos de la landing
├── scripts.js           # JS de la landing (generador de rutinas)
│
├── fitlab-web/          # App Vue 3 (SPA)
│   ├── src/
│   │   ├── views/       # Login, Register, Rutinas, Progreso, Admin, Catálogo, Perfil...
│   │   ├── stores/      # Pinia (auth, routines, catalog, admin)
│   │   ├── router/      # Vue Router
│   │   └── api/         # Cliente API (index.js)
│   ├── dist/            # Build de producción
│   └── vercel.json      # Config de deploy (rewrites SPA)
│
├── fitlab-api/          # Backend Go
│   ├── main.go          # Entry point + enrutado
│   ├── cmd/scheduler/   # Binario del scheduler de notificaciones
│   ├── internal/
│   │   ├── handlers/    # Auth, Routine, Exercise, Catalog, Admin, Email
│   │   ├── middleware/  # Auth, AdminOnly, ProfessorOnly
│   │   ├── models/      # Modelos de datos
│   │   ├── services/    # Notifier (emails de recordatorio/completación)
│   │   └── database/    # Inicialización SQLite + migraciones
│   ├── migrations/      # SQL DDL (001_initial.sql)
│   ├── .env.example     # Plantilla de variables de entorno
│   ├── fly.toml         # Config de Fly.io (API)
│   └── fly-scheduler.toml  # Config de Fly.io (scheduler)
│
└── README.md
```

## Funcionalidades

### Landing Page
- Información del gimnasio
- Generador de rutinas (exporta a print/PDF)
- Planes de membresía
- Contacto y ubicación

### App Web
- **Auth**: Registro / Login con sesiones, recuperación de contraseña por email, cambio de contraseña
- **Admin Dashboard**: Login especial + gestión de usuarios (lista, editar, eliminar) y backup/restore de la base de datos
- **Rutinas**: CRUD completo, copiar rutinas existentes
- **Ejercicios**: Con peso, series y repeticiones
- **Catálogo de ejercicios**: Lista reutilizable gestionada por profesores
- **Tracking**: Alumnos marcan ejercicios completados con peso real
- **Progreso**: Historial de sesiones y estadísticas
- **Notificaciones**: Emails automáticos de recordatorio y resumen de rutina completada (scheduler)

## Roles

| Rol | Permisos |
|-----|----------|
| **Admin** | Gestión de usuarios vía `ADMIN_EMAILS`, backup/restore de la base |
| **Profesor** | Crea, edita y asigna rutinas; gestiona el catálogo de ejercicios |
| **Alumno** | Ve sus rutinas, marca ejercicios completados, sigue su progreso |

## API Endpoints

### Auth

| Método | Path | Auth | Descripción |
|--------|------|------|-------------|
| POST | `/api/auth/register` | - | Registro (profesores requieren código) |
| POST | `/api/auth/login` | - | Login |
| POST | `/api/auth/logout` | ✓ | Cerrar sesión |
| POST | `/api/auth/forgot-password` | - | Recuperar contraseña (envía email) |
| GET  | `/api/auth/me` | ✓ | Usuario actual |
| PUT  | `/api/auth/password` | ✓ | Cambiar contraseña |
| GET  | `/api/users/students` | Prof | Listar alumnos |

### Rutinas

| Método | Path | Auth | Descripción |
|--------|------|------|-------------|
| GET  | `/api/routines` | ✓ | Listar rutinas |
| GET  | `/api/routines/:id` | ✓ | Ver rutina |
| POST | `/api/routines` | Prof | Crear rutina |
| PUT  | `/api/routines/:id` | Prof | Actualizar rutina |
| DELETE | `/api/routines/:id` | Prof | Eliminar rutina |
| POST | `/api/routines/:id/copy` | Prof | Duplicar rutina |
| POST | `/api/routines/:id/exercises` | Prof | Agregar ejercicio |
| PUT  | `/api/exercises/:id` | Prof | Actualizar ejercicio |
| DELETE | `/api/exercises/:id` | Prof | Eliminar ejercicio |

### Catálogo de ejercicios

| Método | Path | Auth | Descripción |
|--------|------|------|-------------|
| GET  | `/api/catalog/exercises` | Prof | Listar catálogo |
| POST | `/api/catalog/exercises` | Prof | Crear ejercicio de catálogo |
| PUT  | `/api/catalog/exercises/:id` | Prof | Actualizar ejercicio |
| DELETE | `/api/catalog/exercises/:id` | Prof | Eliminar ejercicio |

### Tracking y progreso

| Método | Path | Auth | Descripción |
|--------|------|------|-------------|
| POST | `/api/exercises/:id/logs` | ✓ | Marcar ejercicio completado |
| GET  | `/api/exercises/:id/logs` | ✓ | Ver logs del ejercicio |
| DELETE | `/api/logs/:id` | ✓ | Eliminar un log |
| GET  | `/api/progress` | ✓ | Estadísticas de progreso |

### Admin

| Método | Path | Auth | Descripción |
|--------|------|------|-------------|
| POST | `/api/admin/login` | - | Login especial de admin |
| GET  | `/api/admin/users` | Admin | Listar usuarios |
| PUT  | `/api/admin/users/:id` | Admin | Renombrar usuario |
| DELETE | `/api/admin/users/:id` | Admin | Eliminar usuario |
| GET  | `/api/admin/backup` | Admin | Descargar copia de la base |
| POST | `/api/admin/restore` | Admin | Restaurar base de datos |

### Otros

| Método | Path | Auth | Descripción |
|--------|------|------|-------------|
| GET  | `/health` | - | Health check |

## Tech Stack

### Backend (`fitlab-api`)
- **Go** 1.21+
- **Gin** — web framework
- **gorilla/sessions** — autenticación con cookies
- **modernc/sqlite** — base de datos embebida según CGO
- **bcrypt** — hash de contraseñas
- **SMTP** — envío de emails (recuperación + notificaciones)

### Frontend (`fitlab-web`)
- **Vue 3** (Composition API)
- **Vue Router 4**
- **Pinia** — state management
- **Vite 5**
- **Chart.js / vue-chartjs** — gráficos de progreso

### Landing
- HTML5 + CSS3 + Vanilla JS

## Quick Start

### 1. Backend

```bash
cd fitlab-api
cp .env.example .env   # y editá tus valores

# dev (usa ADDR/DEV_MODE del .env o defaults)
go run .

# en modo dev la API levanta en http://localhost:8080
```

### 2. Frontend

```bash
cd fitlab-web
npm install
npm run dev   # servidor de Vite en http://localhost:5173
```

### 3. Landing

No requiere build. Abrí `index.html` directamente en el navegador, o servila con cualquier servidor estático.

## Configuración

Todas las variables de entorno del backend están documentadas en `fitlab-api/.env.example`:

| Variable | Requerida | Descripción |
|----------|-----------|-------------|
| `SESSION_SECRET` | Producción | Clave para cifrar sesiones (hex de 32 bytes) |
| `ALLOWED_ORIGINS` | Producción | Orígenes CORS permitidos (separados por coma) |
| `ADMIN_EMAILS` | Para admin | Emails con acceso al dashboard de admin |
| `PROFESSOR_REGISTRATION_CODE` | Producción | Código para registrar profesores |
| `DATABASE_PATH` | No | Ruta de la base SQLite (default `./fitlab.db`) |
| `ADDR` / `PORT` | No | Dirección del servidor (default `:8080`) |
| `DEV_MODE` | No | `true` para desarrollo (cookies no secure, código de prof hardcodeado) |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM` | Para emails | Configuración SMTP (recuperación + notificaciones) |

> **Nota**: en `DEV_MODE=true`, el código de registro de profesor es `PROF2024` (solo local). En producción debe configurarse `PROFESSOR_REGISTRATION_CODE`.

## Deployment

### Backend (Fly.io)

```bash
cd fitlab-api
fly launch
fly secrets set SESSION_SECRET=$(openssl rand -hex 32)
fly secrets set ADMIN_EMAILS=admin@tu-gimnasio.com
fly secrets set SMTP_HOST=smtp.gmail.com SMTP_PORT=587 SMTP_USER=tu@email.com SMTP_PASS=app-password SMTP_FROM="Fitlab <tu@email.com>"
fly secrets set ALLOWED_ORIGINS=https://tu-dominio.vercel.app
fly deploy
```

### Notificaciones (scheduler independiente)

El proyecto incluye un scheduler separado que envía emails de recordatorio de rutinas por vencer (3 días antes) y resúmenes de rutinas completadas. Se despliega como una app Fly.io aparte:

```bash
cd fitlab-api
# fly-scheduler.toml ya viene preparado; seguí los pasos que están en su header
fly launch --config fly-scheduler.toml --no-deploy
fly secrets import < .env.scheduler
fly deploy --config fly-scheduler.toml
```

### Frontend (Vercel)

```bash
cd fitlab-web
npm run build
vercel --prod
```

> `vercel.json` ya está configurado con rewrites SPA, así que las rutas del router (ej. `/progress`, `/admin`) funcionan directas.

## Scheduler de notificaciones

El scheduler (`RUN_MODE=scheduler` o el binario `cmd/scheduler`) corre en loop y ejecuta dos tareas:

1. **Recordatorios de vencimiento**: avisa al alumno (y al profesor) 3 días antes de que su rutina expire.
2. **Resumen de completación**: cuando una rutina termina hoy, envía un resumen con las estadísticas de progreso.

Configurable vía `CHECK_INTERVAL` (default `5m`).

## Seguridad

- Contraseñas con hash bcrypt
- Sesiones HttpOnly con cookies firmadas
- CORS con whitelist de orígenes
- Código de profesor requerido para registros de rol profesor
- `SameSite` dinámico según entorno (Lax en dev, None en producción)
- Cookies `Secure` en producción (HTTPS)
- Recuperación de contraseña no revela si un email existe

---

💪 Entrená en FITLAB
