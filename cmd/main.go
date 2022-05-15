package main

import (
	"log"
	"todo-app/pkg/handler"
	"todo-app/pkg/repository"
	"todo-app/pkg/service"

	"github.com/spf13/viper"
)

func main() {
	if err := initConfig(); err != nil {
		log.Fatalf("error initializing configs: %s", err.Error())
	}

	repos := repository.NewRepository() // Создание зависимостей
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	handlers = new(handler.Handler)
	if err := handlers.InitRoutes().Run(viper.GetString("port")); err != nil {
		log.Fatalf("Error run web serv")
	}
}

func initConfig() error { //Инициализация конфигураций
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
