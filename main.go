package main

import (
	"github.com/gin-gonic/gin"
	"github.com/matheusfillipe/girafiles/api"
)

func main() {
	var settings = api.GetSettings()
	if !settings.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	api.StartServer()
}
