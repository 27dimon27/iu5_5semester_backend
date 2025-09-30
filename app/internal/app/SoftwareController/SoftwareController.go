package SoftwareController

import (
	"errors"
	"io"
	"math"
	"software/internal/app/SoftwareDatabase"
	"software/internal/app/ds"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SoftwareController struct {
	SoftwareDatabase *SoftwareDatabase.SoftwareDatabase
}

func NewSoftwareController(d *SoftwareDatabase.SoftwareDatabase) *SoftwareController {
	return &SoftwareController{
		SoftwareDatabase: d,
	}
}

func (c *SoftwareController) RegisterController(router *gin.Engine) {
	router.GET("/softwares", c.GetSoftwareServices)
	router.GET("/softwares/:softwareID", c.GetSoftwareService)
	router.GET("/software-bids/:softwareBidID", c.GetSoftwareServicesBid)

	router.POST("/:softwareBidID/add-software/:softwareID", c.AddSoftwareServiceToBid)
	router.POST("/software-bids/delete-software-bid/:softwareBidID", c.SoftDeleteSoftwareBid)

	router.POST("/software-bids/calc-software-bid/:softwareBidID", c.GetSoftwareServicesBid) // ВОЗМОЖНО СВЯЗАТЬ С /update-software-in-bid

	router.GET("/api/softwares", c.GetAllSoftwareServices)
	router.GET("/api/softwares/:softwareID", c.GetSoftwareServiceByID)
	router.POST("/api/softwares", c.AddNewSoftware)
	router.PUT("/api/softwares/:softwareID", c.UpdateSoftware)
	router.DELETE("/api/softwares/:softwareID", c.DeleteSoftware) // добавить удаление фото
	router.POST("/api/add-software/:softwareID", c.AddSoftwareServiceToBidByID)
	router.POST("/api/add-photo/:softwareID", c.AddPhotoToSoftwareService) // реализовать добавление фото

	router.GET("/api/software-bids-icon", c.GetSoftwareServiceCountInBid)
	router.GET("/api/software-bids", c.GetSoftwareBids)
	router.GET("/api/software-bids/:softwareBidID", c.GetSoftwareBidByID)
	router.PUT("/api/software-bids", c.UpdateActiveSoftwareBid)
	router.PUT("/api/formation-software-bids", c.FormateActiveSoftwareBid)
	router.PUT("/api/moderator-software-bids/:softwareBidID", c.ModerateSoftwareBid)
	router.DELETE("/api/software-bids/:softwareBidID", c.DeleteSoftwareBid)

	router.DELETE("/api/delete-software-from-bid/:softwareID", c.DeleteSoftwareFromBid)
	router.PUT("/api/update-software-in-bid", c.UpdateSoftwareInBid)

	router.POST("/api/registration", c.RegisterNewUser)
	router.GET("/api/account/:userID", c.GetUserAccountData)
	router.PUT("/api/account/:userID", c.UpdateUserAccountData)
	router.POST("/api/authentication", c.AuthenticateUser)
	router.POST("/api/deauthorization/:userID", c.DeauthorizeUser)
}

func (c *SoftwareController) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (c *SoftwareController) errorController(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

func (c *SoftwareController) GetSoftwareServices(ctx *gin.Context) {
	var services []ds.SoftwareService
	var bidCount []ds.SoftwareService
	var bidID int
	var err error

	searchQuery := ctx.Query("software-search")
	if searchQuery == "" {
		services, err = c.SoftwareDatabase.GetSoftwareServices()
		if err != nil {
			c.errorController(ctx, http.StatusInternalServerError, err)
		}
	} else {
		services, err = c.SoftwareDatabase.GetSoftwareServicesByTitle(searchQuery)
		if err != nil {
			c.errorController(ctx, http.StatusInternalServerError, err)
		}
	}

	userID := c.SoftwareDatabase.SingletonGetCreator()
	bidID, err = c.SoftwareDatabase.FindUserActiveBid(userID)

	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	if bidID != 0 {
		_, bidCount, err = c.SoftwareDatabase.GetSoftwareServicesBid(bidID)
		if err != nil {
			c.errorController(ctx, http.StatusBadRequest, err)
		}
	}

	ctx.HTML(http.StatusOK, "MainPage.html", gin.H{
		"services":       services,
		"softwareSearch": searchQuery,
		"bidCount":       len(bidCount),
		"bidID":          bidID,
	})
}

func (c *SoftwareController) GetSoftwareService(ctx *gin.Context) {
	idStr := ctx.Param("softwareID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	service, err := c.SoftwareDatabase.GetSoftwareService(id)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
	}

	ctx.HTML(http.StatusOK, "ServicePage.html", gin.H{
		"service": service,
	})
}

func (c *SoftwareController) GetSoftwareServicesBid(ctx *gin.Context) {
	var bid ds.SoftwareBid
	var services []ds.SoftwareService
	var err error

	bidID, err := strconv.Atoi(ctx.Param("softwareBidID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	// bidID := c.SoftwareDatabase.FindUserActiveBid(userID)

	if bidID != 0 {
		bid, services, err = c.SoftwareDatabase.GetSoftwareServicesBid(bidID)
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

	allServices, err := c.SoftwareDatabase.GetSoftwareServices()
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
	}

	var sums []float32
	var cur_sum float32
	coefs := c.SoftwareDatabase.GetCoefficients()
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
		"bidID":     bidID,
	})
}

func (c *SoftwareController) AddSoftwareServiceToBid(ctx *gin.Context) {
	serviceID, err := strconv.Atoi(ctx.Param("softwareID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}
	bidID, err := strconv.Atoi(ctx.Param("softwareBidID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	userID := c.SoftwareDatabase.SingletonGetCreator()
	// bidID := c.SoftwareDatabase.FindUserActiveBid(userID)

	if bidID == 0 {
		bidID = c.SoftwareDatabase.CreateUserActiveBid(userID)
	}

	_, err = c.SoftwareDatabase.AddSoftwareServiceToBid(serviceID, bidID)
	if err != nil {
		if err.Error() == "duplicate" {
			c.errorController(ctx, http.StatusBadRequest, err)
		}
		c.errorController(ctx, http.StatusInternalServerError, err)
	}
	ctx.Redirect(http.StatusFound, "/softwares")
}

func (c *SoftwareController) SoftDeleteSoftwareBid(ctx *gin.Context) {
	bidID, err := strconv.Atoi(ctx.Param("softwareBidID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	// bidID := c.SoftwareDatabase.FindUserActiveBid(userID)

	result := c.SoftwareDatabase.SoftDeleteSoftwareBid(bidID)
	if result {
		ctx.Redirect(http.StatusFound, "/softwares")
	}
}

func (c *SoftwareController) GetAllSoftwareServices(ctx *gin.Context) {
	var services []ds.SoftwareService
	var err error

	searchQuery := ctx.Query("software-search")
	if searchQuery == "" {
		services, err = c.SoftwareDatabase.GetSoftwareServices()
		if err != nil {
			c.errorController(ctx, http.StatusInternalServerError, err)
			return
		}
	} else {
		services, err = c.SoftwareDatabase.GetSoftwareServicesByTitle(searchQuery)
		if err != nil {
			c.errorController(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":   true,
		"action":    "selection",
		"softwares": services,
	})
}

func (c *SoftwareController) GetSoftwareServiceByID(ctx *gin.Context) {
	idStr := ctx.Param("softwareID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	service, err := c.SoftwareDatabase.GetSoftwareService(id)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":  true,
		"action":   "selection",
		"software": service,
	})
}

func (c *SoftwareController) AddNewSoftware(ctx *gin.Context) {
	software := ds.SoftwareService{}
	softwareID := 0
	err := ctx.ShouldBindJSON(&software)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	softwareID, err = c.SoftwareDatabase.AddNewSoftware(software)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":    true,
		"action":     "insert",
		"softwareID": softwareID,
	})
}

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

	softwareID, err = c.SoftwareDatabase.UpdateSoftware(softwareID, software)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":    true,
		"action":     "update",
		"softwareID": softwareID,
	})
}

func (c *SoftwareController) DeleteSoftware(ctx *gin.Context) {
	strSoftwareID := ctx.Param("softwareID")
	softwareID, err := strconv.Atoi(strSoftwareID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	softwareID, err = c.SoftwareDatabase.DeleteSoftware(softwareID)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":    true,
		"action":     "delete",
		"softwareID": softwareID,
	})
}

func (c *SoftwareController) AddSoftwareServiceToBidByID(ctx *gin.Context) {
	strSoftwareID := ctx.Param("softwareID")
	softwareID, err := strconv.Atoi(strSoftwareID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	var bidID int

	creatorID := c.SoftwareDatabase.SingletonGetCreator()
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
		"success":    true,
		"action":     "insert in bid",
		"softwareID": softwareID,
		"bidID":      bidID,
	})
}

func (c *SoftwareController) AddPhotoToSoftwareService(ctx *gin.Context) {

}

func (c *SoftwareController) GetSoftwareServiceCountInBid(ctx *gin.Context) {
	var services []ds.SoftwareService

	creatorID := c.SoftwareDatabase.SingletonGetCreator()
	bidID, err := c.SoftwareDatabase.FindUserActiveBid(creatorID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	_, services, err = c.SoftwareDatabase.GetSoftwareServicesBid(bidID)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bidID": bidID,
		"count": len(services),
	})
}

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

	creatorID := c.SoftwareDatabase.SingletonGetCreator()
	bids, err = c.SoftwareDatabase.GetSoftwareBids(creatorID, filter)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bids": bids,
	})
}

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

func (c *SoftwareController) UpdateActiveSoftwareBid(ctx *gin.Context) {
	var bid ds.SoftwareBid
	err := ctx.ShouldBindJSON(&bid)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	creatorID := c.SoftwareDatabase.SingletonGetCreator()
	bidID, err := c.SoftwareDatabase.FindUserActiveBid(creatorID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	bidID, err = c.SoftwareDatabase.UpdateActiveSoftwareBid(bidID, bid)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"action":  "update",
		"bidID":   bidID,
	})
}

func (c *SoftwareController) FormateActiveSoftwareBid(ctx *gin.Context) {
	var servicesInBid []ds.Service_n_Bid

	creatorID := c.SoftwareDatabase.SingletonGetCreator()
	bidID, err := c.SoftwareDatabase.FindUserActiveBid(creatorID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	servicesInBid, err = c.SoftwareDatabase.CountServicesInBid(bidID)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	for _, service := range servicesInBid {
		if service.Count == 0 {
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
		"success": true,
		"action":  "formate",
		"bidID":   bidID,
	})
}

func (c *SoftwareController) ModerateSoftwareBid(ctx *gin.Context) {
	bidID, err := strconv.Atoi(ctx.Param("softwareBidID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	approved := false
	if len(ctx.Query("approved")) > 0 {
		approved = true
	}

	bidID, err = c.SoftwareDatabase.ModerateSoftwareBid(bidID, approved)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"action":  "moderate",
		"bidID":   bidID,
	})
}

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
		"success": true,
		"action":  "delete",
		"bidID":   bidID,
	})
}

func (c *SoftwareController) DeleteSoftwareFromBid(ctx *gin.Context) {
	softwareID, err := strconv.Atoi(ctx.Param("softwareID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	creatorID := c.SoftwareDatabase.SingletonGetCreator()
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
		"success":    true,
		"action":     "delete",
		"bidID":      bidID,
		"softwareID": softwareID,
	})
}

func (c *SoftwareController) UpdateSoftwareInBid(ctx *gin.Context) {
	softwaresInBid := []map[string]int{}

	err := ctx.ShouldBindJSON(&softwaresInBid)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	creatorID := c.SoftwareDatabase.SingletonGetCreator()
	bidID, err := c.SoftwareDatabase.FindUserActiveBid(creatorID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	err = c.SoftwareDatabase.UpdateSoftwareInBid(bidID, softwaresInBid)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"action":  "update",
		"bidID":   bidID,
	})
}

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
		"success":  true,
		"action":   "create",
		"username": newUser.Login,
	})
}

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
		"success":  true,
		"action":   "select",
		"userData": userData,
	})
}

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

	err = c.SoftwareDatabase.UpdateUserAccountData(userID, updateUser)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"action":  "update",
		"userID":  userID,
	})
}

func (c *SoftwareController) AuthenticateUser(ctx *gin.Context) {
	type UserData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var userData UserData

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

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"action":  "authenticate",
		"user":    foundUser,
	})
}

func (c *SoftwareController) DeauthorizeUser(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	err = c.SoftwareDatabase.DeauthorizeUser(userID)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"action":  "deauthorize",
		"userID":  userID,
	})
}
