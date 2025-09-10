package SoftwareDevServiceDatabase

import (
	"fmt"
	"strings"
)

type SoftwareDevServiceDatabase struct {
}

func NewSoftwareDevServiceDatabase() (*SoftwareDevServiceDatabase, error) {
	return &SoftwareDevServiceDatabase{}, nil
}

type SoftwareDevService struct {
	ID           int
	Image        string
	Title        string
	Description  string
	Team         string
	Price        int
	PriceMeasure string
	Grade        string
	Deadline     int
}

type SoftwareDevServiceBid struct {
	ID       int
	Services []SoftwareDevService
}

type Coefficient struct {
	Level string
	Coeff float32
}

func (r *SoftwareDevServiceDatabase) GetCoefficients() []Coefficient {
	Coefficients := []Coefficient{
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

func (r *SoftwareDevServiceDatabase) GetSoftwareDevServices() ([]SoftwareDevService, error) {
	services := []SoftwareDevService{
		{
			ID:           1,
			Image:        "web-project.png",
			Title:        "Проектирование веб-приложения",
			Description:  "Профессиональная команда разработчиков спроектирует веб-приложение по Вашему техническому заданию!",
			Team:         "Больше команда - быстрее результат :D",
			Price:        5_000,
			PriceMeasure: "руб/стр",
		},
		{
			ID:           2,
			Image:        "desktop-project.png",
			Title:        "Проектирование десктопного приложения",
			Description:  "Профессиональная команда разработчиков спроектирует десктопное приложение по Вашему техническому заданию!",
			Team:         "Больше команда - быстрее результат :D",
			Price:        6_000,
			PriceMeasure: "руб/стр",
		},
		{
			ID:           3,
			Image:        "mobile-project.png",
			Title:        "Проектирование мобильного приложения",
			Description:  "Профессиональная команда разработчиков спроектирует мобильное приложение по Вашему техническому заданию!",
			Team:         "Больше команда - быстрее результат :D",
			Price:        7_000,
			PriceMeasure: "руб/стр",
		},
		{
			ID:           4,
			Image:        "web.png",
			Title:        "Разработка веб-приложения",
			Description:  "Профессиональная команда разработчиков разработает веб-приложение по Вашему техническому заданию!",
			Team:         "Больше команда - быстрее результат :D",
			Price:        40_000,
			PriceMeasure: "руб/стр",
		},
		{
			ID:           5,
			Image:        "desktop.png",
			Title:        "Разработка десктопного приложения",
			Description:  "Профессиональная команда разработчиков разработает десктопное приложение по Вашему техническому заданию!",
			Team:         "Больше команда - быстрее результат :D",
			Price:        45_000,
			PriceMeasure: "руб/стр",
		},
		{
			ID:           6,
			Image:        "mobile.png",
			Title:        "Разработка мобильного приложения",
			Description:  "Профессиональная команда разработчиков разработает мобильное приложение по Вашему техническому заданию!",
			Team:         "Больше команда - быстрее результат :D",
			Price:        50_000,
			PriceMeasure: "руб/стр",
		},
		{
			ID:           7,
			Image:        "test.png",
			Title:        "Тестирование десктопного, мобильного и веб-приложений",
			Description:  "Профессиональная команда разработчиков протестирует Ваше приложение на предмет наличия уязвимостей!",
			Team:         "Больше команда - быстрее результат :D",
			Price:        2_500,
			PriceMeasure: "руб/кейс",
		},
		{
			ID:           8,
			Image:        "ui-ux.png",
			Title:        "Проектирование UX/UI дизайна",
			Description:  "Профессиональная команда разработчиков разработает UX/UI дизайн для Вашего приложения по Вашему техническому заданию!",
			Team:         "Больше команда - быстрее результат :D",
			Price:        8_000,
			PriceMeasure: "руб/стр",
		},
		{
			ID:           9,
			Image:        "audit.png",
			Title:        "Техническая консультация и аудит проекта",
			Description:  "Профессиональная команда разработчиков проконсультирует Вас по проекту и даст объективную и честную оценку!",
			Team:         "Больше команда - быстрее результат :D",
			Price:        3_500,
			PriceMeasure: "руб/час",
		},
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("empty array")
	}

	return services, nil
}

func (r *SoftwareDevServiceDatabase) GetSoftwareDevService(id int) (SoftwareDevService, error) {
	services, err := r.GetSoftwareDevServices()
	if err != nil {
		return SoftwareDevService{}, err
	}

	for _, service := range services {
		if service.ID == id {
			return service, nil
		}
	}

	return SoftwareDevService{}, fmt.Errorf("order not found")
}

func (r *SoftwareDevServiceDatabase) GetSoftwareDevServicesByTitle(title string) ([]SoftwareDevService, error) {
	services, err := r.GetSoftwareDevServices()
	if err != nil {
		return []SoftwareDevService{}, err
	}

	var result []SoftwareDevService
	for _, service := range services {
		if strings.Contains(strings.ToLower(service.Title), strings.ToLower(title)) {
			result = append(result, service)
		}
	}

	return result, nil
}

func (r *SoftwareDevServiceDatabase) GetSoftwareDevServicesBid() (SoftwareDevServiceBid, error) {
	bid_services := []int{5, 7, 9}
	services, _ := r.GetSoftwareDevServices()
	bid := SoftwareDevServiceBid{
		ID:       1,
		Services: []SoftwareDevService{},
	}

	for _, bid_service := range bid_services {
		for _, service := range services {
			if bid_service == service.ID {
				bid.Services = append(bid.Services, service)
			}
		}
	}

	if len(bid.Services) == 0 {
		return SoftwareDevServiceBid{}, fmt.Errorf("empty bid")
	}

	return bid, nil
}
