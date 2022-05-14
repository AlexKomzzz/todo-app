package main

import (
	"todo-app/pkg/handler"
)

func main() {
	mux := new(handler.Handler)
	mux.InitRoutes().Run(":8000")
}
