//specialDate.go
package api

import (
    "net/http"
    "github.com/gin-gonic/gin"
)






// GetSpecificDateLimits 處理函數，獲取特定日期的時段限制
func GetSpecificDateLimits(c *gin.Context)  {
    yearMonth := c.Query("month") // 從查詢參數中獲取月份，例如 "2024-01"
    specificDate := c.Query("date") // 新增：從查詢參數中獲取具體日期，例如 "2024-01-02"

    allLimits, err := FetchSpecificDateLimits()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "無法獲取特定日期的時間限制數據"})
        return 
    }

    // 如果提供了具體日期參數，則只返回該日期的數據
    if specificDate != "" {
        yearMonthOfDate := specificDate[:7] // 從日期獲取年月
        if dates, ok := allLimits[yearMonthOfDate]; ok {
            if dateLimits, ok := dates[specificDate]; ok {
                 c.JSON(http.StatusOK, gin.H{specificDate: dateLimits})
                 return
            }
        }
         c.JSON(http.StatusNotFound, gin.H{"error": "未找到指定日期的數據"})
         return
    }

    // 如果提供了月份參數，則只返回該月份的數據
    if yearMonth != "" {
        if limits, ok := allLimits[yearMonth]; ok {
             c.JSON(http.StatusOK, limits)
             return
        }
         c.JSON(http.StatusNotFound, gin.H{"error": "未找到指定月份的數據"})
         return
    }

    // 如果沒有提供月份或日期參數，返回所有數據
     c.JSON(http.StatusOK, allLimits)
}



// CreateSpecificDateLimit  創建特定日期的時段限制
func CreateSpecificDateLimit(c *gin.Context)  {
    var dateLimits map[string]map[string]int
    if err := c.ShouldBind(&dateLimits); err != nil {
  
        c.JSON(http.StatusBadRequest, gin.H{"error": "格式錯誤"})
        return 
    }

    for date, limits := range dateLimits {
        dateLimit := SpecificDateLimit{
            Date:       date,
            TimeLimits: limits,
        }
        if err := InsertSpecificDateLimit(dateLimit); err != nil {
             c.JSON(http.StatusInternalServerError, gin.H{"error": "無法創建特定日期的時段限制"})
             return
        }
    }
     c.JSON(http.StatusCreated, dateLimits)
}

// UpdateSpecificDateLimit 更新特定日期的一個時段或多個限制

func UpdateSpecificDateLimit(c *gin.Context)  {
    var dateLimits map[string]map[string]int
    if err := c.ShouldBind(&dateLimits); err != nil {

         c.JSON(http.StatusBadRequest, gin.H{"error": "格式錯誤"})
         return
    }

    for date, timeLimits := range dateLimits {
        for timeSlot, limitCount := range timeLimits {
            // 檢查原本是否有這時段紀錄
            if exists, err := checkDateLimitExists(date, timeSlot); err != nil {
                 c.JSON(http.StatusInternalServerError, gin.H{"error": "檢查時發生錯誤"})
                 return
            } else if exists {
                // 有的話就改
                err := UpdateDateLimit(date, timeSlot, limitCount)
                if err != nil {
                     c.JSON(http.StatusInternalServerError, gin.H{"error": "加入失敗"})
                     return
                }
            } else {
                // 沒有的話給錯誤或插入新時段
                 c.JSON(http.StatusNotFound, gin.H{"error": "沒有這個時段設定"})
                 return
         
            }
        }
    }
     c.JSON(http.StatusOK, gin.H{"result": "已加入訂單"})
}

// checkDateLimitExists 檢查日期時段有無
func checkDateLimitExists(date string, timeSlot string) (bool, error) {
    var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM DateLimits WHERE Date = ? AND TimeSlot = ?)", date, timeSlot).Scan(&exists)
    if err != nil {
        return false, err
    }
    return exists, nil
}


// UpdateDateLimit 更新特定時段
func UpdateDateLimit(date string, timeSlot string, limitCount int) error  {
    stmt, err := db.Prepare("UPDATE DateLimits SET LimitCount = ? WHERE Date = ? AND TimeSlot = ?")
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(limitCount, date, timeSlot)
    if err != nil {
        return err
    }

    return nil
}

