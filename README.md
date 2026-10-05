# REST-API

A RESTful web server written in Go that serves, receives, and stores data — designed to act as the backend for a mobile app or website. It provides user authentication, full CRUD for events, and event registration/cancellation.

---

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Authentication](#authentication)
- [Data Models & Schema](#data-models--schema)
- [API Reference](#api-reference)
- [Testing the API](#testing-the-api)
- [Error Handling](#error-handling)
- [Known Limitations & Security Notes](#known-limitations--security-notes)

---

## Features

- **User accounts** — sign up and log in with email/password.
- **Password hashing** — passwords stored as bcrypt hashes (cost 14), never in plain text.
- **JWT authentication** — stateless auth using signed JSON Web Tokens with a 2-hour expiry.
- **Events CRUD** — create, read, update, and delete events.
- **Ownership protection** — only the user who created an event may update or delete it.
- **Event registration** — authenticated users can register for events and cancel registrations.
- **Auto-migrating schema** — tables are created on startup if they do not exist.
- **SQLite storage** — zero-config, file-based database (`api.db`).

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.26 |
| Web framework | [Gin](https://github.com/gin-gonic/gin) v1.12 |
| Database | SQLite via [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) |
| Auth | [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) (HS256) |
| Password hashing | [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) |

---

## Project Structure

```
REST-API/
├── main.go                     # Entry point: initializes DB, sets up Gin, starts server
├── go.mod / go.sum             # Module definition and dependency lockfile
├── api.db                      # SQLite database file (created at runtime)
│
├── db/
│   └── db.go                   # DB connection, connection pool, table migrations
│
├── models/
│   ├── user.go                 # User struct, Save(), ValidateCredentials()
│   └── event.go                # Event struct, CRUD, Register()/CancelRegistration()
│
├── routes/
│   ├── routes.go               # Route registration
│   ├── users.go                # signup / login handlers
│   ├── events.go               # Event CRUD handlers
│   └── register.go             # Event registration handlers
│
├── middlewares/
│   └── auth.go                 # JWT authentication middleware
│
├── utils/
│   ├── hash.go                 # bcrypt hashing and verification
│   └── jwt.go                  # Token generation and verification
│
└── api-test/                   # .http request files for manual testing
```

---

## Prerequisites

- **Go 1.26+** — [download](https://go.dev/dl/)
- **CGO-enabled toolchain** — `mattn/go-sqlite3` is a CGO package, so a C compiler (`gcc`/`clang`) must be installed and `CGO_ENABLED=1` (the default on most systems).

Verify your setup:

```bash
go version
gcc --version
```

---

## Getting Started

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd REST-API
   ```

2. **Download dependencies**

   ```bash
   go mod download
   ```

3. **Run the server**

   ```bash
   go run .
   ```

   The server starts on **`http://localhost:8080`** and creates `api.db` with the required tables on first run.

4. **(Optional) Build a binary**

   ```bash
   go build -o rest-api .
   ./rest-api
   ```

---

## Configuration

Currently the app is configured directly in code with sensible defaults:

| Setting | Value | Location |
|---------|-------|----------|
| Server port | `8080` | `main.go` |
| Database file | `api.db` (working directory) | `db/db.go` |
| Max open connections | 10 | `db/db.go` |
| Max idle connections | 5 | `db/db.go` |
| JWT secret | `"supersecret"` | `utils/jwt.go` |
| Token expiry | 2 hours | `utils/jwt.go` |
| bcrypt cost | 14 | `utils/hash.go` |

> ⚠️ The JWT secret is currently hardcoded. See [Known Limitations](#known-limitations--security-notes) before deploying.

---

## Authentication

The API uses **JWT bearer-style tokens** for protected routes.

### Flow

1. **Register** via `POST /signup`.
2. **Log in** via `POST /login` to receive a token.
3. **Send the token** in the `Authorization` header on protected requests.

> Note: the middleware reads the raw `Authorization` header value. Send the **token itself** (no `Bearer ` prefix):
>
> ```
> Authorization: <your-jwt-token>
> ```

Tokens are HMAC-SHA256 signed and embed the user's `email` and `userId` along with an `exp` claim. The middleware verifies the signature and expiry, then makes the user ID available to handlers via `context.Set("userId", ...)`.

**Protected routes** live under an authenticated route group; protected handlers can assume a valid `userId`.

---

## Data Models & Schema

### `users`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | INTEGER | Primary key, auto-increment |
| `email` | TEXT | Not null, **unique** |
| `password` | TEXT | Not null (bcrypt hash) |

### `events`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | INTEGER | Primary key, auto-increment |
| `name` | TEXT | Not null |
| `description` | TEXT | Not null |
| `location` | TEXT | Not null |
| `dateTime` | DATETIME | Not null |
| `user_id` | INTEGER | Foreign key → `users(id)` (event owner) |

### `registrations`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | INTEGER | Primary key, auto-increment |
| `event_id` | INTEGER | Foreign key → `events(id)` |
| `user_id` | INTEGER | Foreign key → `users(id)` |

---

## API Reference

Base URL: `http://localhost:8080`

**Status codes used:** `200 OK`, `201 Created`, `400 Bad Request`, `401 Unauthorized`, `409 Conflict`, `500 Internal Server Error`.

### Public Endpoints

#### `POST /signup` — Create a new user

**Request body**

```json
{
  "email": "test@example.com",
  "password": "test123"
}
```

**Responses**

| Status | Body |
|--------|------|
| `201` | `{ "message": "User created successfully!" }` |
| `400` | `{ "message": "Could not parse request data." }` |
| `409` | `{ "message": "Email is already registered." }` |
| `500` | `{ "message": "Could not save user." }` |

---

#### `POST /login` — Authenticate and receive a token

**Request body**

```json
{
  "email": "test@example.com",
  "password": "test123"
}
```

**Responses**

| Status | Body |
|--------|------|
| `200` | `{ "message": "Login successful!", "token": "<jwt>" }` |
| `400` | `{ "message": "Could not parse request data." }` |
| `401` | `{ "message": "Could not authenticate user." }` |
| `500` | `{ "message": "Could not process login." }` |

---

#### `GET /events` — List all events

**Response `200`** — array of event objects:

```json
[
  {
    "ID": 1,
    "Name": "Another test event",
    "Description": "An event test",
    "Location": "A test location",
    "DateTime": "2026-09-30T09:55:00Z",
    "UserID": 1
  }
]
```

---

#### `GET /events/:id` — Get a single event

| Status | Body |
|--------|------|
| `200` | Event object |
| `400` | `{ "message": "Could not parse event id." }` |
| `500` | `{ "message": "Could not fetch event." }` |

---

### Protected Endpoints

All routes below require an `Authorization` header containing a valid JWT.

#### `POST /events` — Create an event

The authenticated user becomes the event owner (`user_id` is set server-side).

**Request body**

```json
{
  "name": "A test event",
  "description": "Test event!!",
  "location": "A test location",
  "dateTime": "2026-09-30T09:55:00.000Z"
}
```

**Responses**

| Status | Body |
|--------|------|
| `201` | `{ "message": "Event created!", "event": { ... } }` |
| `400` | `{ "message": "Could not parse request data." }` |
| `500` | `{ "message": "Could not create event. Try again later." }` |

---

#### `PUT /events/:id` — Update an event

Only the event's owner may update it.

**Request body** — same shape as create:

```json
{
  "name": "updated test event",
  "description": "A test event",
  "location": "Test location (Updated!)",
  "dateTime": "2026-09-30T15:15:00Z"
}
```

**Responses**

| Status | Body |
|--------|------|
| `200` | `{ "message": "Events updated successfully!" }` |
| `400` | `{ "message": "Could not parse request data." }` |
| `401` | `{ "message": "Not authorized to update event" }` |
| `500` | `{ "message": "Could not update event." }` |

---

#### `DELETE /events/:id` — Delete an event

Only the event's owner may delete it.

| Status | Body |
|--------|------|
| `200` | `{ "message": "Event deleted successfully!" }` |
| `400` | `{ "message": "Could not parse event id." }` |
| `401` | `{ "message": "Not authorized to delete event" }` |
| `500` | `{ "message": "Could not delete the event." }` |

---

#### `POST /events/:id/register` — Register for an event

No request body. Registers the authenticated user for the given event.

| Status | Body |
|--------|------|
| `201` | `{ "message": "Registered!" }` |
| `400` | `{ "message": "Could not parse event id." }` |
| `500` | `{ "message": "Could not fetch event." }` / `{ "message": "Could not register user." }` |

---

#### `DELETE /events/:id/register` — Cancel a registration

No request body. Removes the authenticated user's registration for the given event.

| Status | Body |
|--------|------|
| `200` | `{ "message": "Registration Cancelled!" }` |
| `500` | `{ "message": "Could not cancel registration." }` |

---

## Testing the API

The `api-test/` directory contains [`.http`](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) request files for use with the VS Code **REST Client** extension (or any `.http`-compatible client such as JetBrains HTTP Client).

| File | Request |
|------|---------|
| `create-user.http` | `POST /signup` |
| `login.http` | `POST /login` |
| `create-event.http` | `POST /events` |
| `get-events.http` | `GET /events` |
| `get-single-event.http` | `GET /events/:id` |
| `update-event.http` | `PUT /events/:id` |
| `delete-event.http` | `DELETE /events/:id` |
| `register.http` | `POST /events/:id/register` |
| `cancel-registration.http` | `DELETE /events/:id/register` |

### Suggested workflow

1. Run `go run .`.
2. Execute `create-user.http` to create an account.
3. Execute `login.http` and copy the `token` from the response.
4. Paste the token into the `Authorization` header of the protected `.http` files.
5. Run the event and registration requests.

> Tokens expire after 2 hours — if a request returns `401 Not Authorized`, log in again for a fresh token.

---

## Error Handling

- Handlers return JSON bodies of the form `{ "message": "..." }` with an appropriate HTTP status code.
- Invalid or missing JSON bodies yield `400 Bad Request`.
- Missing/invalid/expired tokens yield `401 Not Authorized` via the auth middleware.
- Authorization failures (editing someone else's event) yield `401` with a descriptive message.
- Database and other server-side failures yield `500 Internal Server Error` and never leak internal details.

---

## Known Limitations & Security Notes

This project is a solid learning/reference implementation, but review the following before any production use:

- **Hardcoded JWT secret** — `secretKey = "supersecret"` in `utils/jwt.go` should be loaded from an environment variable or secrets manager.
- **No token revocation** — JWTs remain valid until expiry; there is no logout/blacklist mechanism.
- **Duplicate registrations allowed** — the `registrations` table has no unique constraint on `(event_id, user_id)`, so a user can register for the same event multiple times.
- **Missing-event returns 500** — `GetEventByID` returns `sql.ErrNoRows` for unknown IDs, which handlers surface as `500` rather than `404 Not Found`.
- **No pagination or filtering** — `GET /events` returns every event.
- **No endpoint to list a user's own registrations** — registrations can be created/cancelled but not queried.
- **No request rate limiting, CORS configuration, or HTTPS** — consider adding these behind a reverse proxy or middleware.
