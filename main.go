//main.go
package main

import (
	"onlineBingGin/api"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/joho/godotenv"

	"log"
	"os"
	// "net/http"
)
// func IPAuthMiddleware(allowedIPs []string) gin.HandlerFunc {
//     return func(c *gin.Context) {
//         clientIP := c.ClientIP()

//         // 檢查ip是否在列表
//         for _, allowedIP := range allowedIPs {
//             if clientIP == allowedIP {
//                 c.Next()
//                 return
//             }
//         }

//         // 如果IP不在列表拒絕訪問
//         c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "誰允許你傳送了"})
//     }
// }
func main() {
	// 讀取 .env 檔案
	err := godotenv.Load()
	if err != nil {
		log.Fatal("無法讀取 .env 檔案")
	}



	// 根據設定的環境變數值設定 Gin 模式
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = "debug" // 預設為 debug 模式
	}
	gin.SetMode(mode)


	fmt.Println("伺服器開始運行")

	// 初始化數據庫連接
	api.InitDB()

	// 創建 Gin 實例
	r := gin.Default()

// 	// 設置允許的ip
// allowedIPs := []string{"10.0.0.7", "10.0.0.42"}

// // 全局使用中間驗證
// r.Use(IPAuthMiddleware(allowedIPs))

	r.SetTrustedProxies([]string{"10.0.0.7","10.0.0.42"}) //  Nginx 伺服器的 IP 地址


	// 啟用 CORS 中間件
	r.Use(cors.Default())





// config := cors.Config{
//     AllowOrigins:     []string{"https://example.com"},
//     AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
//     AllowHeaders:     []string{"Origin", "Content-Type"},
//     AllowCredentials: true,
//     MaxAge:           12 * time.Hour,
// }
// r.Use(cors.New(config))



	// 加載 API 路由
	api.LoadRoutes(r)

	// 啟動服務
	r.Run(":8080")
}
