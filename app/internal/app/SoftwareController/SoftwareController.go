package SoftwareController

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"path/filepath"
	"software/internal/app/SoftwareDatabase"
	"software/internal/app/ds"
	"software/internal/app/role"
	jwts "software/internal/jwt"
	"software/internal/middlewares"
	"strings"
	"time"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"

	_ "software/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// http://localhost/swagger/index.html

type SoftwareController struct {
	SoftwareDatabase *SoftwareDatabase.SoftwareDatabase
}

func NewSoftwareController(d *SoftwareDatabase.SoftwareDatabase) *SoftwareController {
	return &SoftwareController{
		SoftwareDatabase: d,
	}
}

func (c *SoftwareController) RegisterController(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	Guest := router.Group("/")
	Guest.Use(middlewares.AuthMiddleware(c.SoftwareDatabase.RedisClient, role.Guest, role.User, role.Admin))
	{
		Guest.GET("/api/softwares", c.GetAllSoftwareServices)
		Guest.GET("/api/softwares/:softwareID", c.GetSoftwareServiceByID)
		Guest.POST("/api/registration", c.RegisterNewUser)
		Guest.POST("/api/authentication", c.AuthenticateUser)
		Guest.POST("/api/update-calculations", c.UpdateCalculations)
	}

	User := router.Group("/")
	User.Use(middlewares.AuthMiddleware(c.SoftwareDatabase.RedisClient, role.User, role.Admin))
	{
		User.POST("/api/add-software/:softwareID", c.AddSoftwareServiceToBidByID)
		User.GET("/api/software-bids-icon", c.GetSoftwareServiceCountInBid)
		User.POST("/api/software-bids", c.GetSoftwareBids)
		User.GET("/api/software-bids/:softwareBidID", c.GetSoftwareBidByID)
		User.PUT("/api/software-bids", c.UpdateActiveSoftwareBid)
		User.PUT("/api/formation-software-bids", c.FormateActiveSoftwareBid)
		User.DELETE("/api/software-bids/:softwareBidID", c.DeleteSoftwareBid)
		User.DELETE("/api/delete-software-from-bid/:softwareID", c.DeleteSoftwareFromBid)
		User.PUT("/api/update-software-in-bid", c.UpdateSoftwareInBid)
		User.GET("/api/account/:userID", c.GetUserAccountData)
		User.PUT("/api/account/:userID", c.UpdateUserAccountData)
		User.POST("/api/deauthorization", c.DeauthorizeUser)
		// User.POST("/api/calculate-bid/:bidID", c.CalculateBidServices)
	}

	Admin := router.Group("/")
	Admin.Use(middlewares.AuthMiddleware(c.SoftwareDatabase.RedisClient, role.Admin))
	{
		Admin.POST("/api/softwares", c.AddNewSoftware)
		Admin.PUT("/api/softwares/:softwareID", c.UpdateSoftware)
		Admin.DELETE("/api/softwares/:softwareID", c.DeleteSoftwareWithPhoto)
		Admin.POST("/api/add-photo/:softwareID", c.AddPhotoToSoftwareService)
		Admin.PUT("/api/moderator-software-bids/:softwareBidID", c.ModerateSoftwareBid)
	}
}

func (c *SoftwareController) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (c *SoftwareController) errorController(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"description": err.Error(),
	})
}

// GetAllSoftwareServices godoc
// @Summary Получить все услуги ПО
// @Description Получает список всех услуг ПО с фильтрацией
// @Tags Услуги
// @Accept json
// @Produce json
// @Param software-search query string false "Поисковый запрос"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/softwares [get]
func (c *SoftwareController) GetAllSoftwareServices(ctx *gin.Context) {
	var softwares []ds.SoftwareService
	var err error

	searchQuery := ctx.Query("software-search")
	if searchQuery == "" {
		softwares, err = c.SoftwareDatabase.GetSoftwareServices()
		if err != nil {
			c.errorController(ctx, http.StatusInternalServerError, err)
			return
		}
	} else {
		softwares, err = c.SoftwareDatabase.GetSoftwareServicesByTitle(searchQuery)
		if err != nil {
			c.errorController(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"softwares": softwares,
	})
}

// GetSoftwareServiceByID godoc
// @Summary Получить услугу ПО по ID
// @Description Получает информацию о конкретной услуге ПО по её ID
// @Tags Услуги
// @Accept json
// @Produce json
// @Param softwareID path int true "ID услуги ПО"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/softwares/{softwareID} [get]
func (c *SoftwareController) GetSoftwareServiceByID(ctx *gin.Context) {
	idStr := ctx.Param("softwareID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	software, err := c.SoftwareDatabase.GetSoftwareService(id)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"software": software,
	})
}

// AddNewSoftware godoc
// @Summary Добавить новую услугу ПО
// @Description Создает новую услугу ПО (только для модераторов)
// @Tags Услуги
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param software body ds.SoftwareService true "Данные услуги ПО"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/softwares [post]
func (c *SoftwareController) AddNewSoftware(ctx *gin.Context) {
	rrole, exs := ctx.Get("role")
	if !exs {
		log.Printf("not exist")
	} else {
		log.Printf("%s", rrole)
	}
	software := ds.SoftwareService{}
	err := ctx.ShouldBindJSON(&software)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	err = c.SoftwareDatabase.AddNewSoftware(software)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"software": software,
	})
}

// UpdateSoftware godoc
// @Summary Обновить услугу ПО
// @Description Обновляет информацию об услуге ПО (только для модераторов)
// @Tags Услуги
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param softwareID path int true "ID услуги ПО"
// @Param software body ds.SoftwareService true "Обновленные данные услуги ПО"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/softwares/{softwareID} [put]
func (c *SoftwareController) UpdateSoftware(ctx *gin.Context) {
	strSoftwareID := ctx.Param("softwareID")
	softwareID, err := strconv.Atoi(strSoftwareID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	software := ds.SoftwareService{}
	err = ctx.ShouldBindJSON(&software)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	software, err = c.SoftwareDatabase.UpdateSoftware(softwareID, software)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"software": software,
	})
}

// DeleteSoftwareWithPhoto godoc
// @Summary Удалить услугу ПО с фотографией
// @Description Удаляет услугу ПО и связанную с ней фотографию (только для модераторов)
// @Tags Услуги
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param softwareID path int true "ID услуги ПО"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/softwares/{softwareID} [delete]
func (c *SoftwareController) DeleteSoftwareWithPhoto(ctx *gin.Context) {
	strSoftwareID := ctx.Param("softwareID")
	softwareID, err := strconv.Atoi(strSoftwareID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	softwareID, err = c.SoftwareDatabase.DeleteSoftwareWithPhoto(softwareID)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"softwareID": softwareID,
	})
}

// AddSoftwareServiceToBidByID godoc
// @Summary Добавить услугу ПО в заявку
// @Description Добавляет услугу ПО в активную заявку пользователя
// @Tags Услуги
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param softwareID path int true "ID услуги ПО"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/add-software/{softwareID} [post]
func (c *SoftwareController) AddSoftwareServiceToBidByID(ctx *gin.Context) {
	strSoftwareID := ctx.Param("softwareID")
	softwareID, err := strconv.Atoi(strSoftwareID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	var bidID int

	creatorIDVal, _ := ctx.Get("userID")
	creatorID := creatorIDVal.(int)
	bidID, err = c.SoftwareDatabase.FindUserActiveBid(creatorID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	if bidID == 0 {
		bidID = c.SoftwareDatabase.CreateUserActiveBid(creatorID)
	}

	_, err = c.SoftwareDatabase.AddSoftwareServiceToBid(softwareID, bidID)
	if err != nil {
		if err.Error() == "duplicate" {
			c.errorController(ctx, http.StatusBadRequest, err)
			return
		}
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"softwareID": softwareID,
		"bidID":      bidID,
	})
}

// AddPhotoToSoftwareService godoc
// @Summary Добавить фотографию к услуге ПО
// @Description Загружает и прикрепляет фотографию к услуге ПО (только для модераторов)
// @Tags Услуги
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param softwareID path int true "ID услуги ПО"
// @Param photo_name formData string true "Название фотографии"
// @Param file formData file true "Файл фотографии"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/add-photo/{softwareID} [post]
func (c *SoftwareController) AddPhotoToSoftwareService(ctx *gin.Context) {
	type UploadPhotoRequest struct {
		PhotoName string `form:"photo_name" binding:"required"`
	}

	var request UploadPhotoRequest

	softwareID, err := strconv.Atoi(ctx.Param("softwareID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	if err := ctx.ShouldBind(&request); err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ext := filepath.Ext(file.Filename)
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
	}

	if !allowedExtensions[ext] {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	photo, err := c.SoftwareDatabase.AddPhotoToSoftwareService(file, request.PhotoName, softwareID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"photo": photo,
	})
}

// GetSoftwareServiceCountInBid godoc
// @Summary Получить количество услуг в активной заявке
// @Description Возвращает количество услуг ПО в активной заявке пользователя
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/software-bids-icon [get]
func (c *SoftwareController) GetSoftwareServiceCountInBid(ctx *gin.Context) {
	var softwares []ds.SoftwareService

	creatorIDVal, _ := ctx.Get("userID")
	creatorID := creatorIDVal.(int)
	bidID, err := c.SoftwareDatabase.FindUserActiveBid(creatorID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	_, softwares, err = c.SoftwareDatabase.GetSoftwareServicesBid(bidID)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bidID": bidID,
		"count": len(softwares),
	})
}

// GetSoftwareBids godoc
// @Summary Получить заявки пользователя
// @Description Возвращает список заявок пользователя с фильтрацией
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param filter body ds.FilterRequest false "Параметры фильтрации"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/software-bids [post]
func (c *SoftwareController) GetSoftwareBids(ctx *gin.Context) {
	var filter ds.FilterRequest
	var bids []ds.SoftwareBid

	err := ctx.ShouldBindJSON(&filter)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			c.errorController(ctx, http.StatusBadRequest, err)
			return
		}
	}

	creatorIDVal, _ := ctx.Get("userID")
	creatorID := creatorIDVal.(int)
	bids, err = c.SoftwareDatabase.GetSoftwareBids(creatorID, filter)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bids": bids,
	})
}

// GetSoftwareBidByID godoc
// @Summary Получить заявку по ID
// @Description Возвращает информацию о конкретной заявке по её ID
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param softwareBidID path int true "ID заявки"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/software-bids/{softwareBidID} [get]
func (c *SoftwareController) GetSoftwareBidByID(ctx *gin.Context) {
	bidID, err := strconv.Atoi(ctx.Param("softwareBidID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	var bid ds.SoftwareBid

	bid, err = c.SoftwareDatabase.GetSoftwareBidByID(bidID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bid": bid,
	})
}

// UpdateActiveSoftwareBid godoc
// @Summary Обновить активную заявку
// @Description Обновляет данные активной заявки пользователя
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bid body ds.SoftwareBid true "Обновлённые данные заявки"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/software-bids [put]
func (c *SoftwareController) UpdateActiveSoftwareBid(ctx *gin.Context) {
	var bid ds.SoftwareBid
	err := ctx.ShouldBindJSON(&bid)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	creatorIDVal, _ := ctx.Get("userID")
	creatorID := creatorIDVal.(int)
	bidID, err := c.SoftwareDatabase.FindUserActiveBid(creatorID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	bid, err = c.SoftwareDatabase.UpdateActiveSoftwareBid(bidID, bid)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bid": bid,
	})
}

// FormateActiveSoftwareBid godoc
// @Summary Оформить активную заявку
// @Description Оформляет активную заявку пользователя
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/formation-software-bids [put]
func (c *SoftwareController) FormateActiveSoftwareBid(ctx *gin.Context) {
	var softwaresInBid []ds.SoftwareService_n_SoftwareBid

	creatorIDVal, _ := ctx.Get("userID")
	creatorID := creatorIDVal.(int)
	bidID, err := c.SoftwareDatabase.FindUserActiveBid(creatorID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	softwaresInBid, err = c.SoftwareDatabase.CountServicesInBid(bidID)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	for _, software := range softwaresInBid {
		if software.Count == 0 {
			c.errorController(ctx, http.StatusBadRequest, errors.New("count must be positive"))
			return
		}
	}

	bidID, err = c.SoftwareDatabase.FormateActiveSoftwareBid(bidID)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bidID": bidID,
	})
}

// ModerateSoftwareBid godoc
// @Summary Модерировать заявку
// @Description Одобряет или отклоняет заявку (только для модераторов)
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param softwareBidID path int true "ID заявки"
// @Param approved query boolean false "Статус одобрения"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/moderator-software-bids/{softwareBidID} [put]
func (c *SoftwareController) ModerateSoftwareBid(ctx *gin.Context) {
	bidID, err := strconv.Atoi(ctx.Param("softwareBidID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	bid, err := c.SoftwareDatabase.GetSoftwareBidByID(bidID)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	servicesData := []map[string]interface{}{}
	for _, software := range bid.Softwares {
		serviceData := map[string]interface{}{
			"software_service_id": software.SoftwareServiceID,
			"price":               software.Price,
			"page_count":          software.Count,
			"bid_id":              bidID,
		}
		servicesData = append(servicesData, serviceData)
	}

	requestData := map[string]interface{}{
		"bid_id":     bidID,
		"services":   servicesData,
		"auth_token": "12345678",
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	asyncServiceURL := "http://localhost:8000/api/calculate/"
	resp, err := http.Post(asyncServiceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.errorController(ctx, http.StatusInternalServerError,
			errors.New("async service returned error"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "calculation_started",
		"bid_id":  bidID,
		"message": "Calculation process initiated in async service",
	})

	// approved := false
	// if len(ctx.Query("approved")) > 0 {
	// 	approved = true
	// }

	// bidID, cost, err := c.SoftwareDatabase.ModerateSoftwareBid(bidID, approved)
	// if err != nil {
	// 	c.errorController(ctx, http.StatusInternalServerError, err)
	// 	return
	// }

	// ctx.JSON(http.StatusOK, gin.H{
	// 	"cost":  cost,
	// 	"bidID": bidID,
	// })
}

// UpdateCalculations godoc
// @Summary Обновить результаты расчетов
// @Description Принимает результаты расчетов от асинхронного сервиса
// @Tags Заявки
// @Accept json
// @Produce json
// @Param calculations body ds.CalculationResult true "Результаты расчетов"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/update-calculations [post]
func (c *SoftwareController) UpdateCalculations(ctx *gin.Context) {
	var calculationRequest struct {
		BidID     int                    `json:"bid_id"`
		Services  []ds.CalculationResult `json:"services"`
		AuthToken string                 `json:"auth_token"`
	}

	err := ctx.ShouldBindJSON(&calculationRequest)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	if calculationRequest.AuthToken != "12345678" {
		c.errorController(ctx, http.StatusUnauthorized, errors.New("invalid auth token"))
		return
	}

	for _, result := range calculationRequest.Services {
		err = c.SoftwareDatabase.UpdateCalculationResult(result, calculationRequest.BidID)
		if err != nil {
			c.errorController(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	_, _, err = c.SoftwareDatabase.ModerateSoftwareBid(calculationRequest.BidID, true)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":        "calculations_updated",
		"updated_count": len(calculationRequest.Services),
		"message":       "Calculation results successfully updated",
	})
}

// DeleteSoftwareBid godoc
// @Summary Удалить заявку
// @Description Удаляет заявку по ID
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param softwareBidID path int true "ID заявки"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/software-bids/{softwareBidID} [delete]
func (c *SoftwareController) DeleteSoftwareBid(ctx *gin.Context) {
	bidID, err := strconv.Atoi(ctx.Param("softwareBidID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	bidID, err = c.SoftwareDatabase.DeleteSoftwareBid(bidID)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bidID": bidID,
	})
}

// DeleteSoftwareFromBid godoc
// @Summary Удалить услугу ПО из заявки
// @Description Удаляет услугу ПО из активной заявки пользователя
// @Tags М-М
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param softwareID path int true "ID услуги ПО"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/delete-software-from-bid/{softwareID} [delete]
func (c *SoftwareController) DeleteSoftwareFromBid(ctx *gin.Context) {
	softwareID, err := strconv.Atoi(ctx.Param("softwareID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	creatorIDVal, _ := ctx.Get("userID")
	creatorID := creatorIDVal.(int)
	bidID, err := c.SoftwareDatabase.FindUserActiveBid(creatorID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	err = c.SoftwareDatabase.DeleteSoftwareFromBid(bidID, softwareID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bidID":      bidID,
		"softwareID": softwareID,
	})
}

// UpdateSoftwareInBid godoc
// @Summary Обновить услугу ПО в заявке
// @Description Обновляет информацию об услуге ПО в активной заявке
// @Tags М-М
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param softwareInBid body ds.SoftwareService_n_SoftwareBid true "Данные услуги ПО в заявке"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/update-software-in-bid [put]
func (c *SoftwareController) UpdateSoftwareInBid(ctx *gin.Context) {
	softwareInBid := ds.SoftwareService_n_SoftwareBid{}

	err := ctx.ShouldBindJSON(&softwareInBid)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	creatorIDVal, _ := ctx.Get("userID")
	creatorID := creatorIDVal.(int)
	bidID, err := c.SoftwareDatabase.FindUserActiveBid(creatorID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	bid, err := c.SoftwareDatabase.UpdateSoftwareInBid(bidID, softwareInBid)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bid": bid,
	})
}

// RegisterNewUser godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя в системе
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param newUser body ds.Users true "Данные нового пользователя"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/registration [post]
func (c *SoftwareController) RegisterNewUser(ctx *gin.Context) {
	var newUser ds.Users

	err := ctx.ShouldBindJSON(&newUser)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	err = c.SoftwareDatabase.RegisterNewUser(newUser)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"username": newUser.Login,
	})
}

// GetUserAccountData godoc
// @Summary Получить данные аккаунта пользователя
// @Description Возвращает информацию о пользователе по его ID
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userID path int true "ID пользователя"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/account/{userID} [get]
func (c *SoftwareController) GetUserAccountData(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	userData, err := c.SoftwareDatabase.GetUserAccountData(userID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"userData": userData,
	})
}

// UpdateUserAccountData godoc
// @Summary Обновить данные аккаунта пользователя
// @Description Обновляет информацию о пользователе
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userID path int true "ID пользователя"
// @Param updateUser body ds.Users true "Обновленные данные пользователя"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/account/{userID} [put]
func (c *SoftwareController) UpdateUserAccountData(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	var updateUser ds.Users

	err = ctx.ShouldBindJSON(&updateUser)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := c.SoftwareDatabase.UpdateUserAccountData(userID, updateUser)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// AuthenticateUser godoc
// @Summary Аутентификация пользователя
// @Description Выполняет вход пользователя в систему и возвращает JWT токен
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param request body ds.UserData true "Данные для аутентификации"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/authentication [post]
func (c *SoftwareController) AuthenticateUser(ctx *gin.Context) {
	var userData ds.UserData

	err := ctx.ShouldBindJSON(&userData)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	var foundUser ds.Users

	foundUser, err = c.SoftwareDatabase.AuthenticateUser(userData.Username, userData.Password)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	token, err := jwts.GenerateToken(foundUser)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":  foundUser,
		"token": token,
	})
}

// DeauthorizeUser godoc
// @Summary Выход пользователя из системы
// @Description Добавляет JWT токен в черный список для деактивации
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Успешный выход"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/deauthorization [post]
func (c *SoftwareController) DeauthorizeUser(ctx *gin.Context) {
	jwtStr := ctx.GetHeader("Authorization")
	prefix := "Bearer "
	if !strings.HasPrefix(jwtStr, prefix) {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	jwtStr = jwtStr[len(prefix):]

	claims, err := jwts.ValidateToken(jwtStr)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		log.Println(err)
		return
	}

	_, err = jwt.ParseWithClaims(jwtStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(ds.JWTSecret), nil
	})
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		log.Println(err)
		return
	}

	err = c.SoftwareDatabase.RedisClient.WriteJWTToBlacklist(ctx.Request.Context(), jwtStr, time.Hour*24)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusOK)
}
