package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/alfafaa/alfafaa-blog/docs"
	"github.com/alfafaa/alfafaa-blog/internal/config"
	"github.com/alfafaa/alfafaa-blog/internal/database"
	"github.com/alfafaa/alfafaa-blog/internal/handlers"
	"github.com/alfafaa/alfafaa-blog/internal/middlewares"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/alfafaa/alfafaa-blog/internal/services"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title Alfafaa Blog API
// @version 1.0
// @description Production-ready blog backend API for Alfafaa Community - a community blogging platform.
// @termsOfService http://swagger.io/terms/

// @contact.name Alfafaa Support
// @contact.url https://alfafaa.com/support
// @contact.email support@alfafaa.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Parse command line flags
	migrateFlag := flag.Bool("migrate", false, "Run database migrations")
	seedFlag := flag.Bool("seed", false, "Seed the database with initial data")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize structured logger
	if err := utils.InitLogger(cfg.Server.Mode); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer utils.Logger.Sync()

	utils.Info("Starting Alfafaa Blog API", zap.String("mode", cfg.Server.Mode))

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// Run migrations if flag is set
	if *migrateFlag {
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Migrations completed successfully")
		return
	}

	// Run migrations automatically on startup
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed database if flag is set
	if *seedFlag {
		if err := database.SeedDatabase(db); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
		log.Println("Database seeded successfully")
		return
	}

	// Register custom validators
	utils.RegisterCustomValidators()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	articleRepo := repositories.NewArticleRepository(db)
	mediaRepo := repositories.NewMediaRepository(db)
	commentRepo := repositories.NewCommentRepository(db)
	engagementRepo := repositories.NewEngagementRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWT)
	userService := services.NewUserService(userRepo, articleRepo, engagementRepo)
	categoryService := services.NewCategoryService(categoryRepo, articleRepo)
	tagService := services.NewTagService(tagRepo, articleRepo)
	articleService := services.NewArticleService(db, articleRepo, categoryRepo, tagRepo,
		services.WithEngagementRepo(engagementRepo),
		services.WithUserRepo(userRepo),
	)
	mediaService := services.NewMediaService(mediaRepo, cfg.Upload)
	searchService := services.NewSearchService(articleRepo, categoryRepo, tagRepo)
	engagementService := services.NewEngagementService(engagementRepo, articleRepo, commentRepo, userRepo)


	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	userActionHandler := handlers.NewUserActionHandler(userService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	tagHandler := handlers.NewTagHandler(tagService)
	articleHandler := handlers.NewArticleHandler(articleService, engagementService)
	mediaHandler := handlers.NewMediaHandler(mediaService, cfg.Upload)
	searchHandler := handlers.NewSearchHandler(searchService)
	engagementHandler := handlers.NewEngagementHandler(engagementService)

	// Create Gin router (use gin.New() to avoid default middleware)
	router := gin.New()

	// Set trusted proxies
	if len(cfg.Security.TrustedProxies) > 0 {
		router.SetTrustedProxies(cfg.Security.TrustedProxies)
	}

	// Apply middleware chain in correct order:
	// 1. Request ID - First to track all requests
	router.Use(middlewares.RequestIDMiddleware())

	// 2. Recovery - Catch panics early
	router.Use(middlewares.RecoveryMiddleware())

	// 3. Logger - Log all requests with context
	router.Use(middlewares.LoggerMiddleware())

	// 4. Security Headers - Add security headers to all responses
	router.Use(middlewares.SecurityHeadersMiddleware())

	// 5. CORS - Handle cross-origin requests
	if cfg.Server.Mode == "release" || cfg.Server.Mode == "production" {
		router.Use(middlewares.CORSMiddleware(middlewares.ProductionCORSConfig(cfg.CORS.AllowedOrigins)))
	} else {
		router.Use(middlewares.DevelopmentCORS())
	}

	// Apply general rate limiting if enabled
	if cfg.Security.EnableRateLimit {
		router.Use(middlewares.GeneralRateLimiter())
	}

	// Serve uploaded files
	router.Static("/uploads", cfg.Upload.Path)

	// Swagger documentation - override host dynamically
	if swaggerHost := os.Getenv("SWAGGER_HOST"); swaggerHost != "" {
		docs.SwaggerInfo.Host = swaggerHost
	}
	if os.Getenv("SWAGGER_SCHEMES") == "https" {
		docs.SwaggerInfo.Schemes = []string{"https"}
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "message": "Alfafaa Blog API is running"})
		})

		// Auth routes (with strict rate limiting to prevent brute force)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", middlewares.AuthRateLimiter(), authHandler.Register)
			auth.POST("/login", middlewares.AuthRateLimiter(), authHandler.Login)
			auth.POST("/google", middlewares.AuthRateLimiter(), authHandler.GoogleAuth)
			auth.POST("/refresh-token", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", middlewares.AuthMiddleware(cfg.JWT.Secret), authHandler.GetMe)
			auth.POST("/change-password", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.StrictRateLimiter(), authHandler.ChangePassword)
		}

		// User routes
		users := v1.Group("/users")
		{
			users.GET("", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireEditor(), userHandler.GetUsers)
			users.GET("/:id", userHandler.GetUser)
			users.GET("/:id/profile", middlewares.OptionalAuthMiddleware(cfg.JWT.Secret), userActionHandler.GetUserProfile)
			users.PUT("/:id", middlewares.AuthMiddleware(cfg.JWT.Secret), userHandler.UpdateUser)
			users.PUT("/:id/admin", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireAdmin(), userHandler.AdminUpdateUser)
			users.DELETE("/:id", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireAdmin(), userHandler.DeleteUser)
			users.GET("/:id/articles", userHandler.GetUserArticles)
			// Social graph routes
			users.POST("/:id/follow", middlewares.AuthMiddleware(cfg.JWT.Secret), userActionHandler.FollowUser)
			users.POST("/:id/unfollow", middlewares.AuthMiddleware(cfg.JWT.Secret), userActionHandler.UnfollowUser)
			users.GET("/:id/followers", userActionHandler.GetFollowers)
			users.GET("/:id/following", userActionHandler.GetFollowing)
			// Interest routes (for current user)
			users.POST("/interests", middlewares.AuthMiddleware(cfg.JWT.Secret), userActionHandler.SetInterests)
			users.GET("/interests", middlewares.AuthMiddleware(cfg.JWT.Secret), userActionHandler.GetInterests)
			// Bookmarked articles
			users.GET("/bookmarks", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.GetBookmarkedArticles)
		}

		// Article routes
		articles := v1.Group("/articles")
		{
			// Public routes (with optional auth for view tracking)
			articles.GET("", middlewares.OptionalAuthMiddleware(cfg.JWT.Secret), articleHandler.GetArticles)
			articles.GET("/trending", articleHandler.GetTrendingArticles)
			articles.GET("/recent", articleHandler.GetRecentArticles)
			articles.GET("/staff-picks", userActionHandler.GetStaffPicks)
			articles.GET("/feed", middlewares.AuthMiddleware(cfg.JWT.Secret), userActionHandler.GetPersonalizedFeed)
			articles.GET("/:slug", middlewares.OptionalAuthMiddleware(cfg.JWT.Secret), articleHandler.GetArticle)
			articles.GET("/:slug/related", articleHandler.GetRelatedArticles)

			// Engagement routes (likes, bookmarks, comments)
			articles.POST("/:slug/like", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.LikeArticle)
			articles.DELETE("/:slug/like", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.UnlikeArticle)
			articles.GET("/:slug/like", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.GetLikeStatus)
			articles.POST("/:slug/bookmark", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.BookmarkArticle)
			articles.DELETE("/:slug/bookmark", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.UnbookmarkArticle)
			articles.GET("/:slug/comments", engagementHandler.GetComments)
			articles.POST("/:slug/comments", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.CreateComment)
			articles.PUT("/:slug/comments/:id", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.UpdateComment)
			articles.DELETE("/:slug/comments/:id", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.DeleteComment)

			// Protected routes (use :slug param name to match Gin's requirement for
			// consistent wildcard names; the value is still a UUID for these routes)
			articles.POST("", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireAuthor(), articleHandler.CreateArticle)
			articles.PUT("/:slug", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireAuthor(), articleHandler.UpdateArticle)
			articles.DELETE("/:slug", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireAuthor(), articleHandler.DeleteArticle)
			articles.PATCH("/:slug/publish", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireEditor(), articleHandler.PublishArticle)
			articles.PATCH("/:slug/unpublish", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireEditor(), articleHandler.UnpublishArticle)
		}

		// Category routes
		categories := v1.Group("/categories")
		{
			categories.GET("", categoryHandler.GetCategories)
			categories.GET("/:slug", categoryHandler.GetCategory)
			categories.GET("/:slug/articles", categoryHandler.GetCategoryArticles)
			categories.POST("", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireEditor(), categoryHandler.CreateCategory)
			categories.PUT("/:slug", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireEditor(), categoryHandler.UpdateCategory)
			categories.DELETE("/:slug", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireEditor(), categoryHandler.DeleteCategory)
		}

		// Tag routes
		tags := v1.Group("/tags")
		{
			tags.GET("", tagHandler.GetTags)
			tags.GET("/popular", tagHandler.GetPopularTags)
			tags.GET("/:slug", tagHandler.GetTag)
			tags.GET("/:slug/articles", tagHandler.GetTagArticles)
			tags.POST("", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireEditor(), tagHandler.CreateTag)
			tags.PUT("/:slug", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireEditor(), tagHandler.UpdateTag)
			tags.DELETE("/:slug", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireEditor(), tagHandler.DeleteTag)
		}

		// Media routes (with upload rate limiting to prevent abuse)
		media := v1.Group("/media")
		{
			media.POST("/upload", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.UploadRateLimiter(), mediaHandler.UploadMedia)
			media.GET("/:id", mediaHandler.GetMedia)
			media.GET("", middlewares.AuthMiddleware(cfg.JWT.Secret), middlewares.RequireAdmin(), mediaHandler.GetAllMedia)
			media.DELETE("/:id", middlewares.AuthMiddleware(cfg.JWT.Secret), mediaHandler.DeleteMedia)
		}

		// Notification routes
		notifications := v1.Group("/notifications")
		{
			notifications.GET("", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.GetNotifications)
			notifications.GET("/unread-count", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.GetUnreadCount)
			notifications.PUT("/:id/read", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.MarkNotificationAsRead)
			notifications.PUT("/read-all", middlewares.AuthMiddleware(cfg.JWT.Secret), engagementHandler.MarkAllNotificationsAsRead)
		}

		// Search route (with search rate limiting)
		v1.GET("/search", middlewares.SearchRateLimiter(), searchHandler.Search)
	}

	// Ensure upload directory exists
	if err := os.MkdirAll(cfg.Upload.Path, 0755); err != nil {
		utils.Warn("Failed to create upload directory", zap.Error(err))
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	utils.Info("Server starting",
		zap.String("address", addr),
		zap.String("swagger", fmt.Sprintf("http://%s/swagger/index.html", addr)),
		zap.String("health", fmt.Sprintf("http://%s/api/v1/health", addr)),
	)

	if err := router.Run(addr); err != nil {
		utils.Fatal("Failed to start server", zap.Error(err))
	}
}
