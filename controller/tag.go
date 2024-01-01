package controller

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "one-api/model"
)

func GetTags(c *gin.Context) {
    tags := make([]string, 0)
    var checkMap = make(map[string]int)
    channels, err := model.GetAllChannels(0, 0, true)
    if err == nil{
        for i := range channels {
            if _, ok := checkMap[channels[i].Tag];ok{
                continue
            }
            tags = append(tags, channels[i].Tag)
            checkMap[channels[i].Tag] = 1
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "",
        "data":    tags,
    })
}
