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
    r.GET("/dosage_drugs", h.GetDrug)
    r.GET("/dosage_drugs/:id", h.GetDrug)
    r.GET("/dosage_drugs/add", h.GetAddDrug)
    r.GET("/dosage_drugs/liked", h.GetLikesDrug)
    r.POST("/dosage_drugs/create", h.CreateDraft)
    r.POST("/dosage_drugs/publish", h.PublishDrug)
    r.POST("/dosage_drugs/delete", h.DeleteDrug)

    // ---- JSON API (ЛР-3) ----
    api := r.Group("/api")
    {
        api.GET("/dosage_drugs", h.APIListDrugs)
        api.GET("/dosage_drugs/feed", h.APIFeed)
        api.GET("/dosage_drugs/draft", h.APIGetDraft)
        api.POST("/dosage_drugs", h.APICreateDrug)
        api.PUT("/dosage_drugs/:id/publish", h.APIPublishDrug)
        api.DELETE("/dosage_drugs/:id", h.APIDeleteDrug)
        api.POST("/dosage_drugs/:id/like", h.APILike)

        api.POST("/users/register", h.APIRegister)
        api.POST("/users/login", h.APILogin)
        api.POST("/users/logout", h.APILogout)
    }

    logrus.Info("Server is running on :8090")
    if err := r.Run(":8090"); err != nil {
        logrus.Fatal(err)
    }
}
