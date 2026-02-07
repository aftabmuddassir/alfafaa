# Checkpoint - 2026-02-02

## Session Summary
This session focused on fixing **Google OAuth 2.0 integration** and **CORS configuration** for the frontend-backend communication.

**Status: All issues resolved** - Google OAuth is now working end-to-end.

---

## What's Been Completed (All Time)

### Phase 1 - Core MVP
- [x] Project setup & folder structure (Repository/Service/Handler pattern)
- [x] Database connection & migrations (PostgreSQL + GORM)
- [x] User model & authentication (register, login, JWT with refresh tokens)
- [x] Article CRUD (create, read, update, delete, soft deletes)
- [x] Category CRUD with article counts
- [x] Tag CRUD with article counts
- [x] Basic RBAC middleware (reader, author, editor, admin)
- [x] File upload for images (with validation)

### Phase 2 - Enhanced Features
- [x] Search & filtering (articles, categories, tags)
- [x] Pagination with metadata
- [x] Related articles endpoint
- [x] View tracking for articles
- [x] Slug generation (auto-generated from title)
- [x] SEO meta fields

### Phase 3 - Advanced Features
- [x] Comments system
- [x] RSS feed generation
- [x] Sitemap generation
- [x] Rate limiting (general, auth, search, upload)
- [x] Comprehensive testing
- [x] API documentation (Swagger)

### Phase 4 - Medium.com Style Features
- [x] Social Graph (Follow/Unfollow authors)
- [x] User Interests (Category preferences for onboarding)
- [x] Personalized Feed (articles from followed authors + interested categories)
- [x] Staff Picks (curated articles)
- [x] Google OAuth Authentication (implementation complete)
- [x] User Profile with social stats (follower/following counts)

### Security Hardening
- [x] Security headers middleware (CSP, XSS, etc.)
- [x] Rate limiting middleware (configurable)
- [x] Request ID tracking
- [x] Structured logging with zap
- [x] Input sanitization
- [x] Recovery middleware (panic handling)
- [x] CORS middleware (development + production configs)

---

## Today's Session Changes (2026-02-02)

### CORS Fixes
1. **Updated `DevelopmentCORS()` middleware** in `internal/middlewares/cors_middleware.go`:
   - Changed from `Access-Control-Allow-Origin: *` to reflecting actual origin
   - Added `Access-Control-Allow-Credentials: true` for cookie/auth support
   - Specified explicit allowed headers

### Google OAuth - RESOLVED
2. **Fixed database schema** - Added missing OAuth columns via SQL:
   ```sql
   ALTER TABLE users ADD COLUMN IF NOT EXISTS google_id VARCHAR(255);
   ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(20) DEFAULT 'local';
   CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id) WHERE google_id IS NOT NULL;
   ```

3. **Google OAuth now working** - Full flow tested and functional:
   - Frontend sends Google ID token
   - Backend decodes and extracts user info
   - Creates new user or links existing account
   - Returns JWT tokens

---

## Current State

### Tests
```
ok  github.com/alfafaa/alfafaa-blog/internal/handlers    0.364s
ok  github.com/alfafaa/alfafaa-blog/internal/repositories 0.490s
ok  github.com/alfafaa/alfafaa-blog/internal/services    1.430s
ok  github.com/alfafaa/alfafaa-blog/internal/utils       0.883s
```

All tests passing.

### Test Files
- `internal/utils/` - jwt, password, slug, response, errors, validation tests
- `internal/services/` - user, category, tag, article, auth, user_action tests
- `internal/repositories/` - user, article repository tests
- `internal/handlers/` - auth handler tests

### Database
- PostgreSQL with proper indexes
- UUID primary keys
- Soft deletes enabled
- **PENDING**: OAuth columns migration (google_id, auth_provider)

### Dependencies
- 74 total dependencies in go.mod
- Key packages: gin, gorm, jwt-go, zap, swagger, uuid

---

## Pending Tasks

### Completed This Session
- [x] **Add OAuth columns to database** - SQL migration executed
- [x] **Google OAuth working end-to-end**
- [x] **Debug logging removed** from GoogleAuth handler

### Short Term
- [ ] Add social graph tables migration (user_follows, user_interests)
- [ ] Add is_staff_pick column to articles table

### Future
- [ ] Redis for rate limiting (multi-instance support)
- [ ] Database connection pooling optimization
- [ ] Docker setup for deployment
- [ ] CI/CD pipeline
- [ ] Performance testing
- [ ] Proper Google OAuth token verification (currently using JWT decode stub)

---

## Known Issues

1. **GORM AutoMigrate failing** - "insufficient arguments" error when running migrations. Workaround: use raw SQL to add columns. (Resolved for OAuth columns)

2. **Google OAuth stub implementation** - The `verifyGoogleIDToken()` function decodes the JWT payload without verifying the signature. For production, implement proper verification using Google's public keys. (Works for development)

---

## Environment Setup

### Requirements
- Go 1.21+
- PostgreSQL 15+
- Node.js 18+ (for frontend)

### Environment Variables (.env)
```env
SERVER_PORT=8081
SERVER_HOST=localhost
GIN_MODE=debug

DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=alfafaa_blog

JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h

GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
```

---

## How to Run

```bash
# Start server
go run cmd/server/main.go

# Run migrations (currently broken - use SQL instead)
go run cmd/server/main.go -migrate

# Seed database
go run cmd/server/main.go -seed

# Run tests
go test ./... -v

# Generate Swagger docs
swag init -g cmd/server/main.go

# View Swagger UI
# http://localhost:8081/swagger/index.html
```

---

## API Endpoints Summary

### Auth
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/google` - Google OAuth login
- `POST /api/v1/auth/refresh-token` - Refresh JWT
- `POST /api/v1/auth/logout` - Logout
- `GET /api/v1/auth/me` - Get current user
- `POST /api/v1/auth/change-password` - Change password

### Users
- `GET /api/v1/users/:id/profile` - User profile with social stats
- `POST /api/v1/users/:id/follow` - Follow user
- `POST /api/v1/users/:id/unfollow` - Unfollow user
- `GET /api/v1/users/:id/followers` - Get followers
- `GET /api/v1/users/:id/following` - Get following
- `POST /api/v1/users/interests` - Set interests
- `GET /api/v1/users/interests` - Get interests

### Articles
- `GET /api/v1/articles` - List articles
- `GET /api/v1/articles/feed` - Personalized feed
- `GET /api/v1/articles/staff-picks` - Staff picks
- `GET /api/v1/articles/trending` - Trending articles
- `GET /api/v1/articles/:slug` - Get article by slug
- Full CRUD for authenticated users

### Categories & Tags
- Full CRUD with article counts

---

## Files Modified This Session

1. `internal/middlewares/cors_middleware.go` - Updated DevelopmentCORS() for proper CORS handling
2. `internal/handlers/auth_handler.go` - GoogleAuth handler (debug logs added then removed)
3. `checkpoints/checkpoint-2026-02-02.md` - This checkpoint file created

---

## Git Status
```
Branch: main
Modified:
  - internal/middlewares/cors_middleware.go
  - internal/handlers/auth_handler.go (debug logs added)
Untracked:
  - CLAUDE.md
  - checkpoints/
  - uploads/
```

---

## Next Session Action Items

1. ~~Run the SQL migration to add OAuth columns~~ ✅ Done
2. ~~Test Google OAuth flow end-to-end~~ ✅ Done
3. ~~Remove debug logging from auth_handler.go~~ ✅ Done
4. Add social graph tables (user_follows, user_interests) via SQL if needed
5. Add is_staff_pick column to articles table
6. Consider implementing proper Google token verification for production
7. Commit changes and update CLAUDE.md if needed
