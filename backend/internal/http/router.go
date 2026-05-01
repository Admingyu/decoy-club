package http

import (
	"net/http"
	"strings"

	"decoy-club/backend/internal/activities"
	"decoy-club/backend/internal/auth"
	"decoy-club/backend/internal/comments"
	"decoy-club/backend/internal/common/middleware"
	"decoy-club/backend/internal/common/response"
	"decoy-club/backend/internal/config"
	"decoy-club/backend/internal/notifications"
	"decoy-club/backend/internal/posts"
	"decoy-club/backend/internal/uploads"
	"decoy-club/backend/internal/users"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Dependencies struct {
	Config   config.Config
	Database *mongo.Database
}

func NewRouter(deps *Dependencies) *gin.Engine {
	cfg := config.Config{}
	if deps != nil {
		cfg = deps.Config
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.CORS(cfg.FrontendOrigin))
	router.Static("/uploads", cfg.UploadDir)

	var authRepo auth.Repository
	if deps != nil && deps.Database != nil {
		authRepo = auth.NewMongoRepository(deps.Database)
	} else {
		authRepo = auth.NewMemoryRepository()
	}
	authService := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authService)

	var userRepo users.Repository
	if deps != nil && deps.Database != nil {
		userRepo = users.NewMongoRepository(deps.Database)
	} else {
		userRepo = users.NewMemoryRepository()
	}
	userService := users.NewService(userRepo)
	userHandler := users.NewHandler(userService)

	var postRepo posts.Repository
	if deps != nil && deps.Database != nil {
		postRepo = posts.NewMongoRepository(deps.Database, userRepo)
	} else {
		postRepo = posts.NewMemoryRepository(userRepo)
	}
	commentCounter, _ := postRepo.(comments.PostCounter)
	var commentRepo comments.Repository
	if deps != nil && deps.Database != nil {
		commentRepo = comments.NewMongoRepository(deps.Database, commentCounter)
	} else {
		commentRepo = comments.NewMemoryRepository(commentCounter)
	}

	var notificationRepo notifications.Repository
	if deps != nil && deps.Database != nil {
		notificationRepo = notifications.NewMongoRepository(deps.Database)
	} else {
		notificationRepo = notifications.NewMemoryRepository()
	}
	notificationService := notifications.NewService(notificationRepo, userRepo, postRepo, commentRepo)
	notificationHandler := notifications.NewHandler(notificationService)

	var activityRepo activities.Repository
	if deps != nil && deps.Database != nil {
		activityRepo = activities.NewMongoRepository(deps.Database)
	} else {
		activityRepo = activities.NewMemoryRepository()
	}
	activityService := activities.NewService(activityRepo)
	activityHandler := activities.NewHandler(activityService, userRepo, postRepo, commentRepo)

	postService := posts.NewService(postRepo, notificationService, activityService)
	postHandler := posts.NewHandler(postService, userRepo)

	commentService := comments.NewService(commentRepo, notificationService, activityService)
	commentHandler := comments.NewHandler(commentService, userRepo)

	var uploadRepo uploads.Repository
	if deps != nil && deps.Database != nil {
		uploadRepo = uploads.NewMongoRepository(deps.Database)
	} else {
		uploadRepo = uploads.NewMemoryRepository()
	}
	uploadService := uploads.NewService(uploadRepo, cfg.PublicBaseURL, cfg.UploadDir)
	uploadHandler := uploads.NewHandler(uploadService)

	api := router.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		response.JSON(c, http.StatusOK, gin.H{"status": "ok"})
	})
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	usersGroup := api.Group("/users", optionalViewerID(cfg.JWTSecret))
	usersGroup.GET("/search", userHandler.SearchUsers)
	usersGroup.GET("/:username/profile", userHandler.GetProfile)
	usersGroup.GET("/:username/posts", postHandler.ListPostsByUsername)
	usersGroup.GET("/:username/activity-counts", activityHandler.GetCounts)

	authedUsers := api.Group("/users", auth.Middleware(cfg.JWTSecret), viewerIDFromClaims())
	authedUsers.PUT("/me/status", userHandler.UpdateStatus)
	authedUsers.POST("/:username/follow", userHandler.Follow)
	authedUsers.DELETE("/:username/follow", userHandler.Unfollow)
	authedUsers.GET("/:username/activity", activityHandler.List)

	api.GET("/posts", optionalViewerID(cfg.JWTSecret), postHandler.ListPublicTimeline)
	api.GET("/posts/:postId", optionalViewerID(cfg.JWTSecret), postHandler.GetPost)
	api.GET("/posts/:postId/likes", postHandler.ListPostLikers)
	api.GET("/posts/:postId/comments", optionalViewerID(cfg.JWTSecret), commentHandler.ListPostComments)
	api.GET("/topics/trending", postHandler.ListTrendingTopics)

	authedPosts := api.Group("/posts", auth.Middleware(cfg.JWTSecret), viewerIDFromClaims())
	authedPosts.POST("", postHandler.CreatePost)
	authedPosts.DELETE("/:postId", postHandler.DeletePost)
	authedPosts.POST("/:postId/like", postHandler.LikePost)
	authedPosts.DELETE("/:postId/like", postHandler.UnlikePost)
	authedPosts.POST("/:postId/comments", commentHandler.CreatePostComment)

	authedComments := api.Group("/comments", auth.Middleware(cfg.JWTSecret), viewerIDFromClaims())
	authedComments.POST("/:commentId/like", commentHandler.LikeComment)
	authedComments.DELETE("/:commentId/like", commentHandler.UnlikeComment)
	authedComments.POST("/:commentId/replies", commentHandler.ReplyToComment)
	authedComments.DELETE("/:commentId", commentHandler.DeleteComment)

	authedNotifications := api.Group("/notifications", auth.Middleware(cfg.JWTSecret), viewerIDFromClaims())
	authedNotifications.GET("/unread-count", notificationHandler.GetUnreadCount)
	authedNotifications.GET("", notificationHandler.ListNotifications)
	authedNotifications.GET("/:notificationId", notificationHandler.GetNotification)
	authedNotifications.POST("/read", notificationHandler.MarkRead)

	authedUploads := api.Group("/uploads", auth.Middleware(cfg.JWTSecret), viewerIDFromClaims())
	authedUploads.POST("/images", uploadHandler.UploadImage)

	authedTimeline := api.Group("/timeline", auth.Middleware(cfg.JWTSecret), viewerIDFromClaims())
	authedTimeline.GET("/following", postHandler.ListFollowingTimeline)

	return router
}

func viewerIDFromClaims() gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsAny, ok := c.Get(auth.ContextKeyClaims)
		if !ok {
			c.Next()
			return
		}

		claims, ok := claimsAny.(*auth.JWTClaims)
		if !ok || claims == nil {
			c.Next()
			return
		}

		c.Set(users.ContextKeyViewerID, claims.Subject)
		c.Next()
	}
}

func optionalViewerID(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.Next()
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer"))
		token = strings.TrimSpace(token)
		if token == "" {
			c.Next()
			return
		}

		claims, err := auth.VerifyJWT(secret, token)
		if err == nil && claims != nil {
			c.Set(users.ContextKeyViewerID, claims.Subject)
		}

		c.Next()
	}
}
