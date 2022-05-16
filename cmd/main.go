package main

import (
	"os"
	"todo-app/pkg/handler"
	"todo-app/pkg/repository"
	"todo-app/pkg/service"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter)) // Установка логирования в формат JSON

	if err := initConfig(); err != nil { //Инициализируем конфигурации
		logrus.Fatalf("error initializing configs: %s", err.Error())
		return
	}

	if err := godotenv.Load(); err != nil { //Загрузка переменного окружения (для передачи пароля из файла .env)
		logrus.Fatalf("error loading env variables: %s", err.Error())
		return
	}

	db, err := repository.NewPostgresDB(repository.Config{ //Инициализация БД
		Host:     viper.GetString("db.host"), // Читаем данные из файла config.yml по ключу
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"), // Читаем пароль из файла .env по ключу
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})
	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
		return
	}

	repos := repository.NewRepository(db) // Создание зависимостей
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	if err := handlers.InitRoutes().Run(viper.GetString("port")); err != nil {
		logrus.Fatalf("Error run web serv")
		return
	}
}

func initConfig() error { //Инициализация конфигураций
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
