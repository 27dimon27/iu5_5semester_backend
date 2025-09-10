package SoftwareDevServiceController

import (
	"lab1/internal/app/SoftwareDevServiceDatabase"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SoftwareDevServiceController struct {
	SoftwareDevServiceDatabase *SoftwareDevServiceDatabase.SoftwareDevServiceDatabase
}

func NewSoftwareDevServiceController(s *SoftwareDevServiceDatabase.SoftwareDevServiceDatabase) *SoftwareDevServiceController {
	return &SoftwareDevServiceController{
		SoftwareDevServiceDatabase: s,
	}
}

func (c *SoftwareDevServiceController) GetSoftwareDevServices(ctx *gin.Context) {
	var services []SoftwareDevServiceDatabase.SoftwareDevService
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

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"services":      services,
		"serviceSearch": searchQuery,
	})
}

func (c *SoftwareDevServiceController) GetSoftwareDevService(ctx *gin.Context) {
	idStr := ctx.Param("idService")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	service, err := c.SoftwareDevServiceDatabase.GetSoftwareDevService(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "service.html", gin.H{
		"service": service,
	})
}

func (c *SoftwareDevServiceController) GetSoftwareDevServicesBid(ctx *gin.Context) {
	var bid SoftwareDevServiceDatabase.SoftwareDevServiceBid
	var err error

	bid, err = c.SoftwareDevServiceDatabase.GetSoftwareDevServicesBid()
	if err != nil {
		logrus.Error(err)
	}

	var sum int
	for _, service := range bid.Services {
		sum += service.Price
	}

	coefficients := c.SoftwareDevServiceDatabase.GetCoefficients()

	ctx.HTML(http.StatusOK, "bid.html", gin.H{
		"bid":          bid,
		"sum":          sum,
		"coefficients": coefficients,
	})
}
