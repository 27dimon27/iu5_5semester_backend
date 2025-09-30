package main

import (
	"fmt"

	"software/internal/app/SoftwareController"
	"software/internal/app/SoftwareDatabase"
	"software/internal/app/config"
	"software/internal/app/dsn"
	"software/internal/pkg"

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

	rep, errRep := SoftwareDatabase.NewSoftwareDatabase(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := SoftwareController.NewSoftwareController(rep)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
