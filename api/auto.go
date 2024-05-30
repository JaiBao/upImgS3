// auto.go
package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TriggerAutoCreateLimits 自動創建預設日期
func TriggerAutoCreateLimits(c *gin.Context)  {
    period := c.Query("add") // 從查詢參數獲取時間範圍oneWeek、twoWeeks、oneMonth
    coverStr :=c.Query("cover")
    //預設覆蓋false
    cover :=false
    if coverStr == "true"{
        cover = true
    }

    // 如果沒有提供時間範圍，則預設為兩個月
    if period == "" {
        period = "twoMonths"
    }

    err := AutoCreateNextTwoMonthsLimits(period , cover)
    if err != nil {
         c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
         return
    }
     c.JSON(http.StatusOK, gin.H{"message": "已成功新增時段預設"})
     
}


var schedulerActive bool
var ticker *time.Ticker

// StartSchedulerHandler 定時新增
func StartSchedulerHandler(c *gin.Context)  {
    if !schedulerActive {
        StartScheduler()
        schedulerActive = true
         c.JSON(http.StatusOK,gin.H{"message": "定時自動新增已啟動"})
         return
    }
     c.JSON(http.StatusBadRequest, gin.H{"error": "定時任務進行中"})
}

// StopSchedulerHandler 停止定時
func StopSchedulerHandler(c *gin.Context)  {
    if schedulerActive {
        StopScheduler()
        schedulerActive = false
         c.JSON(http.StatusOK,gin.H{"message": "定時任務已停止"})
         return
    }
     c.JSON(http.StatusBadRequest, gin.H{"error": "定時任務未運作"})
}

func StartScheduler() {
    // 計算現在時間到今天早上6點的時間差
    now := time.Now()
    next := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, now.Location())
    if now.After(next) {
        // 如果現在時間已經過了今天的6點，那麼設定下一個觸發時間為明天6點
        next = next.Add(24 * time.Hour)
    }

    // 計算時間差
    duration := next.Sub(now)

    // 設置一個一次性的定時器，在今天或明天的早上6點觸發
    time.AfterFunc(duration, func() {
        AutoCreateNextTwoMonthsLimits("oneMonth", false)
        // 現在設置每24小時觸發一次的定時器
        ticker = time.NewTicker(24 * time.Hour)
        go func() {
            for {
                select {
                case <-ticker.C:
                    AutoCreateNextTwoMonthsLimits("oneMonth", false)
                }
            }
        }()
    })
}


func StopScheduler() {
    if ticker != nil {
        ticker.Stop()
    }
}

// GetSchedulerStatusHandler 定時的狀態
func GetSchedulerStatusHandler(c *gin.Context)  {
    status := gin.H{"schedulerActive": schedulerActive}
     c.JSON(http.StatusOK, status)
}




