package api

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "dosage_drugs/docs" // создаётся командой swag init
	"dosage_drugs/internal/app/config"
	"dosage_drugs/internal/app/dsn"
	"dosage_drugs/internal/app/handler"
	"dosage_drugs/internal/app/minio"
	rediscli "dosage_drugs/internal/app/redis"
	"dosage_drugs/internal/app/repository"
)

func StartServer() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		logrus.Fatal("Ошибка конфигурации: ", err)
	}

	minio.Init()

	repo, err := repository.New(dsn.FromEnv())
	if err != nil {
		logrus.Fatal("Ошибка инициализации репозитория: ", err)
	}

	rds, err := rediscli.New(context.Background(), cfg.Redis)
	if err != nil {
		logrus.Fatal("Ошибка подключения к Redis: ", err)
	}
	defer rds.Close()

	h := handler.NewHandler(repo, rds, cfg)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	// ---- Swagger UI: http://localhost:8090/swagger/index.html ----
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ---- HTML-роуты (ЛР-2) ----
	r.GET("/dosage_drugs", h.GetDrug)
	r.GET("/dosage_drugs/:id", h.GetDrug)
	r.GET("/dosage_drugs/add", h.GetAddDrug)
	r.GET("/dosage_drugs/liked", h.GetLikesDrug)
	r.POST("/dosage_drugs/create", h.CreateDraft)
	r.POST("/dosage_drugs/publish", h.PublishDrug)
	r.POST("/dosage_drugs/delete", h.DeleteDrug)

	// ---- JSON API ----
	api := r.Group("/api")

	// Доступно гостю (чтение, регистрация, вход). Если токен есть — пользователь
	// попадает в контекст, чтобы заполнить is_mine / is_liked.
	public := api.Group("", h.WithOptionalAuth)
	{
		public.GET("/dosage_drugs", h.APIListDrugs)
		public.GET("/dosage_drugs/feed", h.APIFeed)
		public.POST("/users/register", h.APIRegister)
		public.POST("/users/login", h.APILogin)
	}

	// Методы создателя — только с валидным JWT, которого нет в blacklist.
	creator := api.Group("", h.WithAuthCheck)
	{
		creator.GET("/dosage_drugs/draft", h.APIGetDraft)
		creator.POST("/dosage_drugs", h.APICreateDrug)
		creator.PUT("/dosage_drugs/:id/publish", h.APIPublishDrug)
		creator.DELETE("/dosage_drugs/:id", h.APIDeleteDrug)
		creator.POST("/dosage_drugs/:id/like", h.APILike)
		creator.POST("/users/logout", h.APILogout)
	}

	logrus.Info("Server is running on :8090")
	if err := r.Run(":8090"); err != nil {
		logrus.Fatal(err)
	}
}
