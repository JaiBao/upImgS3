// queries.go
package api

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// FetchRoadsByCityID 從數據庫中獲取特定城市 ID 的所有路名
func FetchRoadsByCityID(cityID int) (Roads, error) {
    var roads Roads

    // 查詢來選擇特定城市 ID 的路名
    rows, err := db.Query("SELECT name, city_id FROM roads WHERE city_id = ?", cityID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var road Road
        if err := rows.Scan(&road.Name, &road.CityID); err != nil {
            return nil, err
        }
        roads = append(roads, road)
    }

    return roads, nil
}


// FetchTimeSlotLimits 從數據庫中獲取所有時段限制
func FetchTimeSlotLimits() ([]TimeSlotLimit, error) {
    var limits []TimeSlotLimit
    rows, err := db.Query("SELECT TimeSlot, LimitCount FROM TimeSlotLimits")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var limit TimeSlotLimit
        if err := rows.Scan(&limit.TimeSlot, &limit.LimitCount); err != nil {
            return nil, err
        }
        limits = append(limits, limit)
    }

    return limits, nil
}

// FetchSpecificDateLimits 從數據庫中獲取特定日期的時段限制
func FetchSpecificDateLimits() (map[string]map[string]map[string]int, error) {
    today := time.Now().Format("2006-01-02")

    rows, err := db.Query("SELECT Date, TimeSlot, LimitCount FROM DateLimits WHERE Date >= ? ORDER BY Date, TimeSlot", today)
  
    // rows, err := db.Query("SELECT Date, TimeSlot, LimitCount FROM DateLimits ORDER BY Date, TimeSlot")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    monthlyLimits := make(map[string]map[string]map[string]int)
    var date, timeSlot string
    var limitCount int

    for rows.Next() {
        err := rows.Scan(&date, &timeSlot, &limitCount)
        if err != nil {
            return nil, err
        }

        // 分解日期
        yearMonth := date[:7] // 前7個字

        if _, exists := monthlyLimits[yearMonth]; !exists {
            monthlyLimits[yearMonth] = make(map[string]map[string]int)
        }

        if _, dayExists := monthlyLimits[yearMonth][date]; !dayExists {
            monthlyLimits[yearMonth][date] = make(map[string]int)
        }

        monthlyLimits[yearMonth][date][timeSlot] = limitCount
    }

    return monthlyLimits, nil
}

func InsertTimeSlotLimits(limits TimeSlotLimits) error {
    for _, limit := range limits {
        stmt, err := db.Prepare("INSERT INTO TimeSlotLimits (TimeSlot, LimitCount) VALUES (?, ?)")
        if err != nil {
            return err 
        }

        _, err = stmt.Exec(limit.TimeSlot, limit.LimitCount)
        stmt.Close() 

        if err != nil {
            return err 
        }
    }
    return nil
}


func UpdateExistingTimeSlotLimits(limits map[string]int) error {
    for timeSlot, limitCount := range limits {
        stmt, err := db.Prepare("UPDATE TimeSlotLimits SET LimitCount = ? WHERE TimeSlot = ?")
        if err != nil {
            return err
        }
        defer stmt.Close()

        _, err = stmt.Exec(limitCount, timeSlot)
        if err != nil {
            return err
        }
    }
    return nil
}

//增加設定日期
func insertDateIfNeeded(date string) error {
    // 檢查日期是否已存在
    var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM Dates WHERE Date = ?)", date).Scan(&exists)
    if err != nil {
        return err
    }

    // 如果日期不存在，則插入
    if !exists {
        stmt, err := db.Prepare("INSERT INTO Dates (Date) VALUES (?)")
        if err != nil {
            return err
        }
        defer stmt.Close()

        _, err = stmt.Exec(date)
        if err != nil {
            return err
        }
    }

    return nil
}
func InsertSpecificDateLimit(dateLimit SpecificDateLimit) error {
    // 首先檢查並可能插入日期到Dates表
    if err := insertDateIfNeeded(dateLimit.Date); err != nil {
        return err 
    }

     // 插入或更新 DateLimits 表
     for timeSlot, limit := range dateLimit.TimeLimits {
        stmt, err := db.Prepare("INSERT INTO DateLimits (Date, TimeSlot, LimitCount) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE LimitCount = VALUES(LimitCount)")
        if err != nil {
            return err 
        }

        _, err = stmt.Exec(dateLimit.Date, timeSlot, limit)
        stmt.Close() // 立即關閉語句

        if err != nil {
            return err 
        }
    }
    return nil
}



func UpdateExistingSpecificDateLimit(dateLimit SpecificDateLimit) error {
    for timeSlot, limit := range dateLimit.TimeLimits {
        stmt, err := db.Prepare("UPDATE DateLimits SET LimitCount = ? WHERE Date = ? AND TimeSlot = ?")
        if err != nil {
            return err
        }
        defer stmt.Close()

        _, err = stmt.Exec(limit, dateLimit.Date, timeSlot)
        if err != nil {
            return err
        }
    }
    return nil
}

// AutoCreateNextTwoMonthsLimits 自動新增特定時間範圍的限制
func AutoCreateNextTwoMonthsLimits(period string , cover bool) error {
    // 取得現有設定日期
    existingLimits, err := FetchSpecificDateLimits()
    if err != nil {
        return err
    }

    // 取得預設
    initialLimits, err := FetchTimeSlotLimits()
    if err != nil {
        return err
    }

    // 轉格式
    initialTimeLimits := make(map[string]int)
    for _, limit := range initialLimits {
        initialTimeLimits[limit.TimeSlot] = limit.LimitCount
    }

    // 時間範圍
    startDate := time.Now()
    var endDate time.Time
    switch period {
    case "oneWeek":
        endDate = startDate.AddDate(0, 0, 7)
    case "twoWeeks":
        endDate = startDate.AddDate(0, 0, 14)
    case "oneMonth":
        endDate = startDate.AddDate(0, 1, 0)
    case "twoMonths":
        endDate = startDate.AddDate(0, 2, 0)
    default:
        return fmt.Errorf("不支援的時間範圍: %s", period)
    }

  // 日期循環
 for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
    dateStr := d.Format("2006-01-02")
    yearMonth := d.Format("2006-01") 

 // 檢查日期是否已存在
if !cover{
    if monthLimits, monthExists := existingLimits[yearMonth]; monthExists {
        if _, dateExists := monthLimits[dateStr]; dateExists {
            continue // 如果该日期已存在，则跳过
        }
    }
}
      // 沒有設定的插入預設
    dateLimit := SpecificDateLimit{
        Date:       dateStr,
        TimeLimits: initialTimeLimits,
    }

    if err := InsertSpecificDateLimit(dateLimit); err != nil {
        return err
    }
}

return nil
}

// FetchUserIDByNameAndMobile 從數據庫中根據手機號碼和姓名獲取使用者 ID
func FetchUserIDByNameAndMobile(name, mobile string) (string, error) {
    var userID string

    var err error
    // 根據提供的參數動態選擇SQL查詢
    if name != "" && mobile != "" {
        // 同時有名字和手機號碼時的查詢
        err = db.QueryRow("SELECT id FROM users WHERE mobile = ? AND name = ?", mobile, name).Scan(&userID)
    } else if mobile != "" {
        // 僅有手機號碼時的查詢
        err = db.QueryRow("SELECT id FROM users WHERE mobile = ?", mobile).Scan(&userID)
    } else {
        // 如果既沒有手機號碼也沒有名字，直接返回錯誤
        return "", errors.New("需要至少提供手機號碼或名字")
    }

    if err != nil {
        if err == sql.ErrNoRows {
            // 如果沒有找到，返回空字符串
            return "", nil
        }
        return "", err
    }

    return userID, nil
}


//FetchDelivery 根據訂單編號查詢外送員
func FetchDelivery(orderCode int) (*orderDelivery, error) {
    var delivery orderDelivery

    err := db.QueryRow("SELECT id, order_code, delivery_id, name, phone, cartype, fee FROM order_delivery WHERE order_code = ?", orderCode).Scan(
        &delivery.ID,
        &delivery.OrderCode,
        &delivery.DeliveryID,
        &delivery.Name,
        &delivery.Phone,
        &delivery.CarType,
        &delivery.Fee,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }

    return &delivery, nil
}



// FetchOrderByCriteria 根據指定條件查詢訂單
func FetchOrderByCriteria(criteria map[string]string) ([]Order, error) {
    query := "SELECT id AS orderID,code, personal_name, DATE_FORMAT(delivery_date, '%Y-%m-%d') as delivery_date, shipping_state_id, shipping_city_id, shipping_road, shipping_address1, status_code, delivery_time_range, mobile,shipping_status FROM orders WHERE "
    var args []interface{}

    for key, value := range criteria {
        query += fmt.Sprintf("%s = ? AND ", key)
        args = append(args, value)
    }

    // 移除最後的 "AND "
    query = query[:len(query)-5]
    log.Printf("SQL Query: %s", query)  // 添加日誌輸出
    rows, err := db.Query(query, args...)
    if err != nil {
        log.Printf("Query error: %v", err)  // 添加錯誤日誌
        return nil, err
    }
    defer rows.Close()

    var orders []Order // Order 是你的訂單結構

    for rows.Next() {
        var order Order
        // 可以根據Order結構調整scan
        if err := rows.Scan(&order.OrderID,&order.Code, &order.PersonalName, &order.DeliveryDate, &order.ShippingStateID, &order.ShippingCityID, &order.ShippingRoad, &order.ShippingAddress1, &order.StatusCode, &order.DeliveryTimeRange, &order.Mobile,&order.ShippingStatus); err != nil {
            log.Printf("Scan error: %v", err)  // 添加錯誤日誌
            return nil, err
        }
        orders = append(orders, order)
        log.Printf("Found order: %+v", order) // 日誌輸出訂單詳情
    }

    return orders, nil
}



// 插入圖片到sql
func InsertImage(s3URL, title, description string) (int, error) {
    stmt, err := db.Prepare("INSERT INTO images (s3_url, title, description) VALUES (?, ?, ?)")
    if err != nil {
        log.Printf("錯誤：sql語法錯誤 - %v", err)
        return 0, err
    }
    defer stmt.Close()

    res, err := stmt.Exec(s3URL, title, description)
    if err != nil {
        log.Printf("錯誤：SQL語法執行錯誤- %v", err)
        return 0, err
    }

    id, err := res.LastInsertId()
    if err != nil {
        log.Printf("錯誤：獲取插入ID錯誤 - %v", err)
    }
    return int(id), err
}


// 更新圖片信息
func UpdateImage(id int, s3URL, title, description string) error {
    stmt, err := db.Prepare("UPDATE images SET s3_url = ?, title = ?, description = ? WHERE id = ?")

    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(s3URL, title, description, id)
    return err
}



// 獲取圖片
func FetchImage(id int) (*Image, error) {
    var img Image
    var createdAtString string // 增加一個字符串變量來臨時存儲日期時間

    err := db.QueryRow("SELECT id, s3_url, title, description, created_at FROM images WHERE id = ?", id).Scan(
        &img.ID, 
        &img.S3URL, 
        &img.Title, 
        &img.Description, 
        &createdAtString, // 掃描created_at到字符串
    )
    if err != nil {
        log.Printf("FetchImage: error fetching image with id %d: %v", id, err)
        return nil, err
    }

    // 解析日期時間字符串為time.Time類型
    img.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtString)
    if err != nil {
        log.Printf("FetchImage: error parsing created_at for image with id %d: %v", id, err)
        return nil, err
    }

    return &img, nil
}

// FetchAllImages 一次拿全部
func FetchAllImages() ([]Image, error) {
    var images []Image
    rows, err := db.Query("SELECT id, s3_url, title, description, created_at FROM images")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var img Image
        var createdAtString string 

        if err := rows.Scan(&img.ID, &img.S3URL, &img.Title, &img.Description, &createdAtString); err != nil {
            return nil, err
        }

       
        img.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtString)
        if err != nil {
            return nil, err
        }

        images = append(images, img)
    }

    return images, nil
}

// FetchOrderProducts 根據訂單 ID 獲取訂單餐點
func FetchOrderProducts(orderID int) ([]OrderProduct, error) {
    var products []OrderProduct
    rows, err := db.Query("SELECT id, order_id, product_id, name, quantity FROM order_products WHERE order_id = ?", orderID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var product OrderProduct
        if err := rows.Scan(&product.ID, &product.OrderID, &product.ProductID, &product.Name, &product.Quantity); err != nil {
            return nil, err
        }
        products = append(products, product)
    }

    return products, nil
}

// FetchOrderProductOptions 根據 OrderProduct ID 獲取附餐選項
func FetchOrderProductOptions(orderProductID int) ([]OrderProductOption, error) {
    var options []OrderProductOption
    rows, err := db.Query("SELECT id, order_product_id, product_id, name, value, quantity FROM order_product_options WHERE order_product_id = ?", orderProductID)
    if err != nil {
        log.Printf("Error FetchOrderProductOptions: %v", err)
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var option OrderProductOption
        if err := rows.Scan(&option.ID, &option.OrderProductID, &option.ProductID, &option.Name, &option.Value, &option.Quantity); err != nil {
            return nil, err
        }
        options = append(options, option)
    }

    return options, nil
}


// UpdateOrderStatus 用訂單編號更改訂單狀態
func UpdateOrderStatus(orderCode string) error {
    stmt, err := db.Prepare("UPDATE orders SET status_code = 'Void' WHERE code = ?")
    if err != nil {
        log.Printf("Prepare update error: %v", err)
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(orderCode)
    if err != nil {
        log.Printf("Execute update error: %v", err)
        return err
    }

    log.Printf("Order status updated to 'Void' for order code: %s", orderCode)
    return nil
}




// FetchTest2ByName 通过名称从 test2 表中检索条目。
func FetchTest2ByName(typeStr, name string) (*Test2, error) {
    var t Test2
    err := db.QueryRow("SELECT type, name, mainMeal, slide, slide2, slide3, slide4, slide5, drink ,mainMeal2 FROM test2 WHERE type = ? AND name = ?", typeStr, name).Scan(&t.Type, &t.Name, &t.MainMeal, &t.Slide, &t.Slide2, &t.Slide3, &t.Slide4, &t.Slide5, &t.Drink , &t.MainMeal2)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        log.Println("FetchTest2ByName error:", err)
        return nil, err
    }
    return &t, nil
}

// UpdateTest2ByName 根据名称更新 test2 表中的条目。
func UpdateTest2ByNameInDB(t *Test2) error {
    result, err := db.Exec("UPDATE test2 SET type = ?, mainMeal = ?, slide = ?, slide2 = ?, slide3 = ?, slide4 = ?, slide5 = ?, drink = ?, mainMeal2 = ? WHERE type = ? AND name = ?", t.Type, t.MainMeal, t.Slide, t.Slide2, t.Slide3, t.Slide4, t.Slide5, t.Drink, t.MainMeal2, t.Type, t.Name)

    if err != nil {
        log.Println("UpdateTest2ByName error:", err)
        return err
    }
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        // 沒有的話插入
        _, err := db.Exec("INSERT INTO test2 (type, name, mainMeal, slide, slide2, slide3, slide4, slide5, drink, mainMeal2) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", t.Type, t.Name, t.MainMeal, t.Slide, t.Slide2, t.Slide3, t.Slide4, t.Slide5, t.Drink, t.MainMeal2)
        if err != nil {
            return err
        }
    }
    return nil
}
