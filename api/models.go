// models.go
package api

import "time"

// TimeSlotLimit  表示時段限制的結構
type TimeSlotLimit struct {
    TimeSlot   string `json:"time_slot"`
    LimitCount int    `json:"limit_count"`
}
// TimeSlotLimits 是多個 TimeSlotLimit 的切片
type TimeSlotLimits []TimeSlotLimit

// SpecificDateLimit  表示特定日期的時段限制結構
type SpecificDateLimit struct {
    Date       string                  `json:"date"`
    TimeLimits map[string]int          `json:"time_limits"`
    
}



// Road 表示路名和城市 ID 的結構
type Road struct {
    Name    string `json:"name"`
    CityID  int    `json:"city_id"`
}

// Roads 是多個 Road 的切片
type Roads []Road

// Order 表示訂單的結構
type Order struct {
    OrderID           int    `json:"orderID"`
	ID                int    `json:"id"`
    Code              string `json:"code"`
    PersonalName      string `json:"personal_name"`
    DeliveryDate      string `json:"delivery_date"` // 新增的字段，只包含日期
    ShippingStateID   int    `json:"shipping_state_id"`
    ShippingCityID    int    `json:"shipping_city_id"`
    ShippingRoad      string `json:"shipping_road"`
    ShippingAddress1  string `json:"shipping_address1"`
    // StatusID          int    `json:"status_id"`
    StatusCode        string `json:"status_code"`
    DeliveryTimeRange string `json:"delivery_time_range"`
    Mobile            string `json:"mobile"`
    ShippingStatus   int  `json:"shipping_status"`

}
// 外送員
type orderDelivery struct{
    ID                int    `json:"id"`
    OrderCode       int    `json:"order_code"`
    DeliveryID       int    `json:"delivery_id"`
    Name       string    `json:"name"`
    Phone       string    `json:"phone"`
    CarType       string    `json:"cartype"`
    Fee       int    `json:"fee"`
}


// Image 表圖片的結構
type Image struct {
    ID            int       `json:"id"`
    S3URL string    `json:"cloudinary_url"`
    PublicID      string    `json:"public_id"` 
    Title         string    `json:"title"`
    Description   string    `json:"description"`
    CreatedAt     time.Time `json:"created_at"`
}

// OrderProduct 表示 order_products 表的結構
type OrderProduct struct {
    ID           int    `json:"id"`
    OrderID      int    `json:"order_id"`
    ProductID    int    `json:"product_id"`
    Name         string `json:"name"`
    Quantity     int    `json:"quantity"`
}

// OrderProductOption 表示 order_product_options 表的結構
type OrderProductOption struct {
    ID              int    `json:"id"`
    OrderProductID  int    `json:"order_product_id"`
    ProductID       int    `json:"product_id"`
    Name            string `json:"name"`
    Value           string `json:"value"`
    Quantity        float64    `json:"quantity"`
}
//餐點呈現
type CompleteMeal struct {
    MainMeal  OrderProduct       `json:"main_meal"`
    SideMeals []OrderProductOption `json:"side_meals"`
}
//新的餐點
type NewOrderRequest struct {
	Code                   int  `json:"code"`
	CustomerID       int   `json:"customer_id"`
	LocationID         int  `json:"location_id"`
    PersonalName     string `json:"personal_name"`
    DeliveryDate     string `json:"delivery_date"`
    ShippingStateID  int    `json:"shipping_state_id"`
    ShippingCityID   int    `json:"shipping_city_id"`
    ShippingRoad     string `json:"shipping_road"`
    ShippingAddress1 string `json:"shipping_address1"`
    StatusCode       string `json:"status_code"`
    DeliveryTimeRange string `json:"delivery_time_range"`
    OrderMeals       []OrderMeal `json:"order_meals"`
}

type OrderMeal struct {
    MainMeal  OrderProduct        `json:"main_meal"`
    SideMeals []OrderProductOption `json:"side_meals"`
}

// OrderData 用於接收從前端傳來的訂單數據
type OrderData struct {
    Amount    int      `json:"amount"`    // 訂單金額
    TradeDesc string   `json:"tradeDesc"` // 交易描述
    ItemNames []string `json:"itemNames"` // 商品名稱列表
    // 根據需要添加其他字段，如客戶信息等
}

//測試
type Test2 struct {
    ID       int    `json:"id"`
    Type     string `json:"type"`
    Name     string `json:"name"`
    MainMeal string `json:"mainMeal"`
    Slide     string `json:"slide"`
    Slide2    string `json:"slide2"`
    Slide3    string `json:"slide3"`
    Slide4    string `json:"slide4"`
    Slide5    string `json:"slide5"`
    Drink    string `json:"drink"`
    MainMeal2 string `json:"mainMeal2"`
}








