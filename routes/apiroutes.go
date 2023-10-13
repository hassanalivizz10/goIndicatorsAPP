package routes

import (
	"github.com/gin-gonic/gin"
	"indicatorsAPP/api"
)

func SetupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
        c.String(200, "Hello, world!")
    })
	// Define API routes and attach handlers
	router.POST("/apiEndPoint/createUpdateDailyIndicators", api.CreateUpdateDailyIndicatorsHandler)
	router.POST("/apiEndPoint/setHourlyIndicators", api.SetHourlyIndicatorsHandler)
	router.POST("/apiEndPoint/fetchTradingDataByCoin", api.FetchTradingDataByCoinHandler)
	// router.GET("/api/resource", handler.GetResource)
	// Add more routes here

	return router
}

func StartServer() {
	router := SetupRouter()
	// Configure server settings (e.g., port)
	router.Run(":2608")
}
