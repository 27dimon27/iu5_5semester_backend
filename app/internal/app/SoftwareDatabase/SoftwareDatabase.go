package SoftwareDatabase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"time"

	"software/internal/app/config"
	"software/internal/app/ds"
	"software/internal/app/redis"

	"github.com/minio/minio-go/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SoftwareDatabase struct {
	DB          *gorm.DB
	Client      *minio.Client
	RedisClient *redis.Client
}

func NewSoftwareAndPhotoDatabase(dsn string) (*SoftwareDatabase, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	client, err := config.NewMinioClient()
	if err != nil {
		return nil, err
	}

	redisClient, err := redis.New(context.Background(), config.RedisClientConfig)
	if err != nil {
		return nil, err
	}

	return &SoftwareDatabase{
		DB:          db,
		Client:      client,
		RedisClient: redisClient,
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

func (d *SoftwareDatabase) AddPhotoToSoftwareService(file *multipart.FileHeader, photoName string, softwareID int) (*ds.Photo, error) {
	var software ds.SoftwareService
	result := d.DB.Where("id = ?", softwareID).First(&software)
	if result.RowsAffected == 0 {
		return nil, errors.New("software not found")
	}
	if result.Error != nil {
		return nil, result.Error
	}

	ctx := context.Background()

	if software.Image != "" {
		err := d.Client.RemoveObject(ctx, config.MinioClientConfig.Bucket, software.Image, minio.RemoveObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to remove existing file: %v", err)
		}
		photoName = software.Image
	}

	exists, err := d.checkObjectExists(photoName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if object exists: %v", err)
	}
	if exists {
		return nil, fmt.Errorf("photo with name '%s' already exists", photoName)
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = d.Client.PutObject(
		context.Background(),
		config.MinioClientConfig.Bucket,
		photoName,
		src,
		file.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to MinIO: %v", err)
	}

	uploaded, err := d.verifyUpload(photoName, file.Size)
	if err != nil {
		del_err := d.DeletePhotoFromMinio(photoName)
		if del_err != nil {
			err = del_err
		}
		return nil, fmt.Errorf("upload verification failed: %v", err)
	}

	if !uploaded {
		del_err := d.DeletePhotoFromMinio(photoName)
		if del_err != nil {
			return nil, del_err
		}
		return nil, fmt.Errorf("file was not uploaded successfully")
	}

	log.Printf("Successfully uploaded photo")

	url := fmt.Sprintf("http://%s/%s/%s", config.MinioClientConfig.Endpoint, config.MinioClientConfig.Bucket, photoName)

	photo := &ds.Photo{
		Name:      photoName,
		Size:      file.Size,
		URL:       url,
		CreatedAt: time.Now(),
	}

	if software.Image == "" {
		result = d.DB.Model(&ds.SoftwareService{}).Where("id = ?", softwareID).Update("image", photoName)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	return photo, nil
}

func (d *SoftwareDatabase) checkObjectExists(objectName string) (bool, error) {
	_, err := d.Client.StatObject(
		context.Background(),
		config.MinioClientConfig.Bucket,
		objectName,
		minio.StatObjectOptions{},
	)

	if err != nil {
		if minioErr, ok := err.(minio.ErrorResponse); ok {
			if minioErr.Code == "NoSuchKey" {
				return false, nil
			}
		}
		return false, err
	}

	return true, nil
}

func (d *SoftwareDatabase) verifyUpload(objectName string, expectedSize int64) (bool, error) {
	info, err := d.Client.StatObject(
		context.Background(),
		config.MinioClientConfig.Bucket,
		objectName,
		minio.StatObjectOptions{},
	)

	if err != nil {
		return false, err
	}

	if info.Size != expectedSize {
		return false, fmt.Errorf("size mismatch: expected %d, got %d", expectedSize, info.Size)
	}

	if _, exists := info.UserMetadata["X-Amz-Meta-Incomplete"]; exists {
		return false, fmt.Errorf("object is marked as incomplete")
	}

	return true, nil
}

func (d *SoftwareDatabase) DeletePhotoFromMinio(objectName string) error {
	err := d.Client.RemoveObject(
		context.Background(),
		config.MinioClientConfig.Bucket,
		objectName,
		minio.RemoveObjectOptions{},
	)
	return err
}

func (d *SoftwareDatabase) GetSoftwareServices() ([]ds.SoftwareService, error) {
	var softwares []ds.SoftwareService
	err := d.DB.Where("status = ?", true).Find(&softwares).Error

	if err != nil {
		return nil, err
	}
	if len(softwares) == 0 {
		return nil, fmt.Errorf("empty array")
	}

	return softwares, nil
}

func (d *SoftwareDatabase) GetSoftwareService(id int) (ds.SoftwareService, error) {
	software := ds.SoftwareService{}
	err := d.DB.Where("id = ?", id).First(&software).Error
	if err != nil {
		return ds.SoftwareService{}, err
	}
	return software, nil
}

func (d *SoftwareDatabase) GetSoftwareServicesByTitle(title string) ([]ds.SoftwareService, error) {
	var softwares []ds.SoftwareService
	err := d.DB.Where("status = ? AND title ILIKE ?", true, "%"+title+"%").Find(&softwares).Error
	if err != nil {
		return nil, err
	}
	return softwares, nil
}

func (d *SoftwareDatabase) AddNewSoftware(software ds.SoftwareService) error {
	result := d.DB.Create(&software)
	return result.Error
}

func (d *SoftwareDatabase) UpdateSoftware(softwareID int, software ds.SoftwareService) (ds.SoftwareService, error) {
	result := d.DB.Model(&ds.SoftwareService{}).Where("id = ?", softwareID).Updates(software).First(&software)
	if result.Error != nil {
		return ds.SoftwareService{}, result.Error
	}
	return software, nil
}

func (d *SoftwareDatabase) GetSoftwareBids(userID int, filter ds.FilterRequest) ([]ds.SoftwareBid, error) {
	var bids []ds.SoftwareBid

	query := d.DB.Model(&ds.SoftwareBid{})
	query = query.Where("status NOT IN ('черновик', 'удалён')")

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

func (d *SoftwareDatabase) UpdateActiveSoftwareBid(bidID int, bid ds.SoftwareBid) (ds.SoftwareBid, error) {
	result := d.DB.Where("id = ?", bidID).Updates(bid).First(&bid)
	if result.Error != nil {
		return ds.SoftwareBid{}, result.Error
	}
	return bid, nil
}

func (d *SoftwareDatabase) CountServicesInBid(bidID int) ([]ds.SoftwareService_n_SoftwareBid, error) {
	var softwaresInBid []ds.SoftwareService_n_SoftwareBid
	result := d.DB.Where("software_bid_id = ?", bidID).Find(&softwaresInBid)
	if result.Error != nil {
		return []ds.SoftwareService_n_SoftwareBid{}, result.Error
	}
	return softwaresInBid, nil
}

func (d *SoftwareDatabase) FormateActiveSoftwareBid(bidID int) (int, error) {
	result := d.DB.Model(&ds.SoftwareBid{}).Where("id = ?", bidID).Updates(map[string]any{
		"date_update": time.Now(),
		"status":      "сформирован",
	})
	if result.Error != nil {
		return 0, result.Error
	}
	return bidID, nil
}

func (d *SoftwareDatabase) ModerateSoftwareBid(bidID int, approved bool) (int, int, error) {
	newStatus := "отклонён"
	if approved {
		newStatus = "завершён"
	}

	var bid ds.SoftwareBid

	result := d.DB.Model(&ds.SoftwareBid{}).Where("id = ?", bidID).Update("status", newStatus)
	if result.Error != nil {
		return 0, 0, result.Error
	} else if result.RowsAffected == 0 {
		return 0, 0, errors.New("bid not found")
	}

	result = d.DB.Preload("Softwares.SoftwareService").First(&bid)
	if result.Error != nil {
		return 0, 0, result.Error
	}

	var sum int
	softwares := bid.Softwares
	for _, item := range softwares {
		sum += item.Price
	}

	return bidID, sum, nil
}

func (d *SoftwareDatabase) DeleteSoftwareBid(bidId int) (int, error) {
	// result := d.DB.Model(&ds.SoftwareService_n_SoftwareBid{}).Where("software_bid_id = ?", bidId).Delete(&ds.SoftwareService_n_SoftwareBid{})
	// if result.Error != nil {
	// 	return 0, result.Error
	// }
	result := d.DB.Model(&ds.SoftwareBid{}).Where("id = ?", bidId).Update("status", "удалён")
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, errors.New("bid not found")
	}
	return bidId, nil
}

func (d *SoftwareDatabase) GetSoftwareServicesBid(bidID int) (ds.SoftwareBid, []ds.SoftwareService, error) {
	var softwareBid ds.SoftwareBid
	err := d.DB.Where("id = ?", bidID).First(&softwareBid).Error
	if err != nil {
		return ds.SoftwareBid{}, []ds.SoftwareService{}, err
	}

	var software_n_bid []ds.SoftwareService_n_SoftwareBid
	err = d.DB.Where("software_bid_id = ?", bidID).Order("index ASC").Find(&software_n_bid).Error
	if err != nil {
		return ds.SoftwareBid{}, []ds.SoftwareService{}, err
	}

	softwares, _ := d.GetSoftwareServices()

	softwaresInUse := []ds.SoftwareService{}
	for _, bid_sofware := range software_n_bid {
		for _, software := range softwares {
			if bid_sofware.SoftwareServiceID == software.ID {
				softwaresInUse = append(softwaresInUse, software)
			}
		}
	}

	if len(softwaresInUse) == 0 {
		return ds.SoftwareBid{}, []ds.SoftwareService{}, fmt.Errorf("empty bid")
	}

	return softwareBid, softwaresInUse, nil
}

func (d *SoftwareDatabase) AddSoftwareServiceToBid(softwareID int, bidID int) (bool, error) {
	var count int64
	d.DB.Model(&ds.SoftwareService_n_SoftwareBid{}).Where("software_service_id = ? AND software_bid_id = ?", softwareID, bidID).Count(&count)

	if count > 0 {
		return false, errors.New("duplicate")
	}

	var software ds.SoftwareService
	d.DB.Model(&ds.SoftwareService{}).Where("id = ?", softwareID).First(&software)

	_, bidSize, _ := d.GetSoftwareServicesBid(bidID)
	dataToAdd := ds.SoftwareService_n_SoftwareBid{
		SoftwareServiceID: softwareID,
		SoftwareBidID:     bidID,
		Count:             1,
		Index:             len(bidSize) + 1,
		Price:             int(software.Price),
	}

	result := d.DB.Create(&dataToAdd)
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}

func (d *SoftwareDatabase) SoftDeleteSoftwareBid(bidID int) bool {
	result := d.DB.Exec(`
		UPDATE software_bids SET status = 'удалён'
		WHERE id = ?
	`, bidID)
	return result.Error == nil
}

func (d *SoftwareDatabase) FindUserActiveBid(userID int) (int, error) {
	var bid ds.SoftwareBid
	err := d.DB.Where("creator_id = ? AND status = 'черновик'", userID).First(&bid).Error
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
	err := d.DB.Create(&createdBid).Error
	if err != nil {
		return 0
	}
	return int(createdBid.ID)
}

func (d *SoftwareDatabase) DeleteSoftwareWithPhoto(SoftwareID int) (int, error) {
	var software ds.SoftwareService
	result := d.DB.Where("id = ?", SoftwareID).First(&software)
	if result.RowsAffected == 0 {
		return 0, errors.New("software not found")
	}
	if result.Error != nil {
		return 0, result.Error
	}

	result = d.DB.Model(&ds.SoftwareService{}).Where("id = ?", SoftwareID).Update("status", false).Update("image", "")
	if result.Error != nil {
		return 0, result.Error
	}
	err := d.DeletePhotoFromMinio(software.Image)
	if err != nil {
		return 0, err
	}
	return SoftwareID, nil
}

func (d *SoftwareDatabase) GetSoftwareBidByID(bidID int) (ds.SoftwareBid, error) {
	var bid ds.SoftwareBid
	result := d.DB.Preload("Softwares.SoftwareService").First(&bid, bidID)
	if result.Error != nil {
		return ds.SoftwareBid{}, result.Error
	}
	return bid, nil
}

func (d *SoftwareDatabase) DeleteSoftwareFromBid(bidID, SoftwareID int) error {
	result := d.DB.Model(&ds.SoftwareService_n_SoftwareBid{}).Where("software_bid_id = ? AND software_service_id = ?", bidID, SoftwareID).Delete(&ds.SoftwareService_n_SoftwareBid{})
	if result.RowsAffected == 0 {
		return errors.New("software not found in current bid")
	}
	return result.Error
}

func (d *SoftwareDatabase) UpdateSoftwareInBid(bidID int, software ds.SoftwareService_n_SoftwareBid) (ds.SoftwareService_n_SoftwareBid, error) {
	var bid ds.SoftwareService_n_SoftwareBid
	result := d.DB.Model(&ds.SoftwareService_n_SoftwareBid{}).Where("software_bid_id = ? AND software_service_id = ?", bidID, software.SoftwareServiceID).Update("count", software.Count).First(&bid)
	if result.Error != nil {
		return ds.SoftwareService_n_SoftwareBid{}, result.Error
	}
	return bid, nil
}

func (d *SoftwareDatabase) RegisterNewUser(userData ds.Users) error {
	result := d.DB.Create(&userData)
	return result.Error
}

func (d *SoftwareDatabase) GetUserAccountData(userID int) (ds.Users, error) {
	var user ds.Users
	result := d.DB.Model(&ds.Users{}).Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return ds.Users{}, result.Error
	}
	return user, nil
}

func (d *SoftwareDatabase) UpdateUserAccountData(userID int, updateUser ds.Users) (ds.Users, error) {
	var doubleUser ds.Users
	result := d.DB.Where("login = ?", updateUser.Login).First(&doubleUser)
	if result.RowsAffected != 0 {
		return ds.Users{}, errors.New("duplicate username")
	}
	result = d.DB.Where("id = ?", userID).Updates(updateUser).First(&updateUser)
	if result.Error != nil {
		return ds.Users{}, result.Error
	}
	if result.RowsAffected == 0 {
		return ds.Users{}, errors.New("user not found")
	}
	return updateUser, result.Error
}

func (d *SoftwareDatabase) AuthenticateUser(username, password string) (ds.Users, error) {
	var user ds.Users
	result := d.DB.Where("login = ?", username).First(&user)
	if result.RowsAffected == 0 {
		return ds.Users{}, errors.New("user not found")
	}
	result = d.DB.Where("password = ?", password).First(&user)
	if result.RowsAffected == 0 {
		return ds.Users{}, errors.New("incorrect password")
	}
	if result.Error != nil {
		return ds.Users{}, result.Error
	}
	return user, nil
}

func (d *SoftwareDatabase) DeauthorizeUser(jwtToken string) error {
	// var user ds.Users
	// result := d.DB.Where("id = ?", userID).First(&user)
	// if result.RowsAffected == 0 {
	// 	return errors.New("user not found")
	// }
	// return result.Error
	return nil
}
