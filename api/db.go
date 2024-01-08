//db.go
package api

import (
    "fmt"
    "os"
    "github.com/joho/godotenv"
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



