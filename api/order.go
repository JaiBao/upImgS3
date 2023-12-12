//order.go
package api

import (


    "net/http"
    "github.com/gin-gonic/gin"
)


// GetOrderByCriteria 根據條件查詢訂單
func GetOrderByCriteria(c *gin.Context)  {
    code := c.Query("code")
    mobile := c.Query("mobile")
    personalName := c.Query("personal_name")

    criteria := make(map[string]string)
    if code != "" {
        criteria["code"] = code
    }
    if mobile != "" {
        criteria["mobile"] = mobile
    }
    if personalName != "" {
        criteria["personal_name"] = personalName
    }

    if len(criteria) < 2 {
         c.JSON(http.StatusBadRequest, gin.H{"error": "請提供至少兩個查詢條件"})
         return
    }

    orders, err := FetchOrderByCriteria(criteria)
    if err != nil {
         c.JSON(http.StatusInternalServerError, gin.H{"error": "無法查詢訂單"})
         return
    }

    if len(orders) == 0 {
         c.JSON(http.StatusNotFound, gin.H{"error": "沒有該筆訂單"})
         return
    }

     c.JSON(http.StatusOK, orders)
}



