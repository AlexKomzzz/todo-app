package handler

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"todo-app/pkg/service"

	_ "todo-app/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

//go:embed web/assets/* web/templates/*
var f embed.FS

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) InitRoutes() (*gin.Engine, error) { // Инициализация групп функций мультиплексора

	//gin.SetMode(gin.ReleaseMode) // Переключение сервера в режим Релиза из режима Отладка

	mux := gin.New()

	mux.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler)) // Для работы сваггера

	// Следующий блок кода отвечает за загрузку шаблонов html и css из директории FS
	templ := template.Must(template.New("").ParseFS(f, "web/templates/*.html"))
	fsys, err := fs.Sub(f, "web/assets")
	if err != nil {
		return mux, err
	}
	mux.StaticFS("/assets", http.FS(fsys))
	mux.SetHTMLTemplate(templ)

	mux.NoRoute(Response404) // При неверном URL вызывает ф-ю Response404

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
	return mux, nil
}
