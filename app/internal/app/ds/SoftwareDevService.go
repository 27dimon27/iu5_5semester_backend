package ds

import (
	"time"
)

type SoftwareDevService struct {
	ID          int     `gorm:"primaryKey"`
	Image       string  `gorm:"type:text;default:null"`
	Title       string  `gorm:"type:varchar(100);default:null"`
	Description string  `gorm:"type:varchar(1000);default:null"`
	Price       float32 `gorm:"type numeric(10,2)"`
	Status      bool    `gorm:"type boolean"`
	// PriceMeasure string `gorm:"type:varchar(10);default:null"`
	// Grade  string `gorm:"type:varchar(10);default:null"`
}

type SoftwareDevBid struct {
	ID          int       `gorm:"primaryKey"`
	Status      string    `gorm:"type:varchar(15);not null"`
	DateCreate  time.Time `gorm:"not null"`
	DateUpdate  time.Time `gorm:"default:null"`
	DateFinish  time.Time `gorm:"default:null"`
	CreatorID   int       `gorm:"not null"`
	ModeratorID int       `gorm:"default:null"`
}

type Service_n_Bid struct {
	ID        int `gorm:"primaryKey"`
	ServiceID int `gorm:"not null;uniqueIndex:idx_service_bid"`
	BidID     int `gorm:"not null;uniqueIndex:idx_service_bid"`
	Count     int `gorm:"not null;default:1"`
	Index     int `gorm:"not null"`

	Service SoftwareDevService `gorm:"foreignKey:ServiceID"`
	Bid     SoftwareDevBid     `gorm:"foreignKey:BidID"`
}

type Users struct {
	ID          int    `gorm:"primaryKey"`
	Login       string `gorm:"type:varchar(30);not null;unique"`
	Password    string `gorm:"type:varchar(100);not null"`
	IsModerator bool   `gorm:"default:false"`
}

type Coefficients struct {
	Level string
	Coeff float32
}

type ServiceInBid struct {
	Service SoftwareDevService
	Count   int
	Grade   string
	Sum     int
}
