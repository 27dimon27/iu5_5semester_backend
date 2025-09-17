package SoftwareDevServiceDatabase

import (
	"fmt"
	"time"

	"softwareDev/internal/app/ds"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SoftwareDevServiceDatabase struct {
	db *gorm.DB
}

func NewSoftwareDevServiceDatabase(dsn string) (*SoftwareDevServiceDatabase, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &SoftwareDevServiceDatabase{
		db: db,
	}, nil
}

func (d *SoftwareDevServiceDatabase) GetCoefficients() []ds.Coefficients {
	Coefficients := []ds.Coefficients{
		{
			Level: "junior",
			Coeff: 1.00,
		},
		{
			Level: "junior+",
			Coeff: 1.05,
		},
		{
			Level: "middle",
			Coeff: 1.10,
		},
		{
			Level: "middle+",
			Coeff: 1.15,
		},
		{
			Level: "senior",
			Coeff: 1.20,
		},
		{
			Level: "senior+",
			Coeff: 1.25,
		},
	}

	return Coefficients
}

func (d *SoftwareDevServiceDatabase) GetSoftwareDevServices() ([]ds.SoftwareDevService, error) {
	var services []ds.SoftwareDevService
	err := d.db.Find(&services).Error

	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("empty array")
	}

	return services, nil
}

func (d *SoftwareDevServiceDatabase) GetSoftwareDevService(id int) (ds.SoftwareDevService, error) {
	service := ds.SoftwareDevService{}
	err := d.db.Where("id = ?", id).First(&service).Error
	if err != nil {
		return ds.SoftwareDevService{}, err
	}
	return service, nil
}

func (d *SoftwareDevServiceDatabase) GetSoftwareDevServicesByTitle(title string) ([]ds.SoftwareDevService, error) {
	var services []ds.SoftwareDevService
	err := d.db.Where("title ILIKE ?", "%"+title+"%").Find(&services).Error
	if err != nil {
		return nil, err
	}
	return services, nil
}

func (d *SoftwareDevServiceDatabase) GetSoftwareDevServicesBid(bidID int) (ds.SoftwareDevBid, []ds.SoftwareDevService, error) {
	var bid ds.SoftwareDevBid
	err := d.db.Where("id = ?", bidID).First(&bid).Error
	if err != nil {
		return ds.SoftwareDevBid{}, []ds.SoftwareDevService{}, err
	}

	var service_n_bid []ds.Service_n_Bid
	err = d.db.Where("bid_id = ?", bidID).Order("index ASC").Find(&service_n_bid).Error
	if err != nil {
		return ds.SoftwareDevBid{}, []ds.SoftwareDevService{}, err
	}

	services, _ := d.GetSoftwareDevServices()

	servicesInUse := []ds.SoftwareDevService{}
	for _, bid_service := range service_n_bid {
		for _, service := range services {
			if bid_service.ServiceID == service.ID {
				servicesInUse = append(servicesInUse, service)
			}
		}
	}

	if len(servicesInUse) == 0 {
		return ds.SoftwareDevBid{}, []ds.SoftwareDevService{}, fmt.Errorf("empty bid")
	}

	return bid, servicesInUse, nil
}

func (d *SoftwareDevServiceDatabase) AddSoftwareDevServiceToBid(serviceID int, bidID int) bool {
	var count int64
	d.db.Model(&ds.Service_n_Bid{}).Where("service_id = ? AND bid_id = ?", serviceID, bidID).Count(&count)

	if count > 0 {
		return false
	}

	_, bidSize, _ := d.GetSoftwareDevServicesBid(bidID)
	dataToAdd := ds.Service_n_Bid{
		ServiceID: uint(serviceID),
		BidID:     uint(bidID),
		Count:     1,
		Index:     uint(len(bidSize)) + 1,
	}

	result := d.db.Create(&dataToAdd)
	return result.Error == nil
}

func (d *SoftwareDevServiceDatabase) DeleteSoftwareDevBid(bidID int) bool {
	result := d.db.Exec(`
		UPDATE software_dev_bids SET status = 'удалён'
		WHERE id = ?
	`, bidID)
	return result.Error == nil
}

func (d *SoftwareDevServiceDatabase) FindUserActiveBid(userID int) int {
	var bid ds.SoftwareDevBid
	err := d.db.Where("creator_id = ? AND status = 'черновик'", userID).First(&bid).Error
	if err != nil {
		return 0
	}
	return int(bid.ID)
}

func (d *SoftwareDevServiceDatabase) CreateUserActiveBid(userID int) int {
	createdBid := ds.SoftwareDevBid{
		Status:     "черновик",
		DateCreate: time.Now(),
		CreatorID:  uint(userID),
	}
	err := d.db.Create(&createdBid).Error
	if err != nil {
		return 0
	}
	return int(createdBid.ID)
}
