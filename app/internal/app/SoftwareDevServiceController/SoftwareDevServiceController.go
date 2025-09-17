package SoftwareDevServiceController

import (
	"math"
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
	router.GET("/bids/:userID", c.GetSoftwareDevServicesBid)

	router.POST("/:userID/addservice/:serviceID", c.AddSoftwareDevServiceToBid)
	router.POST("/bids/deletebid/:userID", c.DeleteSoftwareDevBid)
	router.POST("/bids/calcbid/:userID", c.GetSoftwareDevServicesBid)
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
			c.errorController(ctx, http.StatusInternalServerError, err)
		}
	} else {
		services, err = c.SoftwareDevServiceDatabase.GetSoftwareDevServicesByTitle(searchQuery)
		if err != nil {
			c.errorController(ctx, http.StatusInternalServerError, err)
		}
	}

	userID := 1
	bidID := c.SoftwareDevServiceDatabase.FindUserActiveBid(userID)

	if bidID != 0 {
		_, bidCount, err = c.SoftwareDevServiceDatabase.GetSoftwareDevServicesBid(bidID)
		if err != nil {
			c.errorController(ctx, http.StatusBadRequest, err)
		}
	}

	ctx.HTML(http.StatusOK, "MainPage.html", gin.H{
		"services":      services,
		"serviceSearch": searchQuery,
		"bidCount":      len(bidCount),
	})
}

func (c *SoftwareDevServiceController) GetSoftwareDevService(ctx *gin.Context) {
	idStr := ctx.Param("serviceID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	service, err := c.SoftwareDevServiceDatabase.GetSoftwareDevService(id)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
	}

	ctx.HTML(http.StatusOK, "ServicePage.html", gin.H{
		"service": service,
	})
}

func (c *SoftwareDevServiceController) GetSoftwareDevServicesBid(ctx *gin.Context) {
	var bid ds.SoftwareDevBid
	var services []ds.SoftwareDevService
	var err error

	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	bidID := c.SoftwareDevServiceDatabase.FindUserActiveBid(userID)

	if bidID != 0 {
		bid, services, err = c.SoftwareDevServiceDatabase.GetSoftwareDevServicesBid(bidID)
		if err != nil {
			c.errorController(ctx, http.StatusInternalServerError, err)
		}

		if len(services) == 0 {
			ctx.Redirect(http.StatusSeeOther, ctx.Request.Referer())
			return
		}
	} else {
		ctx.Redirect(http.StatusSeeOther, ctx.Request.Referer())
		return
	}

	company := ctx.PostForm("company")

	var countsIDs, gradesIDs []string
	for _, service := range services {
		id := strconv.Itoa(int(service.ID))
		countsIDs = append(countsIDs, "count_"+id)
		gradesIDs = append(gradesIDs, "grade_"+id)
	}

	var counts []int
	var grades []string
	var intCount int
	var strGrade string

	for idx := range countsIDs {
		intCount, _ = strconv.Atoi(ctx.PostForm(countsIDs[idx]))
		if intCount == 0 {
			intCount = 1
		}
		counts = append(counts, intCount)

		strGrade = ctx.PostForm(gradesIDs[idx])
		if strGrade == "" {
			strGrade = "junior"
		}
		grades = append(grades, strGrade)
	}

	allServices, err := c.SoftwareDevServiceDatabase.GetSoftwareDevServices()
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
	}

	var sums []float32
	var cur_sum float32
	coefs := c.SoftwareDevServiceDatabase.GetCoefficients()
	coeffMap := make(map[string]float32)

	for _, coef := range coefs {
		coeffMap[coef.Level] = coef.Coeff
	}

	for idxService, service := range services {
		for _, oneService := range allServices {
			if service.ID == oneService.ID {
				cur_sum = service.Price * float32(counts[idxService]) * coeffMap[grades[idxService]]
				sums = append(sums, cur_sum)
			}
		}
	}

	var sum float32
	for _, service_sum := range sums {
		sum += service_sum
	}

	var keyCoefs []string
	for _, coef := range coefs {
		keyCoefs = append(keyCoefs, coef.Level)
	}

	var servicesInBid []ds.ServiceInBid

	for idx := range services {
		curServiceInBid := ds.ServiceInBid{
			Service: services[idx],
			Count:   counts[idx],
			Grade:   grades[idx],
			Sum:     int(math.Round(float64(sums[idx]))),
		}
		servicesInBid = append(servicesInBid, curServiceInBid)
	}

	companies := []string{
		"Apple", "Microsoft", "Google", "Amazon", "Tesla",
	}

	ctx.HTML(http.StatusOK, "BidPage.html", gin.H{
		"bid":       bid,
		"company":   company,
		"services":  servicesInBid,
		"companies": companies,
		"sum":       sum,
		"keyCoefs":  keyCoefs,
	})
}

func (c *SoftwareDevServiceController) AddSoftwareDevServiceToBid(ctx *gin.Context) {
	serviceID, err := strconv.Atoi(ctx.Param("serviceID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	bidID := c.SoftwareDevServiceDatabase.FindUserActiveBid(userID)

	if bidID == 0 {
		bidID = c.SoftwareDevServiceDatabase.CreateUserActiveBid(userID)
	}

	_ = c.SoftwareDevServiceDatabase.AddSoftwareDevServiceToBid(serviceID, bidID)
	ctx.Redirect(http.StatusFound, "/services")
}

func (c *SoftwareDevServiceController) DeleteSoftwareDevBid(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	bidID := c.SoftwareDevServiceDatabase.FindUserActiveBid(userID)

	result := c.SoftwareDevServiceDatabase.DeleteSoftwareDevBid(bidID)
	if result {
		ctx.Redirect(http.StatusFound, "/services")
	}
}

// func (c *SoftwareDevServiceController) CalculateSoftwareDevBid(ctx *gin.Context) {
// 	userID := ctx.Param("userID")

// 	var counts, grades []string
// 	strServicesIDs := ctx.PostFormArray("service_id[]")
// 	for _, strServiceID := range strServicesIDs {
// 		counts = append(counts, "count_"+strServiceID)
// 		grades = append(grades, "grade_"+strServiceID)
// 	}

// 	services, err := c.SoftwareDevServiceDatabase.GetSoftwareDevServices()
// 	if err != nil {
// 		c.errorController(ctx, http.StatusInternalServerError, err)
// 	}

// 	var sums []float32
// 	var cur_sum float32
// 	var cur_count int
// 	coefs := c.SoftwareDevServiceDatabase.GetCoefficients()

// 	for idxService, strServiceID := range strServicesIDs {
// 		for _, service := range services {
// 			serviceID, _ := strconv.Atoi(strServiceID)
// 			if serviceID == int(service.ID) {
// 				cur_count, _ = strconv.Atoi(counts[idxService])
// 				cur_sum = service.Price * float32(cur_count) * coefs[grades[idxService]]
// 				sums = append(sums, cur_sum)
// 			}
// 		}
// 	}

// 	ctx.Redirect(http.StatusFound, "/bids/"+userID)
// }
