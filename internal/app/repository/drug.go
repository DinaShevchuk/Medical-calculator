package repository

import (
	"errors"
	"time"

	"dosage_drugs/internal/app/ds"
)

const CurrentUserID = 1

func (r *Repository) GetFirstPublishedID() (int, error) {
	var id int
	err := r.db.
		Raw("SELECT id FROM drugs WHERE status = $1 ORDER BY id ASC LIMIT 1", "опубликован").
		Scan(&id).Error
	if err != nil || id == 0 {
		return 0, errors.New("нет опубликованных препаратов")
	}
	return id, nil
}

func (r *Repository) GetPublishedDrugs() ([]ds.Drug, error) {
	var drugs []ds.Drug
	err := r.db.Where("status = ?", "опубликован").Find(&drugs).Error
	return drugs, err
}

func (r *Repository) GetDrugByID(id int) (ds.Drug, error) {
	var drug ds.Drug
	err := r.db.
		Where("id = ? AND status != ?", id, "Удален").
		First(&drug).Error
	return drug, err
}

func (r *Repository) GetDrugByIDCursor(id int) (*ds.Drug, error) {
	query := `SELECT id, title, description, status, image_key, video_key,
	                 adult_dose, concentration, creator_id, created_at, formed_at
	          FROM drugs
	          WHERE id = $1 AND status != 'Удален'`
	row := r.db.Raw(query, id).Row()

	d := &ds.Drug{}
	err := row.Scan(
		&d.ID, &d.Title, &d.Description, &d.Status,
		&d.ImageKey, &d.VideoKey,
		&d.AdultDose, &d.Concentration,
		&d.CreatorID, &d.CreatedAt, &d.FormedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *Repository) CreateDraft(creatorID int, title, imageKey, videoKey string) (*ds.Drug, error) {
	draft := &ds.Drug{
		Title:     title,
		ImageKey:  imageKey,
		VideoKey:  videoKey,
		Status:    "черновик",
		CreatorID: creatorID,
		CreatedAt: time.Now(),
	}
	if err := r.db.Create(draft).Error; err != nil {
		return nil, err
	}
	return draft, nil
}

func (r *Repository) PublishDrug(id int, title string, adultDose int, concentrationPercent float64, description string) error {
	return r.db.Model(&ds.Drug{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"title":         title,
			"adult_dose":    adultDose,
			"Concentration": concentrationPercent,
			"description":   description,
			"status":        "опубликован",
			"formed_at":     time.Now(),
		}).Error
}

func (r *Repository) FilterDrugsByDose(minDose, maxDose int) ([]ds.Drug, error) {
	var drugs []ds.Drug
	err := r.db.
		Where("adult_dose >= ? AND adult_dose <= ? AND status = ?",
			minDose, maxDose, "опубликован").
		Find(&drugs).Error
	return drugs, err
}

func (r *Repository) GetNextDrugID(currentID int) (int, error) {
	var nextID int
	err := r.db.Raw(`
		SELECT id FROM drugs
		WHERE status = 'опубликован' AND id > $1
		ORDER BY id ASC LIMIT 1
	`, currentID).Scan(&nextID).Error
	if err != nil {
		return 0, err
	}
	if nextID != 0 {
		return nextID, nil
	}
	err = r.db.Raw(`
		SELECT id FROM drugs
		WHERE status = 'опубликован'
		ORDER BY id ASC LIMIT 1
	`).Scan(&nextID).Error
	if err != nil || nextID == 0 {
		return 0, errors.New("нет опубликованных препаратов")
	}
	return nextID, nil
}

func (r *Repository) GetLikesCount(drugID int) int64 {
	var count int64
	r.db.Model(&ds.Like{}).Where("drug_id = ?", drugID).Count(&count)
	return count
}

func (r *Repository) DeleteDrug(id int) error {
	row := r.db.Raw(
		"UPDATE drugs SET status = $1 WHERE id = $2 RETURNING id",
		"Удален", id,
	).Row()

	var updatedID int
	return row.Scan(&updatedID)
}

// Список опубликованных + фильтр + флаг «моё»
func (r *Repository) ListPublishedWithFilter(search string, currentUserID int) ([]ds.Drug, error) {
    var drugs []ds.Drug
    q := r.db.Where("status = ?", "опубликован")
    if search != "" {
        q = q.Where("title ILIKE ?", "%"+search+"%")
    }
    err := q.Order("id").Find(&drugs).Error
    return drugs, err
}

// Лента (все опубликованные)
func (r *Repository) Feed(currentUserID int) ([]ds.Drug, error) {
    var drugs []ds.Drug
    err := r.db.Where("status = ?", "опубликован").Order("id").Find(&drugs).Error
    return drugs, err
}

// Черновик текущего пользователя
func (r *Repository) GetDraftByCreator(creatorID int) (ds.Drug, error) {
    var d ds.Drug
    err := r.db.Where("status = ? AND creator_id = ?", "черновик", creatorID).First(&d).Error
    return d, err
}

// Лайк 0/1
func (r *Repository) SetLike(drugID, userID, value int) error {
    if value == 1 {
        return r.db.Create(&ds.Like{DrugID: drugID, UserID: userID}).Error
    }
    return r.db.Where("drug_id = ? AND user_id = ?", drugID, userID).
        Delete(&ds.Like{}).Error
}

// Проверить, лайкнул ли пользователь
func (r *Repository) IsLiked(drugID, userID int) bool {
    var count int64
    r.db.Model(&ds.Like{}).Where("drug_id = ? AND user_id = ?", drugID, userID).Count(&count)
    return count > 0
}

// Список ID, которые лайкнул пользователь — для пометки в ленте
func (r *Repository) LikedDrugIDs(userID int) ([]int, error) {
    var ids []int
    err := r.db.Model(&ds.Like{}).Where("user_id = ?", userID).Pluck("drug_id", &ids).Error
    return ids, err
}
