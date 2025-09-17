package api

import (
	"lab1/internal/app/SoftwareDevServiceController"
	"lab1/internal/app/SoftwareDevServiceDatabase"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Server start up")

	database, err := SoftwareDevServiceDatabase.NewSoftwareDevServiceDatabase()
	if err != nil {
		logrus.Error("array initialization error")
	}

	controller := SoftwareDevServiceController.NewSoftwareDevServiceController(database)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/services", controller.GetSoftwareDevServices)
	r.GET("/services/:idService", controller.GetSoftwareDevService)
	r.GET("/bid/:idBid", controller.GetSoftwareDevServicesBid)

	r.Run()

	log.Println("Server down")
}
