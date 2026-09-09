// internal/app/handler/handler.go
package handler

import (
	"Medical_pediatric_calculator/internal/app/repository"
	"Medical_pediatric_calculator/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

func (h *Handler) GetFeed(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Неверный ID: ", err)
		c.HTML(http.StatusOK, "index.html", gin.H{"error": "Неверный ID", "ActiveTab": "feed"})
		return
	}
	next := c.Query("next")
	if next == "true" {
		newID, err := h.Repository.GetNextServiceID(id, "true")
		if err != nil {
			logrus.Error(err)
			c.HTML(http.StatusOK, "index.html", gin.H{"error": "Услуга не найдена", "ActiveTab": "feed"})
			return
		}
		id = newID
	}
	service, err := h.Repository.GetServiceByID(id)
	if err != nil {
		logrus.Error(err)
		c.HTML(http.StatusOK, "index.html", gin.H{"error": "Услуга не найдена", "ActiveTab": "feed"})
		return
	}
	c.HTML(http.StatusOK, "index.html", gin.H{
		"service":     service,
		"likeCount":   len(service.Likes),
		"currentTime": time.Now().Format("15:04:05"),
		"ActiveTab":   "feed",
	})
}

func (h *Handler) GetAddPage(c *gin.Context) {
	draft, err := h.Repository.GetDraftService()
	if err != nil {
		logrus.Error(err)
		c.HTML(http.StatusOK, "add.html", gin.H{"error": "Черновик не найден", "ActiveTab": "add"})
		return
	}

	c.HTML(http.StatusOK, "add.html", gin.H{
		"service":   draft,
		"ActiveTab": "add",
	})
}

func (h *Handler) GetGrid(c *gin.Context) {
	searchQuery := c.Query("search")

	var services []models.Service
	if searchQuery != "" {
		services = h.Repository.SearchServices(searchQuery)
	} else {
		services = h.Repository.GetPublishedServices()
	}

	type GridItem struct {
		models.Service
		LikeCount int
	}
	var gridItems []GridItem
	for _, s := range services {
		gridItems = append(gridItems, GridItem{
			Service:   s,
			LikeCount: len(s.Likes),
		})
	}

	c.HTML(http.StatusOK, "grid.html", gin.H{
		"services":    gridItems,
		"ActiveTab":   "grid",
		"SearchQuery": searchQuery,
	})
}
