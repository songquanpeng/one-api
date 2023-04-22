package controller

import (
	"fmt"
	"gin-template/common"
	"gin-template/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func GetAllFiles(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	files, err := model.GetAllFiles(p*common.ItemsPerPage, common.ItemsPerPage)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    files,
	})
	return
}

func SearchFiles(c *gin.Context) {
	keyword := c.Query("keyword")
	files, err := model.SearchFiles(keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    files,
	})
	return
}

func UploadFile(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	uploadPath := common.UploadPath
	description := c.PostForm("description")
	if description == "" {
		description = "无描述信息"
	}
	uploader := c.GetString("username")
	if uploader == "" {
		uploader = "访客用户"
	}
	uploaderId := c.GetInt("id")
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	files := form.File["file"]
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		ext := filepath.Ext(filename)
		link := common.GetUUID() + ext
		savePath := filepath.Join(uploadPath, link) // both parts are checked, so this path should be safe to use
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		// save to database
		fileObj := &model.File{
			Description: description,
			Uploader:    uploader,
			UploadTime:  currentTime,
			UploaderId:  uploaderId,
			Link:        link,
			Filename:    filename,
		}
		err = fileObj.Insert()
		if err != nil {
			_ = fmt.Errorf(err.Error())
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func DeleteFile(c *gin.Context) {
	fileIdStr := c.Param("id")
	fileId, err := strconv.Atoi(fileIdStr)
	if err != nil || fileId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	fileObj := &model.File{
		Id: fileId,
	}
	model.DB.Where("id = ?", fileId).First(&fileObj)
	if fileObj.Link == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "文件不存在！",
		})
		return
	}
	err = fileObj.Delete()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": err.Error(),
		})
		return
	} else {
		message := "文件删除成功"
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": message,
		})
	}

}

func DownloadFile(c *gin.Context) {
	path := c.Param("file")
	fullPath := filepath.Join(common.UploadPath, path)
	if !strings.HasPrefix(fullPath, common.UploadPath) {
		// We may being attacked!
		c.Status(403)
		return
	}
	c.File(fullPath)
	// Update download counter
	go func() {
		model.UpdateDownloadCounter(path)
	}()
}
