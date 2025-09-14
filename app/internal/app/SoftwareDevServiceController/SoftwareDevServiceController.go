package SoftwareDevServiceController

import (
	"softwareDev/internal/app/SoftwareDevServiceDatabase"
	"softwareDev/internal/app/ds"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SoftwareDevServiceController struct {
	SoftwareDevServiceDatabase *SoftwareDevServiceDatabase.SoftwareDevServiceDatabase
}

func NewSoftwareDevServiceController(d *SoftwareDevServiceDatabase.SoftwareDevServiceDatabase) *SoftwareDevServiceController {
	return &SoftwareDevServiceController{
		SoftwareDevServiceDatabase: d,
	}
}

func (c *SoftwareDevServiceController) RegisterController(router *gin.Engine) {
	router.GET("/services", c.GetSoftwareDevServices)
	router.GET("/services/:serviceID", c.GetSoftwareDevService)
	router.GET("/bids/:bidID", c.GetSoftwareDevServicesBid)
}

func (c *SoftwareDevServiceController) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (c *SoftwareDevServiceController) errorController(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

func (c *SoftwareDevServiceController) GetSoftwareDevServices(ctx *gin.Context) {
	var services []ds.SoftwareDevService
	var bidCount []ds.SoftwareDevService
	var err error

	searchQuery := ctx.Query("serviceSearch")
	if searchQuery == "" {
		services, err = c.SoftwareDevServiceDatabase.GetSoftwareDevServices()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		services, err = c.SoftwareDevServiceDatabase.GetSoftwareDevServicesByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	userID := 1

	_, bidCount, err = c.SoftwareDevServiceDatabase.GetSoftwareDevServicesBid(userID)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "MainPage.html", gin.H{
		"services":      services,
		"serviceSearch": searchQuery,
		"bidCount":      len(bidCount),
		"userID":        userID,
	})
}

func (c *SoftwareDevServiceController) GetSoftwareDevService(ctx *gin.Context) {
	idStr := ctx.Param("serviceID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	service, err := c.SoftwareDevServiceDatabase.GetSoftwareDevService(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "ServicePage.html", gin.H{
		"service": service,
	})
}

func (c *SoftwareDevServiceController) GetSoftwareDevServicesBid(ctx *gin.Context) {
	var bid ds.SoftwareDevBid
	var services []ds.SoftwareDevService
	var err error

	bidID, err := strconv.Atoi(ctx.Param("bidID"))
	if err != nil {
		logrus.Error(err)
	}

	bid, services, err = c.SoftwareDevServiceDatabase.GetSoftwareDevServicesBid(bidID)
	if err != nil {
		logrus.Error(err)
	}

	var sum float32
	for _, service := range services {
		sum += service.Price
	}

	coefficients := c.SoftwareDevServiceDatabase.GetCoefficients()

	companies := []string{
		"Apple", "Microsoft", "Google", "Amazon", "Tesla",
	}

	ctx.HTML(http.StatusOK, "BidPage.html", gin.H{
		"bid":          bid,
		"services":     services,
		"companies":    companies,
		"sum":          sum,
		"coefficients": coefficients,
	})
}
