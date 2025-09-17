package main

import (
	"fmt"

	"softwareDev/internal/app/SoftwareDevServiceController"
	"softwareDev/internal/app/SoftwareDevServiceDatabase"
	"softwareDev/internal/app/config"
	"softwareDev/internal/app/dsn"
	"softwareDev/internal/pkg"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := SoftwareDevServiceDatabase.NewSoftwareDevServiceDatabase(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := SoftwareDevServiceController.NewSoftwareDevServiceController(rep)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
