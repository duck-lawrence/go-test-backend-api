package initialization

import (
	"time"

	"github.com/ducklawrence05/go-test-backend-api/config"
	"github.com/ducklawrence05/go-test-backend-api/internal/controller/http/middleware"
	"github.com/ducklawrence05/go-test-backend-api/internal/controller/http/v1/router"
	"github.com/ducklawrence05/go-test-backend-api/internal/controller/http/v1/router/user"
	managerWire "github.com/ducklawrence05/go-test-backend-api/internal/infrastructure/wire/managers"
	"github.com/ducklawrence05/go-test-backend-api/pkg/logger"
	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	Config *config.Config
	Logger logger.Interface
}

func InitRouter(routerCfg *RouterConfig, managers *managerWire.ManagerSet) *gin.Engine {
	r := gin.Default()

	// 1 req/second, max 5 burst
	r.Use(middleware.RateLimitMiddleware(1, 5))
	middleware.StartCleanupJob(5*time.Minute, 1*time.Minute)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	userRouter := router.RouterGroupApp.User

	MainGroup := r.Group("/v1")
	{
		userRouter.NewUserRouter(
			MainGroup,
			&user.UserRouterConfig{
				Config: routerCfg.Config,
				Logger: routerCfg.Logger,
			},
			managerWire.ProvideUserManagerSet(managers),
		)
	}

	return r
}
