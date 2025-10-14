package SoftwareController

import (
	"errors"
	"io"
	"log"
	"math"
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
	// router.GET("/softwares", c.GetSoftwareServices)
	// router.GET("/softwares/:softwareID", c.GetSoftwareService)
	// router.GET("/software-bids/:softwareBidID", c.GetSoftwareServicesBid)
	// router.POST("/:softwareBidID/add-software/:softwareID", c.AddSoftwareServiceToBid)
	// router.POST("/software-bids/delete-software-bid/:softwareBidID", c.SoftDeleteSoftwareBid)

	// router.POST("/software-bids/calc-software-bid/:softwareBidID", c.GetSoftwareServicesBid)

	Guest := router.Group("/")
	Guest.Use(middlewares.AuthMiddleware(c.SoftwareDatabase.RedisClient, role.Guest, role.User, role.Admin))
	{
		Guest.GET("/api/softwares", c.GetAllSoftwareServices)
		Guest.GET("/api/softwares/:softwareID", c.GetSoftwareServiceByID)
		Guest.POST("/api/registration", c.RegisterNewUser)
		Guest.POST("/api/authentication", c.AuthenticateUser)
	}

	User := router.Group("/")
	User.Use(middlewares.AuthMiddleware(c.SoftwareDatabase.RedisClient, role.User, role.Admin))
	{
		User.POST("/api/add-software/:softwareID", c.AddSoftwareServiceToBidByID)
		User.GET("/api/software-bids-icon", c.GetSoftwareServiceCountInBid)
		User.GET("/api/software-bids", c.GetSoftwareBids)
		User.GET("/api/software-bids/:softwareBidID", c.GetSoftwareBidByID)
		User.PUT("/api/software-bids", c.UpdateActiveSoftwareBid)
		User.PUT("/api/formation-software-bids", c.FormateActiveSoftwareBid)
		User.DELETE("/api/software-bids/:softwareBidID", c.DeleteSoftwareBid)
		User.DELETE("/api/delete-software-from-bid/:softwareID", c.DeleteSoftwareFromBid)
		User.PUT("/api/update-software-in-bid", c.UpdateSoftwareInBid)
		User.GET("/api/account/:userID", c.GetUserAccountData)
		User.PUT("/api/account/:userID", c.UpdateUserAccountData)
		User.POST("/api/deauthorization", c.DeauthorizeUser)
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

	// router.GET("/api/softwares", c.GetAllSoftwareServices)
	// router.GET("/api/softwares/:softwareID", c.GetSoftwareServiceByID)
	// router.POST("/api/softwares", c.AddNewSoftware)
	// router.PUT("/api/softwares/:softwareID", c.UpdateSoftware)
	// router.DELETE("/api/softwares/:softwareID", c.DeleteSoftwareWithPhoto)
	// router.POST("/api/add-software/:softwareID", c.AddSoftwareServiceToBidByID)
	// router.POST("/api/add-photo/:softwareID", c.AddPhotoToSoftwareService)

	// router.GET("/api/software-bids-icon", c.GetSoftwareServiceCountInBid)
	// router.GET("/api/software-bids", c.GetSoftwareBids)
	// router.GET("/api/software-bids/:softwareBidID", c.GetSoftwareBidByID)
	// router.PUT("/api/software-bids", c.UpdateActiveSoftwareBid)
	// router.PUT("/api/formation-software-bids", c.FormateActiveSoftwareBid)
	// router.PUT("/api/moderator-software-bids/:softwareBidID", c.ModerateSoftwareBid)
	// router.DELETE("/api/software-bids/:softwareBidID", c.DeleteSoftwareBid)

	// router.DELETE("/api/delete-software-from-bid/:softwareID", c.DeleteSoftwareFromBid)
	// router.PUT("/api/update-software-in-bid", c.UpdateSoftwareInBid)

	// router.POST("/api/registration", c.RegisterNewUser)
	// router.GET("/api/account/:userID", c.GetUserAccountData)
	// router.PUT("/api/account/:userID", c.UpdateUserAccountData)
	// router.POST("/api/authentication", c.AuthenticateUser)
	// router.POST("/api/deauthorization/:userID", c.DeauthorizeUser)

	// router.Use(middlewares.AuthMiddleware()).POST("/ping", pong)
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

func (c *SoftwareController) GetSoftwareServices(ctx *gin.Context) {
	var softwares []ds.SoftwareService
	var bidCount []ds.SoftwareService
	var bidID int
	var err error

	searchQuery := ctx.Query("software-search")
	if searchQuery == "" {
		softwares, err = c.SoftwareDatabase.GetSoftwareServices()
		if err != nil {
			c.errorController(ctx, http.StatusInternalServerError, err)
		}
	} else {
		softwares, err = c.SoftwareDatabase.GetSoftwareServicesByTitle(searchQuery)
		if err != nil {
			c.errorController(ctx, http.StatusInternalServerError, err)
		}
	}

	userID := c.SoftwareDatabase.SingletonGetCreator()
	bidID, err = c.SoftwareDatabase.FindUserActiveBid(userID)

	if err != nil && err.Error() != "record not found" {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	if bidID != 0 {
		_, bidCount, err = c.SoftwareDatabase.GetSoftwareServicesBid(bidID)
		if err != nil {
			c.errorController(ctx, http.StatusBadRequest, err)
		}
	}

	ctx.HTML(http.StatusOK, "MainPage.html", gin.H{
		"services":       softwares,
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

	software, err := c.SoftwareDatabase.GetSoftwareService(id)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
	}

	ctx.HTML(http.StatusOK, "ServicePage.html", gin.H{
		"service": software,
	})
}

func (c *SoftwareController) GetSoftwareServicesBid(ctx *gin.Context) {
	var bid ds.SoftwareBid
	var softwares []ds.SoftwareService
	var err error

	bidID, err := strconv.Atoi(ctx.Param("softwareBidID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	// bidID := c.SoftwareDatabase.FindUserActiveBid(userID)

	if bidID != 0 {
		bid, softwares, err = c.SoftwareDatabase.GetSoftwareServicesBid(bidID)
		if err != nil {
			c.errorController(ctx, http.StatusInternalServerError, err)
		}

		if len(softwares) == 0 {
			ctx.Redirect(http.StatusSeeOther, ctx.Request.Referer())
			return
		}
	} else {
		ctx.Redirect(http.StatusSeeOther, ctx.Request.Referer())
		return
	}

	company := ctx.PostForm("company")

	var countsIDs, gradesIDs []string
	for _, software := range softwares {
		id := strconv.Itoa(int(software.ID))
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

	allSoftwares, err := c.SoftwareDatabase.GetSoftwareServices()
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

	for idxSoftware, software := range softwares {
		for _, oneSoftware := range allSoftwares {
			if software.ID == oneSoftware.ID {
				cur_sum = software.Price * float32(counts[idxSoftware]) * coeffMap[grades[idxSoftware]]
				sums = append(sums, cur_sum)
			}
		}
	}

	var sum float32
	for _, software_sum := range sums {
		sum += software_sum
	}

	var keyCoefs []string
	for _, coef := range coefs {
		keyCoefs = append(keyCoefs, coef.Level)
	}

	var softwaresInBid []ds.SoftwareServiceInSoftwareBid

	for idx := range softwares {
		curSoftwareInBid := ds.SoftwareServiceInSoftwareBid{
			SoftwareService: softwares[idx],
			Count:           counts[idx],
			Grade:           grades[idx],
			Sum:             int(math.Round(float64(sums[idx]))),
		}
		softwaresInBid = append(softwaresInBid, curSoftwareInBid)
	}

	companies := []string{
		"Apple", "Microsoft", "Google", "Amazon", "Tesla",
	}

	ctx.HTML(http.StatusOK, "BidPage.html", gin.H{
		"bid":       bid,
		"company":   company,
		"services":  softwaresInBid,
		"companies": companies,
		"sum":       sum,
		"keyCoefs":  keyCoefs,
		"bidID":     bidID,
	})
}

func (c *SoftwareController) AddSoftwareServiceToBid(ctx *gin.Context) {
	softwareID, err := strconv.Atoi(ctx.Param("softwareID"))
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

	_, err = c.SoftwareDatabase.AddSoftwareServiceToBid(softwareID, bidID)
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

func (c *SoftwareController) GetSoftwareServiceCountInBid(ctx *gin.Context) {
	var softwares []ds.SoftwareService

	// creatorID := c.SoftwareDatabase.SingletonGetCreator()
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

	// creatorID := c.SoftwareDatabase.SingletonGetCreator()
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

	// creatorID := c.SoftwareDatabase.SingletonGetCreator()
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

func (c *SoftwareController) FormateActiveSoftwareBid(ctx *gin.Context) {
	var softwaresInBid []ds.SoftwareService_n_SoftwareBid

	// creatorID := c.SoftwareDatabase.SingletonGetCreator()
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

	bidID, cost, err := c.SoftwareDatabase.ModerateSoftwareBid(bidID, approved)
	if err != nil {
		c.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"cost":  cost,
		"bidID": bidID,
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
		"bidID": bidID,
	})
}

func (c *SoftwareController) DeleteSoftwareFromBid(ctx *gin.Context) {
	softwareID, err := strconv.Atoi(ctx.Param("softwareID"))
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	// creatorID := c.SoftwareDatabase.SingletonGetCreator()
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

func (c *SoftwareController) UpdateSoftwareInBid(ctx *gin.Context) {
	softwareInBid := ds.SoftwareService_n_SoftwareBid{}

	err := ctx.ShouldBindJSON(&softwareInBid)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
		return
	}

	// creatorID := c.SoftwareDatabase.SingletonGetCreator()
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

	token, err := jwts.GenerateToken(foundUser)
	if err != nil {
		c.errorController(ctx, http.StatusBadRequest, err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":  foundUser,
		"token": token,
	})
}

func (c *SoftwareController) DeauthorizeUser(ctx *gin.Context) {
	// userID, err := strconv.Atoi(ctx.Param("userID"))
	// if err != nil {
	// 	c.errorController(ctx, http.StatusBadRequest, err)
	// 	return
	// }

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

	// err = c.SoftwareDatabase.DeauthorizeUser(ctx, jwtStr)
	// if err != nil {
	// 	c.errorController(ctx, http.StatusBadRequest, err)
	// 	return
	// }

	err = c.SoftwareDatabase.RedisClient.WriteJWTToBlacklist(ctx.Request.Context(), jwtStr, time.Hour*24)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusOK)
}

// ПРОТЕСТИРОВАТЬ НОВЫЕ ВАРИКИ РУЧЕК С НОВЫМИ ВОЗВРАТАМИ (УСЛУГИ, ЗАЯВКИ, ЮЗЕРЫ, И Т.Д.)
