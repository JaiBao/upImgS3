// routes.go

package api

import (

	"github.com/gin-gonic/gin"
)

// LoadRoutes  加載 API 路由
func LoadRoutes(r *gin.Engine) {

	r.GET("/get-timeslot", GetTimeSlotLimits)
	r.GET("/get-road", GetRoadsByCityID)
	r.GET("/get-special", GetSpecificDateLimits)
	r.POST("/add-timeslot",CreateTimeSlotLimit)
	r.POST("/add-special",CreateSpecificDateLimit)
	r.PUT("/update-timeslot",UpdateTimeSlotLimit)
	r.PUT("/add-order",UpdateSpecificDateLimit)
	r.POST("/auto-add", TriggerAutoCreateLimits)
	r.POST("/start-scheduler", StartSchedulerHandler)
	r.POST("/stop-scheduler", StopSchedulerHandler)
	r.GET("/scheduler-status", GetSchedulerStatusHandler)
	r.GET("/get-member", GetUserByID)
	r.GET("/order", GetOrderByCriteria)
	r.GET("/get-image/:id", GetImage)     // 取得圖片
	r.POST("/upload-image", UploadImage)
    r.PUT("/replace-image/:image_id", ReplaceImage)
	r.GET("/all-images", GetAllImages)



}
