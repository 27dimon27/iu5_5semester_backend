package pkg

import (
	"fmt"

	"software/internal/app/SoftwareController"
	"software/internal/app/config"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config     *config.Config
	Router     *gin.Engine
	Controller *SoftwareController.SoftwareController
}

func NewApp(config *config.Config, router *gin.Engine, controller *SoftwareController.SoftwareController) *Application {
	return &Application{
		Config:     config,
		Router:     router,
		Controller: controller,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	a.Controller.RegisterController(a.Router)
	a.Controller.RegisterStatic(a.Router)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.Host, a.Config.Port)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}
