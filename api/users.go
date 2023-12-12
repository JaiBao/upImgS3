//users.go
package api

import (
    "net/http"
    "github.com/gin-gonic/gin"
)





// GetUserByID 處理函數，根據姓名和手機號碼獲取使用者 ID
func GetUserByID(c *gin.Context)  {
    mobile := c.Query("tel") // 從查詢參數中獲取手機號碼
    name := c.Query("name")  // 從查詢參數中獲取姓名

    userID, err := FetchUserIDByNameAndMobile(name, mobile)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "無法獲取使用者資料"})
        return
    }

    // 如果 userID 為空，表示沒有找到對應的使用者
    if userID == "" {
      c.JSON(http.StatusOK, gin.H{"id": ""})
      return
    }

     c.JSON(http.StatusOK, gin.H{"id": userID})
}

