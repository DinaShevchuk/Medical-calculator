package handler

import (
    "fmt"
    "net/http"
    "os"                          // ← нужно
    "strconv"

    "dosage_drugs/internal/app/ds"
    "dosage_drugs/internal/app/minio"     // ← нужно
    "dosage_drugs/internal/app/repository"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
)

func (h *Handler) GetDrug(c *gin.Context) {
	idStr := c.Param("id")

	if c.Query("next") == "true" {
		currStr := c.Query("id")
		if currStr == "" {
			currStr = idStr
		}
		currID, _ := strconv.Atoi(currStr)
		if currID != 0 {
			newID, err := h.Repository.GetNextDrugID(currID)
			if err != nil {
				logrus.Error(err)
				c.HTML(http.StatusOK, "drug_info.html", gin.H{
					"error":     "Препарат не найден",
					"ActiveTab": "drug_info",
				})
				return
			}
			c.Redirect(http.StatusFound, fmt.Sprintf("/drugs/%d", newID))
			return
		}
	}

	if idStr == "" {
		firstID, err := h.Repository.GetFirstPublishedID()
		if err != nil || firstID == 0 {
			c.HTML(http.StatusOK, "drug_info.html", gin.H{
				"error":     "Нет опубликованных препаратов",
				"ActiveTab": "drug_info",
			})
			return
		}
		c.Redirect(http.StatusFound, fmt.Sprintf("/drugs/%d", firstID))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.HTML(http.StatusOK, "drug_info.html", gin.H{
			"error":     "Неверный ID",
			"ActiveTab": "drug_info",
		})
		return
	}

	drug, err := h.Repository.GetDrugByIDCursor(id)
	if err != nil || drug == nil {
		firstID, ferr := h.Repository.GetFirstPublishedID()
		if ferr == nil && firstID > 0 && firstID != id {
			c.Redirect(http.StatusFound, fmt.Sprintf("/drugs/%d", firstID))
			return
		}
		c.HTML(http.StatusOK, "drug_info.html", gin.H{
			"error":     "Препарат не найден",
			"ActiveTab": "drug_info",
		})
		return
	}

	c.HTML(http.StatusOK, "drug_info.html", gin.H{
		"drug":      drug,
		"likeCount": h.Repository.GetLikesCount(drug.ID),
		"ActiveTab": "drug_info",
	})
}

// GET /drugs/add
func (h *Handler) GetAddDrug(c *gin.Context) {
	stepStr := c.Query("step")

	draft, err := h.Repository.GetDraftByCreator(repository.CurrentUserID)

	if err != nil {
		c.HTML(http.StatusOK, "add_drug.html", gin.H{
			"ActiveTab": "AddDrug",
			"Step":      1,
			"drug":      ds.Drug{},
		})
		return
	}

	step := 2
	if stepStr == "1" {
		step = 1
	}

	c.HTML(http.StatusOK, "add_drug.html", gin.H{
		"drug":      draft,
		"ActiveTab": "AddDrug",
		"Step":      step,
	})
}

// POST /drugs/create
func (h *Handler) CreateDraft(c *gin.Context) {
    if existing, err := h.Repository.GetDraftByCreator(repository.CurrentUserID); err == nil && existing.ID != 0 {
        c.Redirect(http.StatusFound, "/drugs/add?step=2")
        return
    }

    title := c.PostForm("title")
    if title == "" {
        c.HTML(http.StatusOK, "add_drug.html", gin.H{
            "error":     "Название не может быть пустым",
            "ActiveTab": "add_drug",
            "Step":      1,
            "drug":      ds.Drug{},
        })
        return
    }

    imageKey := ""
    videoKey := ""

    // Фото
    if fh, err := c.FormFile("image"); err == nil && fh != nil {
        tmp := "./tmp_" + fh.Filename
        if err := c.SaveUploadedFile(fh, tmp); err == nil {
			if err := minio.UploadFile(fh.Filename, tmp, fh.Header.Get("Content-Type")); err != nil {
				logrus.Error("minio upload image error:", err)
			} else {
				imageKey = fh.Filename
			}
			_ = os.Remove(tmp)
		} else {
			logrus.Error("save image error:", err)
		}
    }

    // Видео
    if fh, err := c.FormFile("video"); err == nil && fh != nil {
        tmp := "./tmp_" + fh.Filename
        if err := c.SaveUploadedFile(fh, tmp); err == nil {
            _ = minio.UploadFile(fh.Filename, tmp, fh.Header.Get("Content-Type"))
            _ = os.Remove(tmp)
            videoKey = fh.Filename
        } else {
            logrus.Error("save video error:", err)
        }
    }

    _, err := h.Repository.CreateDraft(repository.CurrentUserID, title, imageKey, videoKey)
    if err != nil {
        logrus.Error(err)
        c.HTML(http.StatusOK, "add_drug.html", gin.H{
            "error":     "Ошибка создания черновика: " + err.Error(),
            "ActiveTab": "add_drug",
            "Step":      1,
            "drug":      ds.Drug{Title: title},
        })
        return
    }

    c.Redirect(http.StatusFound, "/drugs/add?step=2")
}
// POST /drugs/publish
func (h *Handler) PublishDrug(c *gin.Context) {
	id, err := strconv.Atoi(c.PostForm("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/drugs/add")
		return
	}

	title := c.PostForm("title")
	adultDose, errDose := strconv.Atoi(c.PostForm("adult_dose"))
	concentration, errConc := strconv.ParseFloat(c.PostForm("concentration_percent"), 64)
	description := c.PostForm("description")

	renderErr := func(msg string) {
		c.HTML(http.StatusOK, "add_drug.html", gin.H{
			"error":     msg,
			"ActiveTab": "AddDrug",
			"Step":      2,
			"drug": ds.Drug{
				ID:            id,
				Title:         title,
				AdultDose:     adultDose,
				Concentration: concentration,
				Description:   description,
			},
		})
	}

	if title == "" {
		renderErr("Название не может быть пустым")
		return
	}
	if errDose != nil || adultDose <= 0 {
		renderErr("Доза должна быть положительным числом")
		return
	}
	if errConc != nil || concentration < 0 {
		renderErr("Концентрация должна быть неотрицательным числом")
		return
	}

	if err := h.Repository.PublishDrug(id, title, adultDose, concentration, description); err != nil {
		logrus.Error(err)
		renderErr("Ошибка публикации: " + err.Error())
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/drugs/%d", id))
}

// POST /drugs/delete
func (h *Handler) DeleteDrug(c *gin.Context) {
	id, err := strconv.Atoi(c.PostForm("drug_id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/drugs/liked")
		return
	}

	if err := h.Repository.DeleteDrug(id); err != nil {
		logrus.Error(err)
	}
	c.Redirect(http.StatusFound, "/drugs/liked")
}

// GET /drugs/liked
func (h *Handler) GetLikesDrug(c *gin.Context) {
	minStr := c.Query("min")
	maxStr := c.Query("max")

	minDose, maxDose := 0, 100
	if v, err := strconv.Atoi(minStr); err == nil {
		minDose = v
	}
	if v, err := strconv.Atoi(maxStr); err == nil {
		maxDose = v
	}

	if minDose < 0 {
		minDose = 0
	}
	if maxDose > 100 {
		maxDose = 100
	}
	if minDose > maxDose {
		minDose, maxDose = maxDose, minDose
	}

	drugs, err := h.Repository.FilterDrugsByDose(minDose, maxDose)
	if err != nil {
		logrus.Error(err)
		c.HTML(http.StatusOK, "likes_drug.html", gin.H{
			"error":     "Ошибка загрузки",
			"ActiveTab": "LikesDrug",
		})
		return
	}

	type drugItem struct {
		ds.Drug
		LikeCount int64
	}
	var items []drugItem
	for _, d := range drugs {
		items = append(items, drugItem{
			Drug:      d,
			LikeCount: h.Repository.GetLikesCount(d.ID),
		})
	}

	c.HTML(http.StatusOK, "likes_drug.html", gin.H{
		"drugs":     items,
		"ActiveTab": "LikesDrug",
		"MinDose":   minDose,
		"MaxDose":   maxDose,
	})
}
