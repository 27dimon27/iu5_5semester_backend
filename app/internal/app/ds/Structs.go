package ds

import "time"

type SoftwareService struct {
	ID          int     `gorm:"primaryKey" json:"-"`
	Image       string  `gorm:"type:text;default:null" json:"image,omitempty"`
	Title       string  `gorm:"type:varchar(100);not null" json:"title,omitempty"`
	Description string  `gorm:"type:varchar(1000);default:null" json:"description,omitempty"`
	Price       float32 `gorm:"type numeric(10,2)" json:"price,omitempty"`
	Status      bool    `gorm:"type boolean" json:"status,omitempty"`
}

type SoftwareBid struct {
	ID          int                             `gorm:"primaryKey" json:"-"`
	Status      string                          `gorm:"type:varchar(15);not null" json:"status,omitempty"`
	Company     string                          `gorm:"type:varchar(30);not null;default:Apple" json:"company,omitempty"`
	DateCreate  string                          `gorm:"type:date;not null" json:"dateCreate,omitempty"`
	DateUpdate  string                          `gorm:"type:date;default:null" json:"dateUpdate,omitempty"`
	DateFinish  string                          `gorm:"type:date;default:null" json:"dateFinish,omitempty"`
	CreatorID   int                             `gorm:"not null" json:"creatorID,omitempty"`
	ModeratorID int                             `gorm:"default:null" json:"moderatorID,omitempty"`
	Softwares   []SoftwareService_n_SoftwareBid `gorm:"foreignKey:SoftwareBidID" json:"softwares"`
}

type SoftwareService_n_SoftwareBid struct {
	ID                int `gorm:"primaryKey" json:"-"`
	SoftwareServiceID int `gorm:"not null;uniqueIndex:idx_software" json:"softwareID"`
	SoftwareBidID     int `gorm:"not null;uniqueIndex:idx_software" json:"-"`
	Count             int `gorm:"not null;default:1" json:"count,omitempty"`
	Index             int `gorm:"not null" json:"-"`
	Price             int `gorm:"not null" json:"price,omitempty"`

	SoftwareService SoftwareService `gorm:"foreignKey:SoftwareServiceID"`
	SoftwareBid     SoftwareBid     `gorm:"foreignKey:SoftwareBidID"`
}

type Users struct {
	ID          int    `gorm:"primaryKey" json:"-"`
	Login       string `gorm:"type:varchar(30);not null;unique" json:"login"`
	Password    string `gorm:"type:varchar(100);not null" json:"password"`
	Role        string `gorm:"type:varchar(20)" json:"role"`
	IsModerator bool   `gorm:"default:false" json:"isModerator"`
}

type Coefficients struct {
	Level string
	Coeff float32
}

type SoftwareServiceInSoftwareBid struct {
	SoftwareService SoftwareService
	Count           int
	Grade           string
	Sum             int
}

type FilterRequest struct {
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Status    string `json:"status,omitempty"`
}

type Photo struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}
