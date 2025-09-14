package ds

import (
	"time"
)

type SoftwareDevService struct {
	ID           uint   `gorm:"primaryKey"`
	Image        string `gorm:"type:text;default:null"`
	Title        string `gorm:"type:text;default:null"`
	Description  string `gorm:"type:text;default:null"`
	Price        float32
	PriceMeasure string `gorm:"type:varchar(10);default:null"`
	Grade        string `gorm:"type:varchar(10);default:null"`
	Status       bool
}

type SoftwareDevBid struct {
	ID          uint      `gorm:"primaryKey"`
	Status      string    `gorm:"type:varchar(15);not null"`
	DateCreate  time.Time `gorm:"not null"`
	DateUpdate  time.Time `gorm:"default:null"`
	DateFinish  time.Time `gorm:"default:null"`
	CreatorID   uint      `gorm:"not null"`
	ModeratorID uint
}

type Service_n_Bid struct {
	ID        uint `gorm:"primaryKey"`
	ServiceID uint `gorm:"not null;uniqueIndex:idx_service_bid"`
	BidID     uint `gorm:"not null;uniqueIndex:idx_service_bid"`
	Count     uint `gorm:"not null;default:1"`
	Index     uint `gorm:"not null"`

	Service SoftwareDevService `gorm:"foreignKey:ServiceID"`
	Bid     SoftwareDevBid     `gorm:"foreignKey:BidID"`
}

type Coefficient struct {
	Level string
	Coeff float32
}
