package api

import (
	"Medical_pediatric_calculator/internal/app/handler"
	"Medical_pediatric_calculator/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server...")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Fatal("Ошибка инициализации репозитория: ", err)
	}

	h := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/feed/:id", h.GetFeed)
	r.GET("/add", h.GetAddPage)
	r.GET("/grid", h.GetGrid)

	log.Println("Server is running on :8080")
	err = r.Run(":8080")
	if err != nil {
		logrus.Fatal("Ошибка запуска сервера: ", err)
	}
}
