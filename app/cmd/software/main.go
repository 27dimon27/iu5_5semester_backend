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

// @title Software Development API
// @version 1.0
// @description API для системы разработки программного обеспечения
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:80
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT токен в формате: "Bearer {token}"
func main() {
	router := gin.Default()
	router.Use(cors.Default())
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := SoftwareDatabase.NewSoftwareAndPhotoDatabase(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := SoftwareController.NewSoftwareController(rep)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
