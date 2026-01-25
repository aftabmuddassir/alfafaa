# Alfafaa Blog Backend API

Production-ready blog backend API for Alfafaa Community

## Tech Stack

- **Language**: Go 1.22+
- **Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL
- **Authentication**: JWT

## Quick Start

### Prerequisites

- Go 1.22 or later
- PostgreSQL 14+
- Make (optional, for using Makefile commands)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/alfafaa/alfafaa-blog.git
cd alfafaa-blog
```

2. Copy environment file and configure:
```bash
cp .env.example .env
# Edit .env with your database credentials
```

3. Install dependencies:
```bash
go mod download
```

4. Run migrations and seed data:
```bash
go run cmd/server/main.go -migrate
go run cmd/server/main.go -seed
```

5. Start the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

### Using Makefile

```bash
make deps      # Download dependencies
make run       # Run the server
make migrate   # Run migrations
make seed      # Seed the database
make build     # Build binary
make test      # Run tests
```

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register a new user |
| POST | `/api/v1/auth/login` | Login and get tokens |
| POST | `/api/v1/auth/refresh-token` | Refresh access token |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/auth/me` | Get current user |

### Users
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/users` | List users (editor+) |
| GET | `/api/v1/users/:id` | Get user by ID |
| PUT | `/api/v1/users/:id` | Update user |
| DELETE | `/api/v1/users/:id` | Delete user (admin) |
| GET | `/api/v1/users/:id/articles` | Get user's articles |

### Articles
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/articles` | List articles |
| GET | `/api/v1/articles/:slug` | Get article by slug |
| POST | `/api/v1/articles` | Create article (author+) |
| PUT | `/api/v1/articles/:id` | Update article |
| DELETE | `/api/v1/articles/:id` | Delete article |
| PATCH | `/api/v1/articles/:id/publish` | Publish (editor+) |
| PATCH | `/api/v1/articles/:id/unpublish` | Unpublish (editor+) |
| GET | `/api/v1/articles/trending` | Get trending articles |
| GET | `/api/v1/articles/recent` | Get recent articles |
| GET | `/api/v1/articles/:slug/related` | Get related articles |

### Categories
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/categories` | List categories |
| GET | `/api/v1/categories/:slug` | Get category by slug |
| POST | `/api/v1/categories` | Create category (editor+) |
| PUT | `/api/v1/categories/:id` | Update category (editor+) |
| DELETE | `/api/v1/categories/:id` | Delete category (editor+) |
| GET | `/api/v1/categories/:slug/articles` | Get category articles |

### Tags
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/tags` | List tags |
| GET | `/api/v1/tags/popular` | Get popular tags |
| GET | `/api/v1/tags/:slug` | Get tag by slug |
| POST | `/api/v1/tags` | Create tag (editor+) |
| PUT | `/api/v1/tags/:id` | Update tag (editor+) |
| DELETE | `/api/v1/tags/:id` | Delete tag (editor+) |
| GET | `/api/v1/tags/:slug/articles` | Get tag articles |

### Media
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/media/upload` | Upload file |
| GET | `/api/v1/media/:id` | Get media by ID |
| GET | `/api/v1/media` | List all media (admin) |
| DELETE | `/api/v1/media/:id` | Delete media |

### Search
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/search?q=query` | Search articles, categories, tags |

## User Roles

| Role | Permissions |
|------|-------------|
| **reader** | Read articles, comment |
| **author** | Create/edit own articles |
| **editor** | Manage all articles, categories, tags |
| **admin** | Full system access |

## Project Structure

```
alfafaa-blog/
├── cmd/server/main.go           # Entry point
├── internal/
│   ├── config/                  # Configuration
│   ├── database/                # DB connection & migrations
│   ├── dto/                     # Data Transfer Objects
│   ├── handlers/                # HTTP handlers
│   ├── middlewares/             # Middlewares
│   ├── models/                  # GORM models
│   ├── repositories/            # Data access layer
│   ├── services/                # Business logic
│   └── utils/                   # Utilities
├── uploads/                     # File uploads
├── .env.example                 # Environment template
├── go.mod                       # Go module
└── Makefile                     # Build commands
```


## Response Format

### Success Response
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description",
    "details": [...]
  }
}
```

## License

Copyright (c) 2026 Alfafaa Community
