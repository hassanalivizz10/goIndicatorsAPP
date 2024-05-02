package routes

import (
	"indicatorsAPP/api"
	"indicatorsAPP/utils"

	"github.com/gin-gonic/gin"
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
	router.POST("/apiEndPoint/getLatestTrendData", api.GetCurrentTrendValue)

	// router.GET("/api/resource", handler.GetResource)
	// Add more routes here

	return router
}

func StartServer() {
	port := utils.GetEnv("APP_PORT")
	if port == "" {
		port = "2607"
	}
	router := SetupRouter()
	// Configure server settings (e.g., port)
	router.Run(":" + port)
}
