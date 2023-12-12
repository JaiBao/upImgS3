//image.go
package api

import (

	"log"
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/service/s3"
    "strings"
    "os"
)





// UploadImage 上傳圖片到 Amazon S3
func UploadImage(c *gin.Context) {
    fileHeader, err := c.FormFile("image")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "無法獲取文件"})
        return
    }

    file, err := fileHeader.Open()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "無法打開文件"})
        return
    }
    defer file.Close()

    // 獲取標題和描述
    title := c.PostForm("title")
    description := c.PostForm("description")

	// 確保env導入
    awsRegion := os.Getenv("AWS_REGION")
    if awsRegion == "" {
        log.Println("AWS 區位未設置")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "伺服器設置錯誤"})
        return
    }

	// log.Println("AWS Region:", os.Getenv("AWS_REGION"))


    // 設置 AWS S3 的配置
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(os.Getenv("AWS_REGION")),
		Endpoint: aws.String("https://s3-ap-northeast-1.amazonaws.com"), // 顯示S3URL
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			""),
	})
	if err != nil {
		log.Printf("創建AWS失敗: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "創建AWS失敗"})
		return
	}
	

    svc := s3.New(sess)

  
    // 上傳S3 
    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(os.Getenv("AWS_BUCKET_NAME")),
        Key:    aws.String(fileHeader.Filename),
        Body:   file,
  
    })
	if err != nil {
        log.Printf("上傳S3失敗: %v", err) // 记录详细错误信息
        c.JSON(http.StatusInternalServerError, gin.H{"error": "上傳S3失敗"})
        return
    }

    // 創建S3URL
    s3URL := "https://" + os.Getenv("AWS_BUCKET_NAME") + ".s3.amazonaws.com/" + fileHeader.Filename

    // 存到資料庫
    id, err := InsertImage(s3URL, title, description)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "資料庫保存失敗"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "圖片上傳成功", "id": id, "url": s3URL})
}

// GetAllImages 全部
func GetAllImages(c *gin.Context) {
    images, err := FetchAllImages()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "獲取失敗"})
        return
    }

    c.JSON(http.StatusOK, images)
}

// GetImage 獲取圖片資訊
func GetImage(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "無效的圖片ID"})
        return
    }

    img, err := FetchImage(id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "無法獲取圖片資訊"})
        return
    }

    c.JSON(http.StatusOK, img)
}


// ReplaceImage 替換 Amazon S3 中的圖片並更新數據庫記錄
func ReplaceImage(c *gin.Context) {
    //取得圖片ID
    idStr := c.Param("image_id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "無效的圖片 ID"})
        return
    }
    //取得原有圖片
    existingImage, err := FetchImage(id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "無法取得原有圖片"})
        return
    }

    // 獲取標題和描述
    title := c.PostForm("title")
    description := c.PostForm("description")

    var newS3URL string
    fileHeader, _ := c.FormFile("image")
    if fileHeader != nil {
        // 有新圖片，處理圖片上傳
        file, err := fileHeader.Open()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "無法打開文件"})
            return
        }
        defer file.Close()

        // 構建 AWS S3 設置
        sess, _ := session.NewSession(&aws.Config{
            Region:      aws.String(os.Getenv("AWS_REGION")),
            Credentials: credentials.NewStaticCredentials(
                os.Getenv("AWS_ACCESS_KEY_ID"),
                os.Getenv("AWS_SECRET_ACCESS_KEY"),
                ""),
        })
        svc := s3.New(sess)

        // 上傳新圖片到 S3
        _, err = svc.PutObject(&s3.PutObjectInput{
            Bucket: aws.String(os.Getenv("AWS_BUCKET_NAME")),
            Key:    aws.String(fileHeader.Filename),
            Body:   file,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "無法上傳新圖片"})
            return
        }

        // 創建新的 S3 URL
        newS3URL = "https://" + os.Getenv("AWS_BUCKET_NAME") + ".s3.amazonaws.com/" + fileHeader.Filename
        //刪除舊圖片
        oldS3URL := existingImage.S3URL
        if oldS3URL != "" {
            key := strings.TrimPrefix(oldS3URL, "https://" + os.Getenv("AWS_BUCKET_NAME") + ".s3.amazonaws.com/")

            // 删除操作
            _, delErr := svc.DeleteObject(&s3.DeleteObjectInput{
                Bucket: aws.String(os.Getenv("AWS_BUCKET_NAME")),
                Key:    aws.String(key),
            })
            if delErr != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除舊圖片失敗"})
                return
            }
        }
    } else {
        // 沒有新圖片，獲取原有的 S3 URL
        existingImage, err := FetchImage(id)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "無法獲取原有圖片信息"})
            return
        }
        newS3URL = existingImage.S3URL
    }

    // 更新數據庫記錄
    err = UpdateImage(id, newS3URL, title, description)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "更新圖片失敗"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "圖片資訊更新成功", "new_url": newS3URL})
}




