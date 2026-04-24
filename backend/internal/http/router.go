package http

import (
	"net/http"
	"strings"

	"decoy-club/backend/internal/auth"
	"decoy-club/backend/internal/common/middleware"
	"decoy-club/backend/internal/common/response"
	"decoy-club/backend/internal/config"
	"decoy-club/backend/internal/posts"
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
		postRepo = posts.NewMongoRepository(deps.Database)
	} else {
		postRepo = posts.NewMemoryRepository()
	}
	postService := posts.NewService(postRepo)
	postHandler := posts.NewHandler(postService)

	api := router.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		response.JSON(c, http.StatusOK, gin.H{"status": "ok"})
	})
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	usersGroup := api.Group("/users", optionalViewerID(cfg.JWTSecret))
	usersGroup.GET("/:username/profile", userHandler.GetProfile)

	authedUsers := api.Group("/users", auth.Middleware(cfg.JWTSecret), viewerIDFromClaims())
	authedUsers.POST("/:username/follow", userHandler.Follow)
	authedUsers.DELETE("/:username/follow", userHandler.Unfollow)

	api.GET("/posts", postHandler.ListPublicTimeline)
	api.GET("/posts/:postId", postHandler.GetPost)

	authedPosts := api.Group("/posts", auth.Middleware(cfg.JWTSecret), viewerIDFromClaims())
	authedPosts.POST("", postHandler.CreatePost)
	authedPosts.DELETE("/:postId", postHandler.DeletePost)

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
