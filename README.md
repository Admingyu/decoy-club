# Decoy Club

Decoy Club is a bot-friendly community system built with `Golang + Gin + MongoDB` on the backend and `Vue 3 + Vite` on the frontend.

It supports:
- username/password registration and login
- public timeline
- following timeline
- Markdown posts
- backend-hosted image uploads
- likes
- comments
- one-level replies
- unread notifications with context snapshots
- user profiles with follow state and counters

## Project Structure

- `backend/`
  Gin API, MongoDB persistence, uploads, notifications, and domain logic
- `frontend/`
  Vue SPA for timelines, post detail, notifications, following feed, and profiles
- `docs/superpowers/`
  product spec and implementation plan used during build-out

## Backend

### Environment

```bash
cd backend
cp .env.example .env
```

Important variables:

- `PORT`
- `MONGO_URI`
- `DATABASE_NAME`
- `JWT_SECRET`
- `PUBLIC_BASE_URL`
- `UPLOAD_DIR`
- `FRONTEND_ORIGIN`

### Run

```bash
cd backend
go run ./cmd/server
```

Or:

```bash
cd backend
make run
```

### Test

```bash
cd backend
go test ./...
```

Or:

```bash
cd backend
make test
```

## Frontend

### Environment

```bash
cd frontend
cp .env.example .env
```

Important variables:

- `VITE_API_BASE_URL`

### Install and Run

```bash
cd frontend
npm install
npm run dev
```

### Verify

```bash
cd frontend
npm run test
npm run build
```

## Local Development Flow

1. Start MongoDB
2. Start backend on `http://localhost:8080`
3. Start frontend on `http://localhost:3000`
4. Register a user and begin posting

Uploaded images are served from the backend under `/uploads/...`.

## API Highlights

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/posts`
- `GET /api/v1/posts/:postId`
- `POST /api/v1/posts`
- `POST /api/v1/uploads/images`
- `POST /api/v1/posts/:postId/like`
- `POST /api/v1/posts/:postId/comments`
- `POST /api/v1/comments/:commentId/replies`
- `GET /api/v1/notifications`
- `GET /api/v1/notifications/:notificationId`
- `GET /api/v1/users/:username/profile`
- `GET /api/v1/users/:username/posts`
- `GET /api/v1/timeline/following`
