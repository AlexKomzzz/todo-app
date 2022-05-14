package handler

import "github.com/gin-gonic/gin"

func InitRoutes(mux *gin.Engine) {
	auth := mux.Group("/auth")
	{
		auth.POST("/sign-up")
		auth.POST("/sign-in")
	}

	api := mux.Group("/api")
	{
		lists := api.Group("/lists")
		{
			lists.POST("/")
			lists.GET("/")
			lists.GET("/:id")
			lists.PUT("/:id")
			lists.DELETE("/:id")
		}

		items := api.Group(":id/items")
		{
			items.POST("/")
			items.GET("/")
			items.GET("/:item_id")
			items.PUT("/:item_id")
			items.DELETE("/:item_id")
		}
	}
}
