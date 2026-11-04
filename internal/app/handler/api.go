package handler

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"dosage_drugs/internal/app/ds"
	"dosage_drugs/internal/app/minio"
)

// ---------- DTO (используются и в swagger) ----------

type DrugListResponse struct {
	Status string                  `json:"status" example:"success"`
	Data   []ds.DrugListSerializer `json:"data"`
}

type DrugFeedResponse struct {
	Status string               `json:"status" example:"success"`
	Data   ds.DrugFeedSerializer `json:"data"`
}

type DrugResponse struct {
	Status  string         `json:"status" example:"success"`
	Message string         `json:"message,omitempty" example:"опубликовано"`
	Data    ds.Dosage_drugs `json:"data"`
}

type PublishRequest struct {
	Title         string  `json:"title" example:"Парацетамол"`
	AdultDose     int     `json:"adult_dose" example:"15"`
	Concentration float64 `json:"concentration_percent" example:"2.4"`
	Description   string  `json:"description" example:"Жаропонижающее и обезболивающее средство"`
}

type LikeRequest struct {
	Value int `json:"value" example:"1" enums:"0,1"`
}

type LikeResponse struct {
	Status    string `json:"status" example:"success"`
	LikeCount int64  `json:"like_count" example:"3"`
	IsLiked   bool   `json:"is_liked" example:"true"`
}

const statusPublished = "опубликован"

// uploadMedia сохраняет файл из multipart-формы во временный файл, заливает в MinIO
// и возвращает ключ объекта (пустая строка, если файла нет или загрузка не удалась).
func uploadMedia(c *gin.Context, field string) string {
	fh, err := c.FormFile(field)
	if err != nil || fh == nil {
		return ""
	}
	name := filepath.Base(fh.Filename)
	tmp := "./tmp_" + name
	if err := c.SaveUploadedFile(fh, tmp); err != nil {
		logrus.Error("save "+field+" error: ", err)
		return ""
	}
	defer os.Remove(tmp)

	if err := minio.UploadFile(name, tmp, fh.Header.Get("Content-Type")); err != nil {
		logrus.Error("minio upload "+field+" error: ", err)
		return ""
	}
	return name
}

// ---------- Список с фильтром ----------

// APIListDrugs godoc
// @Summary      Список опубликованных препаратов
// @Description  Доступно гостю. Поиск по названию. Для авторизованного пользователя у его препаратов is_mine=true.
// @Tags         dosage_drugs
// @Produce      json
// @Param        search  query     string  false  "Поиск по названию"
// @Success      200     {object}  DrugListResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /dosage_drugs [get]
func (h *Handler) APIListDrugs(c *gin.Context) {
	search := c.Query("search")
	uid := currentUserID(c)

	drugs, err := h.Repository.ListPublishedWithFilter(search, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{"error", err.Error()})
		return
	}

	out := make([]ds.DrugListSerializer, 0, len(drugs))
	for _, d := range drugs {
		out = append(out, ds.DrugListSerializer{Dosage_drugs: d, IsMine: uid != 0 && d.CreatorID == uid})
	}
	c.JSON(http.StatusOK, DrugListResponse{Status: "success", Data: out})
}

// ---------- Лента ----------

// APIFeed godoc
// @Summary      Лента препаратов
// @Description  Без id — первый опубликованный препарат; с id — выбранный; с id и next=true — следующий после указанного (по кругу). Доступно гостю; is_liked заполняется только для авторизованного.
// @Tags         dosage_drugs
// @Produce      json
// @Param        id    query     int   false  "ID препарата"
// @Param        next  query     bool  false  "Вернуть следующий после id"
// @Success      200   {object}  DrugFeedResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /dosage_drugs/feed [get]
func (h *Handler) APIFeed(c *gin.Context) {
	uid := currentUserID(c)

	id := 0
	if s := c.Query("id"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid id"})
			return
		}
		id = v
	}

	var err error
	switch {
	case id == 0:
		id, err = h.Repository.GetFirstPublishedID()
	case c.Query("next") == "true":
		id, err = h.Repository.GetNextDrugID(id)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{"error", "нет опубликованных препаратов"})
		return
	}

	drug, err := h.Repository.GetDrugByIDCursor(id)
	if err != nil || drug == nil || drug.Status != statusPublished {
		c.JSON(http.StatusNotFound, ErrorResponse{"error", "препарат не найден"})
		return
	}

	c.JSON(http.StatusOK, DrugFeedResponse{
		Status: "success",
		Data: ds.DrugFeedSerializer{
			Dosage_drugs: *drug,
			IsLiked:      uid != 0 && h.Repository.IsLiked(drug.ID, uid),
			LikeCount:    h.Repository.GetLikesCount(drug.ID),
		},
	})
}

// ---------- Черновик ----------

// APIGetDraft godoc
// @Summary      Черновик текущего пользователя
// @Description  Возвращает незавершённый черновик препарата авторизованного создателя.
// @Tags         dosage_drugs
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  DrugResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /dosage_drugs/draft [get]
func (h *Handler) APIGetDraft(c *gin.Context) {
	draft, err := h.Repository.GetDraftByCreator(currentUserID(c))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{"error", "черновик не найден"})
		return
	}
	c.JSON(http.StatusOK, DrugResponse{Status: "success", Data: draft})
}

// ---------- Создание ----------

// APICreateDrug godoc
// @Summary      Создание черновика препарата
// @Description  Создатель загружает название, фото и видео (файлы уходят в MinIO). Поле creator_id заполняется автоматически из JWT.
// @Tags         dosage_drugs
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        title  formData  string  true   "Название препарата"
// @Param        image  formData  file    false  "Фото"
// @Param        video  formData  file    false  "Видео"
// @Success      201    {object}  DrugResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      401    {object}  ErrorResponse
// @Failure      409    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /dosage_drugs [post]
func (h *Handler) APICreateDrug(c *gin.Context) {
	uid := currentUserID(c)

	if err := c.Request.ParseMultipartForm(20 << 20); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", err.Error()})
		return
	}

	title := c.PostForm("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", "title обязателен"})
		return
	}

	if existing, err := h.Repository.GetDraftByCreator(uid); err == nil && existing.ID != 0 {
		c.JSON(http.StatusConflict, ErrorResponse{"error", "у вас уже есть черновик, опубликуйте его"})
		return
	}

	imageKey := uploadMedia(c, "image")
	videoKey := uploadMedia(c, "video")

	drug, err := h.Repository.CreateDraft(uid, title, imageKey, videoKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{"error", err.Error()})
		return
	}
	c.JSON(http.StatusCreated, DrugResponse{Status: "success", Data: *drug})
}

// ownedDrug загружает препарат и проверяет, что он принадлежит текущему пользователю.
func (h *Handler) ownedDrug(c *gin.Context, id int) (ds.Dosage_drugs, bool) {
	drug, err := h.Repository.GetDrugByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{"error", "препарат не найден"})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{"error", err.Error()})
		}
		return drug, false
	}
	if drug.CreatorID != currentUserID(c) {
		c.JSON(http.StatusForbidden, ErrorResponse{"error", "можно изменять только свои препараты"})
		return drug, false
	}
	return drug, true
}

// ---------- Публикация ----------

// APIPublishDrug godoc
// @Summary      Публикация препарата
// @Description  Заполняет дозу, концентрацию и описание и переводит препарат в статус «опубликован». Только владелец.
// @Tags         dosage_drugs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int             true  "ID препарата"
// @Param        body  body      PublishRequest  true  "Данные для публикации"
// @Success      200   {object}  DrugResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      403   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /dosage_drugs/{id}/publish [put]
func (h *Handler) APIPublishDrug(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid id"})
		return
	}

	var req PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", err.Error()})
		return
	}
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", "title не может быть пустым"})
		return
	}
	if req.AdultDose <= 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", "adult_dose должна быть положительной"})
		return
	}
	if req.Concentration < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", "concentration_percent не может быть отрицательной"})
		return
	}

	if _, ok := h.ownedDrug(c, id); !ok {
		return
	}

	if err := h.Repository.PublishDrug(id, req.Title, req.AdultDose, req.Concentration, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{"error", err.Error()})
		return
	}

	updated, err := h.Repository.GetDrugByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{"error", err.Error()})
		return
	}
	c.JSON(http.StatusOK, DrugResponse{Status: "success", Message: "опубликовано", Data: updated})
}

// ---------- Soft delete ----------

// APIDeleteDrug godoc
// @Summary      Логическое удаление препарата
// @Description  Меняет статус на «Удален». Только владелец.
// @Tags         dosage_drugs
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID препарата"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /dosage_drugs/{id} [delete]
func (h *Handler) APIDeleteDrug(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid id"})
		return
	}
	if _, ok := h.ownedDrug(c, id); !ok {
		return
	}
	if err := h.Repository.DeleteDrug(id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{"error", err.Error()})
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{Status: "success", Message: "удалено"})
}

// ---------- Лайк ----------

// APILike godoc
// @Summary      Поставить или снять лайк
// @Description  value=1 ставит лайк, value=0 снимает. Доступно авторизованному пользователю.
// @Tags         dosage_drugs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int          true  "ID препарата"
// @Param        body  body      LikeRequest  true  "value: 0 или 1"
// @Success      200   {object}  LikeResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /dosage_drugs/{id}/like [post]
func (h *Handler) APILike(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid id"})
		return
	}

	var req LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil || (req.Value != 0 && req.Value != 1) {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", "value должен быть 0 или 1"})
		return
	}

	if _, err := h.Repository.GetDrugByID(id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{"error", "препарат не найден"})
		return
	}

	uid := currentUserID(c)
	if err := h.Repository.SetLike(id, uid, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{"error", err.Error()})
		return
	}

	c.JSON(http.StatusOK, LikeResponse{
		Status:    "success",
		LikeCount: h.Repository.GetLikesCount(id),
		IsLiked:   req.Value == 1,
	})
}
