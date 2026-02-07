# Alfafaa Community Blog - Project Context

## Project Overview
This is the backend API for Alfafaa Community Blog (blog.alfafaa.com), a community blogging platform. This blog is **part of a larger Alfafaa ecosystem** that includes a social network, job board, events platform, and classifieds.

## Critical Architecture Decisions

### Multi-Service Integration
**IMPORTANT**: This blog shares users with other Alfafaa services. Design decisions must account for:
- Shared authentication across services (JWT tokens should work across all Alfafaa platforms)
- Centralized user management (users exist in one place, used everywhere)
- Consistent API patterns across all Alfafaa services
- Future microservices architecture

### Tech Stack (Non-Negotiable)
- **Language**: Go (latest stable version)
- **Framework**: Gin (github.com/gin-gonic/gin) - lightweight, fast HTTP framework
- **ORM**: GORM (gorm.io/gorm) - most popular Go ORM
- **Database**: PostgreSQL (chosen for JSON support and future directory/maps features)
- **Authentication**: JWT with refresh tokens
- **Frontend**: Next.js (separate repository, consumes this API)

## Code Standards & Conventions

### Project Structure Pattern
Always follow this structure:
```
alfafaa-blog/
├── cmd/server/main.go          # Entry point only, minimal logic
├── internal/                    # Private application code
│   ├── config/                 # Configuration management
│   ├── models/                 # GORM models (database entities)
│   ├── dto/                    # Data Transfer Objects (request/response)
│   ├── repositories/           # Data access layer
│   ├── services/               # Business logic layer
│   ├── handlers/               # HTTP handlers (controllers)
│   ├── middlewares/            # HTTP middlewares
│   ├── utils/                  # Helper functions
│   └── database/               # DB connection & migrations
├── pkg/                        # Public packages (reusable across Alfafaa)
└── uploads/                    # Local file storage
```

### Naming Conventions
- **Files**: `snake_case.go` (e.g., `article_service.go`)
- **Packages**: lowercase, singular (e.g., `model`, `service`, not `models`, `services`)
- **Interfaces**: `I` prefix optional, use `er` suffix (e.g., `ArticleRepository` or `IArticleRepository`)
- **Structs**: `PascalCase` (e.g., `Article`, `UserDTO`)
- **Functions/Methods**: `PascalCase` for exported, `camelCase` for unexported
- **Constants**: `PascalCase` or `SCREAMING_SNAKE_CASE` for groups
- **Database tables**: `snake_case`, plural (e.g., `articles`, `categories`)

### Code Organization Principles

#### 1. Repository Pattern (Data Access Layer)
```go
// repositories/article_repository.go
type ArticleRepository interface {
    Create(article *models.Article) error
    FindByID(id string) (*models.Article, error)
    FindBySlug(slug string) (*models.Article, error)
    FindAll(filters ArticleFilters) ([]models.Article, int64, error)
    Update(article *models.Article) error
    Delete(id string) error
}

type articleRepository struct {
    db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) ArticleRepository {
    return &articleRepository{db: db}
}
```

**Rules for Repositories:**
- Only database operations
- No business logic
- Return raw models
- Handle GORM relationships (preloading, eager loading)
- Use GORM best practices to prevent N+1 queries

#### 2. Service Pattern (Business Logic Layer)
```go
// services/article_service.go
type ArticleService interface {
    CreateArticle(dto *dto.CreateArticleDTO, userID string) (*dto.ArticleResponseDTO, error)
    GetArticle(slug string) (*dto.ArticleDetailDTO, error)
    UpdateArticle(id string, dto *dto.UpdateArticleDTO, userID string) error
    DeleteArticle(id string, userID string) error
    PublishArticle(id string) error
}

type articleService struct {
    repo ArticleRepository
    categoryRepo CategoryRepository
    tagRepo TagRepository
}

func NewArticleService(repo ArticleRepository, categoryRepo CategoryRepository, tagRepo TagRepository) ArticleService {
    return &articleService{
        repo: repo,
        categoryRepo: categoryRepo,
        tagRepo: tagRepo,
    }
}
```

**Rules for Services:**
- Contains ALL business logic
- Validates business rules
- Coordinates between multiple repositories
- Transforms models to DTOs
- Handles errors with context
- Never return GORM models directly (use DTOs)

#### 3. Handler Pattern (HTTP Layer)
```go
// handlers/article_handler.go
type ArticleHandler struct {
    service ArticleService
}

func NewArticleHandler(service ArticleService) *ArticleHandler {
    return &ArticleHandler{service: service}
}

func (h *ArticleHandler) CreateArticle(c *gin.Context) {
    var dto dto.CreateArticleDTO
    if err := c.ShouldBindJSON(&dto); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
        return
    }
    
    userID := c.GetString("userID") // from auth middleware
    
    article, err := h.service.CreateArticle(&dto, userID)
    if err != nil {
        utils.HandleServiceError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusCreated, "Article created successfully", article)
}
```

**Rules for Handlers:**
- Only HTTP concerns (request/response)
- Parse request, call service, return response
- No business logic
- Use helper functions for consistent responses
- Extract user info from context (set by middleware)

### Error Handling Strategy

#### Custom Error Types
```go
// utils/errors.go
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"-"`
}

var (
    ErrNotFound = &AppError{Code: "NOT_FOUND", Message: "Resource not found", Status: 404}
    ErrUnauthorized = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized", Status: 401}
    ErrForbidden = &AppError{Code: "FORBIDDEN", Message: "Forbidden", Status: 403}
    ErrValidation = &AppError{Code: "VALIDATION_ERROR", Message: "Validation failed", Status: 400}
    ErrInternal = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error", Status: 500}
)
```

#### Error Handling in Services
```go
func (s *articleService) GetArticle(slug string) (*dto.ArticleDetailDTO, error) {
    article, err := s.repo.FindBySlug(slug)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, utils.ErrNotFound
        }
        return nil, fmt.Errorf("failed to fetch article: %w", err)
    }
    
    return s.toDetailDTO(article), nil
}
```

### Response Format Standards

#### Success Response
```go
// utils/response.go
type SuccessResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Page       int   `json:"page"`
    PerPage    int   `json:"per_page"`
    Total      int64 `json:"total"`
    TotalPages int   `json:"total_pages"`
}

func SuccessResponse(c *gin.Context, status int, message string, data interface{}) {
    c.JSON(status, SuccessResponse{
        Success: true,
        Message: message,
        Data:    data,
    })
}
```

#### Error Response
```go
type ErrorResponse struct {
    Success bool         `json:"success"`
    Error   ErrorDetails `json:"error"`
}

type ErrorDetails struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details []ValidationError      `json:"details,omitempty"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

### Database Best Practices

#### Model Definition
```go
// models/article.go
type Article struct {
    ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    Title           string         `gorm:"type:varchar(255);not null;index" json:"title"`
    Slug            string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
    Content         string         `gorm:"type:text;not null" json:"content"`
    Excerpt         string         `gorm:"type:text" json:"excerpt"`
    FeaturedImageURL *string       `gorm:"type:varchar(500)" json:"featured_image_url"`
    AuthorID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"author_id"`
    Author          *User          `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
    Status          ArticleStatus  `gorm:"type:varchar(20);not null;default:'draft';index" json:"status"`
    PublishedAt     *time.Time     `gorm:"index" json:"published_at"`
    ViewCount       int            `gorm:"default:0" json:"view_count"`
    IsStaffPick     bool           `gorm:"default:false;index" json:"is_staff_pick"` // Medium-style staff picks
    Categories      []Category     `gorm:"many2many:article_categories;" json:"categories,omitempty"`
    Tags            []Tag          `gorm:"many2many:article_tags;" json:"tags,omitempty"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
    DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// models/user.go - Social Graph additions
type User struct {
    // ... existing fields ...

    // OAuth fields
    GoogleID     *string `gorm:"type:varchar(255);uniqueIndex" json:"-"`
    AuthProvider string  `gorm:"type:varchar(20);default:'local'" json:"auth_provider"` // local, google

    // Social Graph - User Interests (Categories)
    Interests []Category `gorm:"many2many:user_interests;" json:"interests,omitempty"`

    // Social Graph - Followers/Following
    Followers []*User `gorm:"many2many:user_follows;joinForeignKey:FollowingID;joinReferences:FollowerID" json:"followers,omitempty"`
    Following []*User `gorm:"many2many:user_follows;joinForeignKey:FollowerID;joinReferences:FollowingID" json:"following,omitempty"`
}

// UserFollow - Join table for follow relationships
type UserFollow struct {
    FollowerID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"follower_id"`
    FollowingID uuid.UUID `gorm:"type:uuid;primaryKey" json:"following_id"`
    CreatedAt   time.Time `json:"created_at"`
}

// UserInterest - Join table for user interests
type UserInterest struct {
    UserID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
    CategoryID uuid.UUID `gorm:"type:uuid;primaryKey" json:"category_id"`
    CreatedAt  time.Time `json:"created_at"`
}
```

**Key Points:**
- Use `uuid.UUID` for all IDs
- Index frequently queried fields
- Use `gorm.DeletedAt` for soft deletes
- Use pointer types for nullable fields
- Use enums for status fields
- Tag with both `gorm` and `json` tags

#### Migration Strategy
```go
// database/migrations.go
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.User{},
        &models.Article{},
        &models.Category{},
        &models.Tag{},
        &models.Media{},
        &models.Comment{},
        // Social graph join tables
        &models.UserFollow{},
        &models.UserInterest{},
    )
}
```

#### Prevent N+1 Queries
```go
// GOOD: Preload relationships
articles, err := repo.db.
    Preload("Author").
    Preload("Categories").
    Preload("Tags").
    Where("status = ?", "published").
    Find(&articles).Error

// BAD: Will cause N+1 queries
articles, err := repo.db.Find(&articles).Error
// Then accessing article.Author will trigger additional queries
```

### Authentication & Authorization

#### JWT Structure
```go
type JWTClaims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}
```

#### Auth Middleware
```go
// middlewares/auth_middleware.go
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractTokenFromHeader(c)
        if token == "" {
            utils.ErrorResponse(c, http.StatusUnauthorized, "Missing token", nil)
            c.Abort()
            return
        }
        
        claims, err := utils.ValidateToken(token, jwtSecret)
        if err != nil {
            utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token", err)
            c.Abort()
            return
        }
        
        c.Set("userID", claims.UserID)
        c.Set("userRole", claims.Role)
        c.Next()
    }
}
```

#### Role-Based Middleware
```go
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("userRole")
        
        allowed := false
        for _, role := range roles {
            if userRole == role {
                allowed = true
                break
            }
        }
        
        if !allowed {
            utils.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", nil)
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Validation Rules

#### Use go-playground/validator
```go
// dto/article_dto.go
type CreateArticleDTO struct {
    Title        string   `json:"title" binding:"required,min=5,max=255"`
    Content      string   `json:"content" binding:"required,min=50"`
    Excerpt      string   `json:"excerpt" binding:"max=500"`
    CategoryIDs  []string `json:"category_ids" binding:"required,min=1"`
    TagIDs       []string `json:"tag_ids"`
    Status       string   `json:"status" binding:"omitempty,oneof=draft published"`
}
```

#### Custom Validators
```go
// utils/validators.go
func ValidateUUID(fl validator.FieldLevel) bool {
    _, err := uuid.Parse(fl.Field().String())
    return err == nil
}

// Register in main.go
if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
    v.RegisterValidation("uuid", utils.ValidateUUID)
}
```

### File Upload Best Practices

```go
func (h *MediaHandler) UploadFile(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "No file provided", err)
        return
    }
    
    // Validate file type
    if !utils.IsValidImageType(file.Header.Get("Content-Type")) {
        utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file type", nil)
        return
    }
    
    // Validate file size (10MB max)
    if file.Size > 10*1024*1024 {
        utils.ErrorResponse(c, http.StatusBadRequest, "File too large", nil)
        return
    }
    
    // Generate unique filename
    filename := fmt.Sprintf("%s-%s", uuid.New().String(), file.Filename)
    filepath := fmt.Sprintf("uploads/%s", filename)
    
    // Save file
    if err := c.SaveUploadedFile(file, filepath); err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file", err)
        return
    }
    
    // Create media record
    media := &models.Media{
        Filename:         filename,
        OriginalFilename: file.Filename,
        FilePath:         filepath,
        FileSize:         file.Size,
        MimeType:         file.Header.Get("Content-Type"),
        UploadedBy:       uuid.MustParse(c.GetString("userID")),
    }
    
    if err := h.service.CreateMedia(media); err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create media record", err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusCreated, "File uploaded successfully", media)
}
```

### Environment Configuration

```go
// config/config.go
type Config struct {
    Server      ServerConfig
    Database    DatabaseConfig
    JWT         JWTConfig
    Upload      UploadConfig
    CORS        CORSConfig
    RateLimit   RateLimitConfig
    Security    SecurityConfig
    GoogleOAuth GoogleOAuthConfig  // Added for OAuth
}

type GoogleOAuthConfig struct {
    ClientID     string   // from GOOGLE_CLIENT_ID env var
    ClientSecret string   // from GOOGLE_CLIENT_SECRET env var
    RedirectURLs []string // from GOOGLE_REDIRECT_URLS env var
}

type ServerConfig struct {
    Port string
    Host string
    Mode string // debug, release
}

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

type JWTConfig struct {
    Secret           string
    Expiration       time.Duration
    RefreshExpiration time.Duration
}

func LoadConfig() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }
    
    return &Config{
        Server: ServerConfig{
            Port: getEnv("SERVER_PORT", "8080"),
            Host: getEnv("SERVER_HOST", "localhost"),
            Mode: getEnv("GIN_MODE", "debug"),
        },
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnv("DB_PORT", "5432"),
            User:     getEnv("DB_USER", "postgres"),
            Password: getEnv("DB_PASSWORD", ""),
            DBName:   getEnv("DB_NAME", "alfafaa_blog"),
            SSLMode:  getEnv("DB_SSLMODE", "disable"),
        },
        JWT: JWTConfig{
            Secret:            getEnv("JWT_SECRET", ""),
            Expiration:        parseDuration(getEnv("JWT_EXPIRATION", "24h")),
            RefreshExpiration: parseDuration(getEnv("JWT_REFRESH_EXPIRATION", "168h")),
        },
    }, nil
}
```

## Important Context for AI Assistance

### When Making Changes
1. **Always maintain the three-layer architecture**: Repository → Service → Handler
2. **Never put business logic in handlers or repositories**
3. **Always use DTOs for request/response**, never expose GORM models directly
4. **Always add indexes** on fields used in WHERE clauses or JOINs
5. **Always use UUIDs** for primary keys
6. **Always implement soft deletes** using gorm.DeletedAt
7. **Always validate input** at the handler level before passing to service

### When Adding New Features
1. Create the model first
2. Add migration in database package
3. Create repository interface and implementation
4. Create service interface and implementation
5. Create DTOs for request/response
6. Create handler
7. Register routes in main.go

### Testing Strategy
- **Unit tests**: Test services in isolation (mock repositories)
- **Integration tests**: Test repositories with real database (use testcontainers)
- **E2E tests**: Test full HTTP flow (use httptest)

### Common Pitfalls to Avoid
1. **N+1 Query Problem**: Always use Preload() for relationships
2. **Password in Response**: Always omit password from JSON (use `json:"-"`)
3. **Missing Validation**: Always validate user input
4. **Hardcoded Configs**: Always use environment variables
5. **SQL Injection**: Let GORM handle it (never use raw SQL without parameters)
6. **Memory Leaks**: Close database connections properly
7. **Exposing Internal Errors**: Always sanitize errors before sending to client

### Future Integration Points
This blog will integrate with:
- **Alfafaa Social Network**: Shared user authentication
- **Alfafaa Jobs**: Cross-promotion of content
- **Alfafaa Events**: Event announcements in blog
- **Alfafaa Classifieds**: Marketplace ads in blog

Keep the codebase modular and the user service portable for future microservices architecture.

## Quick Reference Commands

```bash
# Run server
go run cmd/server/main.go

# Run migrations
go run cmd/server/main.go --migrate

# Run with hot reload (using air)
air

# Build binary
go build -o bin/alfafaa-blog cmd/server/main.go

# Run tests
go test ./... -v

# Generate mocks (if using mockery)
mockery --all --output=mocks

# Format code
go fmt ./...

# Lint code
golangci-lint run
```

## Priority Order for Implementation

**Phase 1 - Core MVP (Week 1-2)**
1. ✅ Project setup & folder structure
2. ✅ Database connection & migrations
3. ✅ User model & authentication (register, login, JWT)
4. ✅ Article CRUD (create, read, update, delete)
5. ✅ Category CRUD
6. ✅ Tag CRUD
7. ✅ Basic RBAC middleware
8. ✅ File upload for images

**Phase 2 - Enhanced Features (Week 3)**
9. ✅ Search & filtering
10. ✅ Pagination
11. ✅ Related articles
12. ✅ View tracking
13. ✅ Slug generation
14. ✅ SEO meta fields

**Phase 3 - Advanced (Week 4)**
15. ✅ Comments system
16. ✅ RSS feed generation
17. ✅ Sitemap generation
18. ✅ Rate limiting
19. ✅ Comprehensive testing
20. ✅ API documentation (Swagger)

**Phase 4 - Medium.com Style Features (Completed)**
21. ✅ Social Graph (Follow/Unfollow authors)
22. ✅ User Interests (Category preferences for onboarding)
23. ✅ Personalized Feed (articles from followed authors + interested categories)
24. ✅ Staff Picks (curated articles)
25. ✅ Google OAuth Authentication (stub implementation)
26. ✅ User Profile with social stats (follower/following counts)

## Medium.com Style Features Documentation

### Social Graph (User Following)

**Models Updated:**
- `internal/models/user.go`: Added `Followers` and `Following` (many-to-many User to User via `user_follows` table)
- `internal/models/user.go`: Added `UserFollow` join table model with timestamps

**New Endpoints:**
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/users/:id/follow` | Required | Follow a user |
| POST | `/api/v1/users/:id/unfollow` | Required | Unfollow a user |
| GET | `/api/v1/users/:id/followers` | Optional | Get user's followers |
| GET | `/api/v1/users/:id/following` | Optional | Get users that user follows |
| GET | `/api/v1/users/:id/profile` | Optional | Get user profile with social stats |

### User Interests (Category Preferences)

**Models Updated:**
- `internal/models/user.go`: Added `Interests` (many-to-many User to Category via `user_interests` table)
- `internal/models/user.go`: Added `UserInterest` join table model with timestamps

**New Endpoints:**
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/users/interests` | Required | Set user interests (for onboarding) |
| GET | `/api/v1/users/interests` | Required | Get current user's interests |

**Request Body for Setting Interests:**
```json
{
  "category_ids": ["uuid-1", "uuid-2", "uuid-3"]
}
```

### Personalized Feed & Staff Picks

**Models Updated:**
- `internal/models/article.go`: Added `IsStaffPick` boolean field (default: false)

**New Endpoints:**
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/articles/feed` | Required | Personalized feed (followed authors + interested categories) |
| GET | `/api/v1/articles/staff-picks` | Optional | Staff-curated articles |

**Repository Methods Added:**
- `FindForUser(userID, followingIDs, interestCategoryIDs, filters)` - Returns personalized articles
- `FindStaffPicks(filters)` - Returns articles where `is_staff_pick = true`
- `SetStaffPick(articleID, isStaffPick)` - Mark/unmark as staff pick

### Google OAuth Authentication

**Configuration (Environment Variables - DO NOT COMMIT):**
```env
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URLS=http://localhost:5173,http://localhost:3000
```

**Models Updated:**
- `internal/models/user.go`: Added `GoogleID` (unique, nullable) and `AuthProvider` (default: "local")

**New Endpoint:**
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/auth/google` | None | Authenticate with Google ID token |

**Request Body:**
```json
{
  "id_token": "google-jwt-id-token"
}
```

**Behavior:**
1. If user exists with matching Google ID → Login
2. If user exists with matching email → Link Google account and login
3. If no user exists → Create new user with Google info and login

**Note:** Current implementation is a **stub** that decodes the JWT payload without full signature verification. For production, implement proper verification using Google's public keys.

### New Handler: user_action_handler.go

Location: `internal/handlers/user_action_handler.go`

Handles all social graph operations:
- Follow/Unfollow users
- Get followers/following lists
- Set/Get interests
- Personalized feed
- Staff picks
- User profiles with social stats

### Database Migrations

Run migrations to add new tables and columns:
```bash
go run cmd/server/main.go -migrate
```

**New Tables Created:**
- `user_follows` (follower_id, following_id, created_at)
- `user_interests` (user_id, category_id, created_at)

**Columns Added:**
- `users.google_id` (VARCHAR, UNIQUE, nullable)
- `users.auth_provider` (VARCHAR, default: 'local')
- `articles.is_staff_pick` (BOOLEAN, default: false)

### Test Coverage

New test file: `internal/services/user_action_test.go`

Tests added for:
- Follow/Unfollow (success, already following, self-follow prevention, user not found)
- Get followers/following
- Set/Get interests
- Personalized feed
- Staff picks
- User profile retrieval

## Resources & Documentation
- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [Go Best Practices](https://go.dev/doc/effective_go)
- [PostgreSQL JSON Support](https://www.postgresql.org/docs/current/datatype-json.html)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)

## Environment Variables Reference

Required environment variables in `.env` file:

```env
# Server
SERVER_PORT=8081
SERVER_HOST=localhost
GIN_MODE=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=alfafaa_blog
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h

# Upload
UPLOAD_MAX_SIZE=10485760
UPLOAD_PATH=./uploads

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000

# Rate Limiting
ENABLE_RATE_LIMIT=true

# Google OAuth (DO NOT COMMIT ACTUAL VALUES)
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URLS=http://localhost:5173,http://localhost:3000
```

---

**Remember**: This is a production system serving a Muslim community. Code quality, security, and reliability are paramount. When in doubt, prioritize correctness over speed.
