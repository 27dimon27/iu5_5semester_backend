package main

import (
	"log"
	"softwareDev/internal/app/ds"
	"softwareDev/internal/app/dsn"
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

	err = db.AutoMigrate(
		&ds.SoftwareDevService{},
		&ds.SoftwareDevBid{},
		&ds.Service_n_Bid{},
	)
	if err != nil {
		panic("cant migrate db")
	}

	err = seedDatabase(db)
	if err != nil {
		log.Printf("Warning: could not seed database: %v", err)
	}
}

func seedDatabase(db *gorm.DB) error {
	var count int64
	db.Model(&ds.SoftwareDevService{}).Count(&count)

	if count == 0 {
		initialServices := []ds.SoftwareDevService{
			{
				ID:           1,
				Image:        "web-project.png",
				Title:        "Проектирование веб-приложения",
				Description:  "Профессиональная команда разработчиков спроектирует веб-приложение по Вашему техническому заданию!",
				Price:        5_000,
				PriceMeasure: "руб/стр",
				Status:       true,
			},
			{
				ID:           2,
				Image:        "desktop-project.png",
				Title:        "Проектирование десктопного приложения",
				Description:  "Профессиональная команда разработчиков спроектирует десктопное приложение по Вашему техническому заданию!",
				Price:        6_000,
				PriceMeasure: "руб/стр",
				Status:       true,
			},
			{
				ID:           3,
				Image:        "mobile-project.png",
				Title:        "Проектирование мобильного приложения",
				Description:  "Профессиональная команда разработчиков спроектирует мобильное приложение по Вашему техническому заданию!",
				Price:        7_000,
				PriceMeasure: "руб/стр",
				Status:       true,
			},
			{
				ID:           4,
				Image:        "web.png",
				Title:        "Разработка веб-приложения",
				Description:  "Профессиональная команда разработчиков разработает веб-приложение по Вашему техническому заданию!",
				Price:        40_000,
				PriceMeasure: "руб/стр",
				Status:       true,
			},
			{
				ID:           5,
				Image:        "desktop.png",
				Title:        "Разработка десктопного приложения",
				Description:  "Профессиональная команда разработчиков разработает десктопное приложение по Вашему техническому заданию!",
				Price:        45_000,
				PriceMeasure: "руб/стр",
				Status:       true,
			},
			{
				ID:           6,
				Image:        "mobile.png",
				Title:        "Разработка мобильного приложения",
				Description:  "Профессиональная команда разработчиков разработает мобильное приложение по Вашему техническому заданию!",
				Price:        50_000,
				PriceMeasure: "руб/стр",
				Status:       true,
			},
			{
				ID:           7,
				Image:        "test.png",
				Title:        "Тестирование десктопного, мобильного и веб-приложений",
				Description:  "Профессиональная команда разработчиков протестирует Ваше приложение на предмет наличия уязвимостей!",
				Price:        2_500,
				PriceMeasure: "руб/кейс",
				Status:       true,
			},
			{
				ID:           8,
				Image:        "ui-ux.png",
				Title:        "Проектирование UX/UI дизайна",
				Description:  "Профессиональная команда разработчиков разработает UX/UI дизайн для Вашего приложения по Вашему техническому заданию!",
				Price:        8_000,
				PriceMeasure: "руб/стр",
				Status:       true,
			},
			{
				ID:           9,
				Image:        "audit.png",
				Title:        "Техническая консультация и аудит проекта",
				Description:  "Профессиональная команда разработчиков проконсультирует Вас по проекту и даст объективную и честную оценку!",
				Price:        3_500,
				PriceMeasure: "руб/час",
				Status:       true,
			},
		}

		result := db.Create(&initialServices)
		if result.Error != nil {
			return result.Error
		}
		log.Printf("Successfully seeded database with %d services", len(initialServices))
	}

	db.Model(&ds.SoftwareDevBid{}).Count(&count)

	if count == 0 {
		initialBids := []ds.SoftwareDevBid{
			{
				Status:     "черновик",
				DateCreate: time.Now(),
				CreatorID:  1,
				// Services:   []int64{5, 7, 9},
			},
		}

		result := db.Create(&initialBids)
		if result.Error != nil {
			return result.Error
		}
		log.Printf("Successfully seeded database with %d bids", len(initialBids))
	}

	db.Model(&ds.Service_n_Bid{}).Count(&count)

	if count == 0 {
		initialServiceBid := []ds.Service_n_Bid{
			{
				ServiceID: 5,
				BidID:     1,
				Count:     1,
				Index:     0,
			},
			{
				ServiceID: 7,
				BidID:     1,
				Count:     1,
				Index:     1,
			},
			{
				ServiceID: 9,
				BidID:     1,
				Count:     1,
				Index:     2,
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
