// delivery.go

package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetDeliveryByOrderCode 根據訂單編號查外送員
func GetDeliveryByOrderCode(c *gin.Context) {
    orderCodeStr := c.Param("order_code")
    orderCode, err := strconv.Atoi(orderCodeStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "無此訂單編號"})
        return
    }

    delivery, err := FetchDelivery(orderCode)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "無法查詢外送員訊息"})
        return
    }
    if delivery == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "沒有找到相關外送員訊息"})
        return
    }

    c.JSON(http.StatusOK, delivery)
}
