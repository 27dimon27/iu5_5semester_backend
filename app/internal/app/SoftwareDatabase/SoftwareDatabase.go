package SoftwareDatabase

import (
	"errors"
	"fmt"
	"time"

	"software/internal/app/ds"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SoftwareDatabase struct {
	db *gorm.DB
}

func NewSoftwareDatabase(dsn string) (*SoftwareDatabase, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &SoftwareDatabase{
		db: db,
	}, nil
}

func (d *SoftwareDatabase) GetCoefficients() []ds.Coefficients {
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

func (d *SoftwareDatabase) SingletonGetCreator() int {
	CreatorID := 1
	return CreatorID
}

func (d *SoftwareDatabase) GetSoftwareServices() ([]ds.SoftwareService, error) {
	var services []ds.SoftwareService
	err := d.db.Find(&services).Error

	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("empty array")
	}

	return services, nil
}

func (d *SoftwareDatabase) GetSoftwareService(id int) (ds.SoftwareService, error) {
	service := ds.SoftwareService{}
	err := d.db.Where("id = ?", id).First(&service).Error
	if err != nil {
		return ds.SoftwareService{}, err
	}
	return service, nil
}

func (d *SoftwareDatabase) GetSoftwareServicesByTitle(title string) ([]ds.SoftwareService, error) {
	var services []ds.SoftwareService
	err := d.db.Where("title ILIKE ?", "%"+title+"%").Find(&services).Error
	if err != nil {
		return nil, err
	}
	return services, nil
}

func (d *SoftwareDatabase) AddNewSoftware(software ds.SoftwareService) (int, error) {
	result := d.db.Create(&software)
	if result.Error != nil {
		return 0, result.Error
	}
	return software.ID, nil
}

func (d *SoftwareDatabase) UpdateSoftware(softwareID int, software ds.SoftwareService) (int, error) {
	result := d.db.Where("id = ?", softwareID).Updates(software)
	if result.Error != nil {
		return 0, result.Error
	}
	return softwareID, nil
}

func (d *SoftwareDatabase) GetSoftwareBids(userID int, filter ds.FilterRequest) ([]ds.SoftwareBid, error) {
	var bids []ds.SoftwareBid

	query := d.db.Model(&ds.SoftwareBid{})

	if filter.StartDate != "" {
		query = query.Where("date_create >= ?", filter.StartDate)
	}

	if filter.EndDate != "" {
		query = query.Where("date_create <= ?", filter.EndDate)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	result := query.Find(&bids)
	if result.Error != nil {
		return []ds.SoftwareBid{}, result.Error
	}
	return bids, nil
}

func (d *SoftwareDatabase) UpdateActiveSoftwareBid(bidID int, bid ds.SoftwareBid) (int, error) {
	result := d.db.Where("id = ?", bidID).Updates(bid)
	if result.Error != nil {
		return 0, result.Error
	}
	return bidID, nil
}

func (d *SoftwareDatabase) CountServicesInBid(bidID int) ([]ds.Service_n_Bid, error) {
	var servicesInBid []ds.Service_n_Bid
	result := d.db.Where("bid_id = ?", bidID).Find(&servicesInBid)
	if result.Error != nil {
		return []ds.Service_n_Bid{}, result.Error
	}
	return servicesInBid, nil
}

func (d *SoftwareDatabase) FormateActiveSoftwareBid(bidID int) (int, error) {
	result := d.db.Model(&ds.SoftwareBid{}).Where("id = ?", bidID).Updates(map[string]any{
		"date_update": time.Now(),
		"status":      "сформирован",
	})
	if result.Error != nil {
		return 0, result.Error
	}
	return bidID, nil
}

func (d *SoftwareDatabase) ModerateSoftwareBid(bidID int, approved bool) (int, error) {
	newStatus := "отклонён"
	if approved {
		newStatus = "завершён"
	}
	result := d.db.Model(&ds.SoftwareBid{}).Where("id = ?", bidID).Update("status", newStatus)
	if result.Error != nil {
		return 0, result.Error
	} else if result.RowsAffected == 0 {
		return 0, errors.New("bid not found")
	}
	return bidID, nil
}

func (d *SoftwareDatabase) DeleteSoftwareBid(bidId int) (int, error) {
	result := d.db.Model(&ds.Service_n_Bid{}).Where("bid_id = ?", bidId).Delete(&ds.Service_n_Bid{})
	if result.Error != nil {
		return 0, result.Error
	}
	result = d.db.Model(&ds.SoftwareBid{}).Where("id = ?", bidId).Delete(&ds.SoftwareBid{})
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, errors.New("bid not found")
	}
	return bidId, nil
}

func (d *SoftwareDatabase) GetSoftwareServicesBid(bidID int) (ds.SoftwareBid, []ds.SoftwareService, error) {
	var bid ds.SoftwareBid
	err := d.db.Where("id = ?", bidID).First(&bid).Error
	if err != nil {
		return ds.SoftwareBid{}, []ds.SoftwareService{}, err
	}

	var service_n_bid []ds.Service_n_Bid
	err = d.db.Where("bid_id = ?", bidID).Order("index ASC").Find(&service_n_bid).Error
	if err != nil {
		return ds.SoftwareBid{}, []ds.SoftwareService{}, err
	}

	services, _ := d.GetSoftwareServices()

	servicesInUse := []ds.SoftwareService{}
	for _, bid_service := range service_n_bid {
		for _, service := range services {
			if bid_service.ServiceID == service.ID {
				servicesInUse = append(servicesInUse, service)
			}
		}
	}

	if len(servicesInUse) == 0 {
		return ds.SoftwareBid{}, []ds.SoftwareService{}, fmt.Errorf("empty bid")
	}

	return bid, servicesInUse, nil
}

func (d *SoftwareDatabase) AddSoftwareServiceToBid(serviceID int, bidID int) (bool, error) {
	var count int64
	d.db.Model(&ds.Service_n_Bid{}).Where("service_id = ? AND bid_id = ?", serviceID, bidID).Count(&count)

	if count > 0 {
		return false, errors.New("duplicate")
	}

	_, bidSize, _ := d.GetSoftwareServicesBid(bidID)
	dataToAdd := ds.Service_n_Bid{
		ServiceID: serviceID,
		BidID:     bidID,
		Count:     1,
		Index:     len(bidSize) + 1,
	}

	result := d.db.Create(&dataToAdd)
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}

func (d *SoftwareDatabase) SoftDeleteSoftwareBid(bidID int) bool {
	result := d.db.Exec(`
		UPDATE software_bids SET status = 'удалён'
		WHERE id = ?
	`, bidID)
	return result.Error == nil
}

func (d *SoftwareDatabase) FindUserActiveBid(userID int) (int, error) {
	var bid ds.SoftwareBid
	err := d.db.Where("creator_id = ? AND status = 'черновик'", userID).First(&bid).Error
	if err != nil {
		return 0, err
	}
	return int(bid.ID), nil
}

func (d *SoftwareDatabase) CreateUserActiveBid(userID int) int {
	createdBid := ds.SoftwareBid{
		Status:     "черновик",
		DateCreate: time.Now().Format("2006-01-02"),
		CreatorID:  userID,
	}
	err := d.db.Create(&createdBid).Error
	if err != nil {
		return 0
	}
	return int(createdBid.ID)
}

func (d *SoftwareDatabase) DeleteSoftware(serviceID int) (int, error) {
	result := d.db.Model(&ds.SoftwareService{}).Where("id = ?", serviceID).Delete(&ds.SoftwareService{})
	if result.Error != nil {
		return 0, result.Error
	}
	return serviceID, nil
}

func (d *SoftwareDatabase) GetSoftwareBidByID(bidID int) (ds.SoftwareBid, error) {
	var bid ds.SoftwareBid
	result := d.db.Where("id = ?", bidID).First(&bid)
	if result.Error != nil {
		return ds.SoftwareBid{}, result.Error
	}
	return bid, nil
}

func (d *SoftwareDatabase) DeleteSoftwareFromBid(bidID, serviceID int) error {
	result := d.db.Model(&ds.Service_n_Bid{}).Where("bid_id = ? AND service_id = ?", bidID, serviceID).Delete(&ds.Service_n_Bid{})
	if result.RowsAffected == 0 {
		return errors.New("software not found in current bid")
	}
	return result.Error
}

func (d *SoftwareDatabase) UpdateSoftwareInBid(bidID int, softwares []map[string]int) error {
	lenSoftwares := len(softwares)

	if lenSoftwares > 2 {
		return errors.New("too many values")
	}

	var result *gorm.DB

	switch lenSoftwares {
	case 1:
		software := softwares[0]
		result = d.db.Model(&ds.Service_n_Bid{}).Where("bid_id = ? AND service_id = ?", bidID, software["softwareID"]).Update("count", software["count"])
		if result.Error != nil {
			return result.Error
		}
	case 2:
		if softwares[0]["index"] == softwares[1]["index"] {
			return errors.New("duplicated index")
		}

		for _, software := range softwares {
			updates := map[string]any{
				"index": software["index"],
				"count": software["count"],
			}
			result = d.db.Model(&ds.Service_n_Bid{}).Where("bid_id = ? AND service_id = ?", bidID, software["softwareID"]).Updates(updates)
			if result.Error != nil {
				return result.Error
			}
		}
	}

	return nil
}

func (d *SoftwareDatabase) RegisterNewUser(userData ds.Users) error {
	result := d.db.Create(&userData)
	return result.Error
}

func (d *SoftwareDatabase) GetUserAccountData(userID int) (ds.Users, error) {
	var user ds.Users
	result := d.db.Model(&ds.Users{}).Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return ds.Users{}, result.Error
	}
	return user, nil
}

func (d *SoftwareDatabase) UpdateUserAccountData(userID int, updateUser ds.Users) error {
	result := d.db.Where("id = ?", userID).Updates(updateUser)
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return result.Error
}

func (d *SoftwareDatabase) AuthenticateUser(username, password string) (ds.Users, error) {
	var user ds.Users
	result := d.db.Where("login = ?", username).First(&user)
	if result.RowsAffected == 0 {
		return ds.Users{}, errors.New("user not found")
	}
	result = d.db.Where("password = ?", password).First(&user)
	if result.RowsAffected == 0 {
		return ds.Users{}, errors.New("incorrect password")
	}
	if result.Error != nil {
		return ds.Users{}, result.Error
	}
	return user, nil
}

func (d *SoftwareDatabase) DeauthorizeUser(userID int) error {
	var user ds.Users
	result := d.db.Where("id = ?", userID).First(&user)
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return result.Error
}
