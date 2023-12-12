//road.go
package api

import (

    "strconv"
    "net/http"
    "github.com/gin-gonic/gin"
)

// GetRoadsByCityID 根據城市 ID 獲取路名
func GetRoadsByCityID(c *gin.Context)  {
    cityIDParam := c.Query("city_id")
    cityID, err := strconv.Atoi(cityIDParam)
    if err != nil {
         c.JSON(http.StatusBadRequest, gin.H{"error": "無效的城市 ID"})
         return
    }

    roads, err := FetchRoadsByCityID(cityID)
    if err != nil {
         c.JSON(http.StatusInternalServerError, gin.H{"error": "無法獲取路名數據"})
         return
    }

     c.JSON(http.StatusOK, roads)
}

