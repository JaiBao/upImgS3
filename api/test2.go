package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTest2ByName 处理 GET 请求以通过名称获取 Test2 条目。
func GetTest2ByName(c *gin.Context) {
	typeStr := c.Param("type")
    name := c.Param("name")
    test2, err := FetchTest2ByName(typeStr, name)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "後台錯誤"})
        return
    }
    if test2 == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "找不到"})
        return
    }
    c.JSON(http.StatusOK, test2)
}

// UpdateTest2ByName 处理 PUT 请求以通过名称更新 Test2 条目。

func UpdateTest2ByName(c *gin.Context) {
    var t Test2
    if err := c.BindJSON(&t); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "無效數據"})
        return
    }

    if err := UpdateTest2ByNameInDB(&t); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失敗"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"success": true})
}

