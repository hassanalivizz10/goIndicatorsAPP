package apiroutes

import (

  
    "github.com/gin-gonic/gin"
    ///"indicatorsAPP/api"
)

func SetupRouter() *gin.Engine {
    router := gin.Default()

    // Define API routes and attach handlers
    //router.POST("/createUpdateDailyIndicators",api.CreateUpdateDailyIndicatorsHandler)
    // router.GET("/api/resource", handler.GetResource)
    // Add more routes here

    return router
}

func StartServer() {
    router := SetupRouter()
    // Configure server settings (e.g., port)
    router.Run(":3300")
}


