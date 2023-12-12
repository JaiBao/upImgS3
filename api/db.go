//db.go
package api

import (
    "fmt"
    "os"
    "github.com/joho/godotenv"
    "time"
    "database/sql"
    "log"
    _ "github.com/go-sql-driver/mysql"
)

var db *sql.DB



func InitDB() {
    // 載入環境變量
    err := godotenv.Load()
    if err != nil {
        log.Fatal("環境變量讀取失敗")
    }

   
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASS"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_NAME"),
    )

    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("已連結至資料庫")
}


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

    // 修改 SQL 查詢來匹配名稱和手機號碼
    err := db.QueryRow("SELECT id FROM users WHERE mobile = ? AND name = ?", mobile, name).Scan(&userID)
    if err != nil {
        if err == sql.ErrNoRows {
            // 如果沒有找到，返回空字符串
            return "", nil
        }
        return "", err
    }

    return userID, nil
}



// FetchOrderByCriteria 根據指定條件查詢訂單
func FetchOrderByCriteria(criteria map[string]string) ([]Order, error) {
    query := "SELECT code, personal_name, DATE_FORMAT(delivery_date, '%Y-%m-%d') as delivery_date, shipping_state_id, shipping_city_id, shipping_road, shipping_address1, status_code, delivery_time_range, mobile FROM orders WHERE "
    var args []interface{}

    for key, value := range criteria {
        query += fmt.Sprintf("%s = ? AND ", key)
        args = append(args, value)
    }

    // 移除最後的 "AND "
    query = query[:len(query)-5]

    rows, err := db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orders []Order // Order 是你的訂單結構

    for rows.Next() {
        var order Order
        // 可以根據Order結構調整scan
        if err := rows.Scan(&order.Code, &order.PersonalName, &order.DeliveryDate, &order.ShippingStateID, &order.ShippingCityID, &order.ShippingRoad, &order.ShippingAddress1, &order.StatusCode, &order.DeliveryTimeRange, &order.Mobile); err != nil {
            return nil, err
        }
        orders = append(orders, order)
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


