package api

import (
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "github.com/sirupsen/logrus"

    "dosage_drugs/internal/app/dsn"
    "dosage_drugs/internal/app/handler"
    "dosage_drugs/internal/app/minio"
    "dosage_drugs/internal/app/repository"
)

func StartServer() {
    _ = godotenv.Load()
    minio.Init()

    repo, err := repository.New(dsn.FromEnv())
    if err != nil {
        logrus.Fatal("Ошибка инициализации репозитория: ", err)
    }

    h := handler.NewHandler(repo)

    r := gin.Default()
    r.LoadHTMLGlob("templates/*")
    r.Static("/static", "./resources")

    // ---- HTML-роуты (ЛР-2) ----
    r.GET("/drugs", h.GetDrug)
    r.GET("/drugs/:id", h.GetDrug)
    r.GET("/drugs/add", h.GetAddDrug)
    r.GET("/drugs/liked", h.GetLikesDrug)
    r.POST("/drugs/create", h.CreateDraft)
    r.POST("/drugs/publish", h.PublishDrug)
    r.POST("/drugs/delete", h.DeleteDrug)

    // ---- JSON API (ЛР-3) ----
    api := r.Group("/api")
    {
        api.GET("/drugs", h.APIListDrugs)
        api.GET("/drugs/feed", h.APIFeed)
        api.GET("/drugs/draft", h.APIGetDraft)
        api.POST("/drugs", h.APICreateDrug)
        api.PUT("/drugs/:id/publish", h.APIPublishDrug)
        api.DELETE("/drugs/:id", h.APIDeleteDrug)
        api.POST("/drugs/:id/like", h.APILike)

        api.POST("/users/register", h.APIRegister)
        api.POST("/users/login", h.APILogin)
        api.POST("/users/logout", h.APILogout)
    }

    logrus.Info("Server is running on :8090")
    if err := r.Run(":8090"); err != nil {
        logrus.Fatal(err)
    }
}
