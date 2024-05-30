// order.go
package api

import (
	"bytes"
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

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
    log.Printf("Criteria: %+v", criteria)  // 添加日誌輸出
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
// UpdateOrderStatusHandler 更改訂單狀態
func UpdateOrderStatusHandler(c *gin.Context) {
    orderCode := c.Param("code")
    
    // 訂單是否存在
    var orderId int
    err := db.QueryRow("SELECT id FROM orders WHERE code = ?", orderCode).Scan(&orderId)
    if err != nil {
        if err == sql.ErrNoRows {
            // 不存在
            c.JSON(http.StatusNotFound, gin.H{"error": "沒有該筆訂單"})
        } else {
            // 資料庫錯誤
            log.Printf("Query error: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢訂單時發生錯誤"})
        }
        return
    }

    // 存在订单，更新其状态
    err = UpdateOrderStatus(orderCode)
    if err != nil {
        log.Printf("Update error: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "無法更新訂單狀態"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "訂單狀態已更新為作廢"})
}


// GetOrderProducts 根據訂單 ID 獲取訂單餐點
func GetOrderProducts(c *gin.Context) {
    orderID, err := strconv.Atoi(c.Param("order_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "無效的訂單 ID"})
        return
    }

    products, err := FetchOrderProducts(orderID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "無法獲取訂單餐點"})
        return
    }

    c.JSON(http.StatusOK, products)
}

// GetOrderProductOptions 根據 OrderProduct ID 獲取附餐選項
func GetOrderProductOptions(c *gin.Context) {
    orderProductID, err := strconv.Atoi(c.Param("order_product_id"))
    if err != nil {
        log.Printf("Error converting orderProductID to int: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 OrderProduct ID"})
        return
    }

    options, err := FetchOrderProductOptions(orderProductID)
    if err != nil {
        log.Printf("Error fetching order product options: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "無法獲取附餐選項"})
        return
    }

    c.JSON(http.StatusOK, options)
}

// GetCompleteOrderMeal 根據訂單 ID 獲取完整的訂單餐點
func GetCompleteOrderMeal(c *gin.Context) {
    orderID, err := strconv.Atoi(c.Param("order_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "無效的訂單 ID"})
        return
    }

    // 獲取主餐
    mainMeals, err := FetchOrderProducts(orderID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "無法獲取主餐"})
        return
    }

    var completeMeals []CompleteMeal
    for _, mainMeal := range mainMeals {
        // 為每個主餐獲取附餐
        sideMeals, err := FetchOrderProductOptions(mainMeal.ID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "無法獲取附餐"})
            return
        }

        completeMeals = append(completeMeals, CompleteMeal{
            MainMeal:  mainMeal,
            SideMeals: sideMeals,
        })
    }

    c.JSON(http.StatusOK, completeMeals)
}

// CreateNewOrder 處理創建新訂單的請求
func CreateNewOrder(c *gin.Context) {
    var newOrderReq NewOrderRequest
    if err := c.BindJSON(&newOrderReq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 插入訂單基本資料並獲取 order_id
    res, err := db.Exec("INSERT INTO orders (code,location_id,personal_name, delivery_date, customer_id,shipping_state_id, shipping_city_id, shipping_road, shipping_address1, status_code, delivery_time_range) VALUES (?,?, ?, ?, ?, ?, ?, ?, ?,?,?)",
    newOrderReq.Code,newOrderReq.LocationID,newOrderReq.PersonalName, newOrderReq.DeliveryDate, newOrderReq.CustomerID,newOrderReq.ShippingStateID, newOrderReq.ShippingCityID, newOrderReq.ShippingRoad, newOrderReq.ShippingAddress1, newOrderReq.StatusCode, newOrderReq.DeliveryTimeRange)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    orderID, err := res.LastInsertId()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 使用獲得的 order_id 插入主餐和附餐資料
    for _, meal := range newOrderReq.OrderMeals {
        // 插入主餐並獲得主餐ID
        res, err := db.Exec("INSERT INTO order_products (order_id, product_id, name, quantity) VALUES (?, ?, ?, ?)",
            orderID, meal.MainMeal.ProductID, meal.MainMeal.Name, meal.MainMeal.Quantity)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        mainMealID, err := res.LastInsertId()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // 插入對應的附餐
        for _, sideMeal := range meal.SideMeals {
            _, err := db.Exec("INSERT INTO order_product_options (order_product_id, product_id, name, value, quantity) VALUES (?, ?, ?, ?, ?)",
                mainMealID, sideMeal.ProductID, sideMeal.Name, sideMeal.Value, sideMeal.Quantity)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "新訂單創建成功", "order_id": orderID})
}



// UpdateOrderMeal 處理更新訂單餐點的請求
func UpdateOrderMeal(c *gin.Context) {
    var updateMealReq NewOrderRequest
    if err := c.BindJSON(&updateMealReq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    orderID, err := strconv.Atoi(c.Param("order_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "無效的訂單 ID"})
        return
    }

    // 更新主餐
    for _, meal := range updateMealReq.OrderMeals {
        _, err = db.Exec("UPDATE order_products SET product_id = ?, name = ?, quantity = ? WHERE id = ? AND order_id = ?",
            meal.MainMeal.ProductID, meal.MainMeal.Name, meal.MainMeal.Quantity, meal.MainMeal.ID, orderID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // 更新每個主餐對應的副餐
        for _, sideMeal := range meal.SideMeals {
            if sideMeal.ID > 0 {
                // 更新現有副餐
                _, err = db.Exec("UPDATE order_product_options SET product_id = ?, name = ?, value = ?, quantity = ? WHERE id = ? AND order_product_id = ?",
                    sideMeal.ProductID, sideMeal.Name, sideMeal.Value, sideMeal.Quantity, sideMeal.ID, meal.MainMeal.ID)
                if err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                    return
                }
            } else {
                // 新增副餐
                _, err = db.Exec("INSERT INTO order_product_options (order_product_id, product_id, name, value, quantity) VALUES (?, ?, ?, ?, ?)",
                    meal.MainMeal.ID, sideMeal.ProductID, sideMeal.Name, sideMeal.Value, sideMeal.Quantity)
                if err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                    return
                }
            }
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "訂單餐點更新成功"})
}


// ForwardOrderToHTTPService 轉發
func ForwardOrderToHTTPService(c *gin.Context) {
    // os導入環境變數
    baseURL := os.Getenv("POS_API_URL")
    if baseURL == "" {
        log.Println("Environment variable BASE_API_URL not set")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "伺服器錯誤"})
        return
    }

    // 路徑
    specificPath := "/sale/order/save"
    fullURL := baseURL + specificPath
    // 要
    body, err := io.ReadAll(c.Request.Body)
    if err != nil {
        log.Printf("Error reading body: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "無法讀取"})
        return
    }

    // 轉發
    req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(body))
    if err != nil {
        log.Printf("Error creating request: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "請求失敗"})
        return
    }


    // 複製
    req.Header.Set("Content-Type", c.GetHeader("Content-Type"))

    // 轉發請其
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error sending request: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "轉發失敗"})
        return
    }
    defer resp.Body.Close()

    // 回應
    response, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Error reading response body: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "回應失敗"})
        return
    }

    // 回傳
    c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), response)
}







