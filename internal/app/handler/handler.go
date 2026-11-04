package handler

import (
	"dosage_drugs/internal/app/config"
	rediscli "dosage_drugs/internal/app/redis"
	"dosage_drugs/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *rediscli.Client
	Config     *config.Config
}

func NewHandler(r *repository.Repository, rc *rediscli.Client, cfg *config.Config) *Handler {
	return &Handler{Repository: r, Redis: rc, Config: cfg}
}
