package ds

import (
    "database/sql"
    "time"
)

type Drug struct {
    ID            int     `gorm:"primaryKey" json:"id"`
    Title         string  `gorm:"type:varchar(100);not null" json:"title"`
    Description   string  `gorm:"type:text" json:"description"`
    Status        string  `gorm:"type:varchar(20);default:'черновик'" json:"status"`
    ImageKey      string  `gorm:"type:varchar(255)" json:"image_key"`
    VideoKey      string  `gorm:"type:varchar(255)" json:"video_key"`
    AdultDose     int     `gorm:"not null;default:0" json:"adult_dose"`
    Concentration float64 `gorm:"column:concentration;not null;default:100" json:"concentration"`

    CreatorID int  `gorm:"not null;index" json:"creator_id"`
    Creator   User `gorm:"foreignKey:CreatorID" json:"creator"`

    CreatedAt time.Time    `gorm:"not null" json:"created_at"`
    FormedAt  sql.NullTime `gorm:"default:null" json:"formed_at,omitempty"`

    Likes []Like `gorm:"foreignKey:DrugID" json:"-"`
}
