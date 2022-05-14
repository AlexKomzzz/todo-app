package handler

import (
	//"todo-app/pkg/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	//service *service.Service
}

func (h *Handler) InitRoutes() *gin.Engine { // Инициализация групп функций мультиплексора
	mux := gin.New()

	auth := mux.Group("/auth") // Группа аутентификации
	{
		auth.POST("/sign-up", signUp)
		auth.POST("/sign-in", signIn)
	}

	api := mux.Group("/api") //Группа для взаимодействия с List
	{
		lists := api.Group("/lists")
		{
			lists.POST("/", createList)
			lists.GET("/", getAllLists)
			lists.GET("/:id", getListById)
			lists.PUT("/:id", updateList)
			lists.DELETE("/:id", deleteList)
		}

		items := api.Group(":id/items")
		{
			items.POST("/", createItem)
			items.GET("/", getAllItems)
			items.GET("/:item_id", getItemById)
			items.PUT("/:item_id", updateItem)
			items.DELETE("/:item_id", deleteItem)
		}
	}
	return mux
}
