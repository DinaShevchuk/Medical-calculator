package handler

import (
    "net/http"
    "os"
    "strconv"

    "dosage_drugs/internal/app/ds"
    "dosage_drugs/internal/app/minio"
    "dosage_drugs/internal/app/repository"

    "github.com/gin-gonic/gin"
)

// ---------- Список с фильтром ----------
func (h *Handler) APIListDrugs(c *gin.Context) {
    search := c.Query("search")
    uid := repository.CurrentUser().ID

    drugs, err := h.Repository.ListPublishedWithFilter(search, uid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
        return
    }

    out := make([]ds.DrugListSerializer, 0, len(drugs))
    for _, d := range drugs {
        out = append(out, ds.DrugListSerializer{Drug: d, IsMine: d.CreatorID == uid})
    }

    c.JSON(http.StatusOK, gin.H{"status": "success", "data": out})
}

// ---------- Лента ----------
func (h *Handler) APIFeed(c *gin.Context) {
    uid := repository.CurrentUser().ID

    drugs, err := h.Repository.Feed(uid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
        return
    }

    liked, _ := h.Repository.LikedDrugIDs(uid)
    likedSet := make(map[int]bool, len(liked))
    for _, id := range liked {
        likedSet[id] = true
    }

    out := make([]ds.DrugFeedSerializer, 0, len(drugs))
    for _, d := range drugs {
        out = append(out, ds.DrugFeedSerializer{
            Drug:      d,
            IsLiked:   likedSet[d.ID],
            LikeCount: h.Repository.GetLikesCount(d.ID),
        })
    }

    c.JSON(http.StatusOK, gin.H{"status": "success", "data": out})
}

// ---------- Черновик ----------
func (h *Handler) APIGetDraft(c *gin.Context) {
    uid := repository.CurrentUser().ID
    draft, err := h.Repository.GetDraftByCreator(uid)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "черновик не найден"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "success", "data": draft})
}

// ---------- Создание ----------
func (h *Handler) APICreateDrug(c *gin.Context) {
    if err := c.Request.ParseMultipartForm(20 << 20); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
        return
    }

    title := c.PostForm("title")
    if title == "" {
        c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "title обязателен"})
        return
    }

    imageKey := ""
    videoKey := ""

    if fh, err := c.FormFile("image"); err == nil {
        tmp := "./tmp_" + fh.Filename
        if c.SaveUploadedFile(fh, tmp) == nil {
            _ = minio.UploadFile(fh.Filename, tmp, fh.Header.Get("Content-Type"))
            _ = os.Remove(tmp)
            imageKey = fh.Filename
        }
    }
    if fh, err := c.FormFile("video"); err == nil {
        tmp := "./tmp_" + fh.Filename
        if c.SaveUploadedFile(fh, tmp) == nil {
            _ = minio.UploadFile(fh.Filename, tmp, fh.Header.Get("Content-Type"))
			_ = os.Remove(tmp)
            videoKey = fh.Filename
        }
    }

    uid := repository.CurrentUser().ID
    drug, err := h.Repository.CreateDraft(uid, title, imageKey, videoKey)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"status": "success", "data": drug})
}

// ---------- Публикация ----------
func (h *Handler) APIPublishDrug(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid id"})
        return
    }

    var req struct {
        Title         string  `json:"title"`
        AdultDose     int     `json:"adult_dose"`
        Concentration float64 `json:"concentration_percent"`
        Description   string  `json:"description"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
        return
    }

    if err := h.Repository.PublishDrug(id, req.Title, req.AdultDose, req.Concentration, req.Description); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"status": "success", "message": "опубликовано"})
}

// ---------- Soft delete ----------
func (h *Handler) APIDeleteDrug(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid id"})
        return
    }
    if err := h.Repository.DeleteDrug(id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "success", "message": "удалено"})
}

// ---------- Лайк ----------
func (h *Handler) APILike(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid id"})
        return
    }
    var req struct {
        Value int `json:"value"`
    }
    if err := c.ShouldBindJSON(&req); err != nil || (req.Value != 0 && req.Value != 1) {
        c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "value должен быть 0 или 1"})
        return
    }
    uid := repository.CurrentUser().ID
    if err := h.Repository.SetLike(id, uid, req.Value); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// ---------- Пользователи ----------
func (h *Handler) APIRegister(c *gin.Context) {
    var req struct {
        Login    string `json:"login"`
        Password string `json:"password"`
    }
    if err := c.ShouldBindJSON(&req); err != nil || req.Login == "" || req.Password == "" {
        c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "login и password обязательны"})
        return
    }
    // Заглушка: создаём пользователя, если не существует
    user, err := h.Repository.RegisterUser(req.Login, req.Password)
    if err != nil {
        c.JSON(http.StatusConflict, gin.H{"status": "error", "message": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"status": "success", "data": user})
}

func (h *Handler) APILogin(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "success", "message": "заглушка логина"})
}

func (h *Handler) APILogout(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "success", "message": "заглушка логаута"})
}
