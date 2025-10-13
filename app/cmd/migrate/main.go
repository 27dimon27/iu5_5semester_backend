package main

import (
	"log"
	"software/internal/app/ds"
	"software/internal/app/dsn"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = dropTablesInDatabase(db)
	if err != nil {
		panic("cant drop db")
	}

	err = db.AutoMigrate(
		&ds.SoftwareService{},
		&ds.SoftwareBid{},
		&ds.SoftwareService_n_SoftwareBid{},
		&ds.Users{},
	)
	if err != nil {
		panic("cant migrate db")
	}

	err = seedDatabase(db)
	if err != nil {
		log.Printf("Warning: could not seed database: %v", err)
	}
}

func dropTablesInDatabase(db *gorm.DB) error {
	if err := db.Exec("DROP TABLE IF EXISTS software_service_n_software_bids").Error; err != nil {
		return err
	}
	if err := db.Exec("DROP TABLE IF EXISTS software_bids").Error; err != nil {
		return err
	}
	if err := db.Exec("DROP TABLE IF EXISTS software_services").Error; err != nil {
		return err
	}
	if err := db.Exec("DROP TABLE IF EXISTS users").Error; err != nil {
		return err
	}

	log.Println("All tables cleared successfully")
	return nil
}

func seedDatabase(db *gorm.DB) error {
	var count int64

	db.Model(&ds.Users{}).Count(&count)

	if count == 0 {
		initialUsers := []ds.Users{
			{
				Login:       "user",
				Password:    "user",
				IsModerator: false,
			},
			{
				Login:       "admin",
				Password:    "admin",
				IsModerator: true,
			},
		}

		result := db.Create(&initialUsers)
		if result.Error != nil {
			return result.Error
		}
		log.Printf("Successfully seeded database with %d users", len(initialUsers))
	}

	db.Model(&ds.SoftwareService{}).Count(&count)

	if count == 0 {
		initialServices := []ds.SoftwareService{
			{
				Image:       "web-project.png",
				Title:       "Проектирование веб-приложения",
				Description: "Профессиональная команда разработчиков спроектирует веб-приложение по Вашему техническому заданию!",
				Price:       5_000,
				Status:      true,
			},
			{
				Image:       "desktop-project.png",
				Title:       "Проектирование десктопного приложения",
				Description: "Профессиональная команда разработчиков спроектирует десктопное приложение по Вашему техническому заданию!",
				Price:       6_000,
				Status:      true,
			},
			{
				Image:       "mobile-project.png",
				Title:       "Проектирование мобильного приложения",
				Description: "Профессиональная команда разработчиков спроектирует мобильное приложение по Вашему техническому заданию!",
				Price:       7_000,
				Status:      true,
			},
			{
				Image:       "web.png",
				Title:       "Разработка веб-приложения",
				Description: "Профессиональная команда разработчиков разработает веб-приложение по Вашему техническому заданию!",
				Price:       40_000,
				Status:      true,
			},
			{
				Image:       "desktop.png",
				Title:       "Разработка десктопного приложения",
				Description: "Профессиональная команда разработчиков разработает десктопное приложение по Вашему техническому заданию!",
				Price:       45_000,
				Status:      true,
			},
			{
				Image:       "mobile.png",
				Title:       "Разработка мобильного приложения",
				Description: "Профессиональная команда разработчиков разработает мобильное приложение по Вашему техническому заданию!",
				Price:       50_000,
				Status:      true,
			},
			{
				Image:       "test.png",
				Title:       "Тестирование десктопного, мобильного и веб-приложений",
				Description: "Профессиональная команда разработчиков протестирует Ваше приложение на предмет наличия уязвимостей!",
				Price:       2_500,
				Status:      true,
			},
			{
				Image:       "ui-ux.png",
				Title:       "Проектирование UX/UI дизайна",
				Description: "Профессиональная команда разработчиков разработает UX/UI дизайн для Вашего приложения по Вашему техническому заданию!",
				Price:       8_000,
				Status:      true,
			},
			{
				Image:       "audit.png",
				Title:       "Техническая консультация и аудит проекта",
				Description: "Профессиональная команда разработчиков проконсультирует Вас по проекту и даст объективную и честную оценку!",
				Price:       3_500,
				Status:      true,
			},
		}

		result := db.Create(&initialServices)
		if result.Error != nil {
			return result.Error
		}
		log.Printf("Successfully seeded database with %d services", len(initialServices))
	}

	db.Model(&ds.SoftwareBid{}).Count(&count)

	if count == 0 {
		initialBids := []ds.SoftwareBid{
			{
				Status:     "черновик",
				DateCreate: time.Now().Format("2006-01-02"),
				CreatorID:  1,
			},
		}

		result := db.Create(&initialBids)
		if result.Error != nil {
			return result.Error
		}
		log.Printf("Successfully seeded database with %d bids", len(initialBids))
	}

	db.Model(&ds.SoftwareService_n_SoftwareBid{}).Count(&count)

	if count == 0 {
		initialServiceBid := []ds.SoftwareService_n_SoftwareBid{
			{
				SoftwareServiceID: 5,
				SoftwareBidID:     1,
				Count:             1,
				Index:             1,
				Price:             45_000,
			},
			{
				SoftwareServiceID: 7,
				SoftwareBidID:     1,
				Count:             1,
				Index:             2,
				Price:             2_500,
			},
			{
				SoftwareServiceID: 9,
				SoftwareBidID:     1,
				Count:             1,
				Index:             3,
				Price:             3_500,
			},
		}

		result := db.Create(&initialServiceBid)
		if result.Error != nil {
			return result.Error
		}
		log.Printf("Successfully seeded database with %d service-&-bid", len(initialServiceBid))
	}

	return nil
}
