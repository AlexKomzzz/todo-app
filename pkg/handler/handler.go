package handler

import (
	"context"
	"todo-app/pkg/service"

	_ "todo-app/docs"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

type Handler struct {
	services    *service.Service
	ctx         context.Context
	redisClient *redis.Client
}

func NewHandler(services *service.Service, ctx context.Context, redisClient *redis.Client) *Handler {
	return &Handler{
		services:    services,
		ctx:         ctx,
		redisClient: redisClient,
	}
}

func (h *Handler) InitRoutes() *gin.Engine { // Инициализация групп функций мультиплексора
	mux := gin.New()

	mux.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	auth := mux.Group("/auth") // Группа аутентификации
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
	}

	api := mux.Group("/api", h.userIdentity) //Группа для взаимодействия с List
	{
		lists := api.Group("/lists")
		{
			lists.POST("/", h.createList)
			lists.GET("/", h.getAllLists)
			lists.GET("/:id", h.getListById)
			lists.PUT("/:id", h.updateList)
			lists.DELETE("/:id", h.deleteList)

			items := lists.Group(":id/items")
			{
				items.POST("/", h.createItem)
				items.GET("/", h.getAllItems)
			}
		}

		items := api.Group("items")
		{
			items.GET("/:id", h.getItemById)
			items.PUT("/:id", h.updateItem)
			items.DELETE("/:id", h.deleteItem)
		}
	}
	return mux
}
