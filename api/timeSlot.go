//timeSlot.go
package api

import (

    "net/http"
    "github.com/gin-gonic/gin"
)



// GetTimeSlotLimits 處理函數， 獲取所有時段限制
func GetTimeSlotLimits(c *gin.Context)  {
    limits, err := FetchTimeSlotLimits()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "無法獲取時段限制資料"})
        return
    
    }

    // 将切片转换为映射
    limitsMap := make(map[string]int)
    for _, limit := range limits {
        limitsMap[limit.TimeSlot] = limit.LimitCount
    }

    c.JSON(http.StatusOK, limitsMap)
}





// CreateTimeSlotLimit  創建新的時段限制
func CreateTimeSlotLimit(c *gin.Context)  {
    var limit TimeSlotLimit
    if err := c.ShouldBind(&limit); err != nil {
         c.JSON(http.StatusBadRequest, gin.H{"error": "格式錯誤"})
         return
    }

    // 将单个对象转换为切片
    limits := TimeSlotLimits{limit}

    if err := InsertTimeSlotLimits(limits); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "無法創建時段限制"})
        return
    }
     c.JSON(http.StatusCreated, limit)
}


// UpdateTimeSlotLimit  更新現有的時段限制
func UpdateTimeSlotLimit(c *gin.Context)  {
    var limits map[string]int
    if err := c.ShouldBind(&limits); err != nil {
        c.JSON(http.StatusBadRequest,gin.H{"error": "無效輸入"})
        return
    }

    if err := UpdateExistingTimeSlotLimits(limits); err != nil {
       c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
       return 
    }
    c.JSON(http.StatusOK, gin.H{"result": "更新成功"})
}
