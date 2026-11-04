package ds

type Like struct {
	ID     int `gorm:"primaryKey" json:"id"`
	DrugID int `gorm:"column:drug_id;not null;uniqueIndex:idx_user_drug" json:"drug_id"`
	UserID int `gorm:"not null;uniqueIndex:idx_user_drug" json:"user_id"`
}
