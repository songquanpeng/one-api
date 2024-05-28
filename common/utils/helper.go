package utils

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		log.Println(err)
	}
}

func GetIp() (ip string) {
	ips, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return ip
	}

	for _, a := range ips {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ip = ipNet.IP.String()
				if strings.HasPrefix(ip, "10") {
					return
				}
				if strings.HasPrefix(ip, "172") {
					return
				}
				if strings.HasPrefix(ip, "192.168") {
					return
				}
				ip = ""
			}
		}
	}
	return
}

var sizeKB = 1024
var sizeMB = sizeKB * 1024
var sizeGB = sizeMB * 1024

func Bytes2Size(num int64) string {
	numStr := ""
	unit := "B"
	if num/int64(sizeGB) > 1 {
		numStr = fmt.Sprintf("%.2f", float64(num)/float64(sizeGB))
		unit = "GB"
	} else if num/int64(sizeMB) > 1 {
		numStr = fmt.Sprintf("%d", int(float64(num)/float64(sizeMB)))
		unit = "MB"
	} else if num/int64(sizeKB) > 1 {
		numStr = fmt.Sprintf("%d", int(float64(num)/float64(sizeKB)))
		unit = "KB"
	} else {
		numStr = fmt.Sprintf("%d", num)
	}
	return numStr + " " + unit
}

func Seconds2Time(num int) (time string) {
	if num/31104000 > 0 {
		time += strconv.Itoa(num/31104000) + " 年 "
		num %= 31104000
	}
	if num/2592000 > 0 {
		time += strconv.Itoa(num/2592000) + " 个月 "
		num %= 2592000
	}
	if num/86400 > 0 {
		time += strconv.Itoa(num/86400) + " 天 "
		num %= 86400
	}
	if num/3600 > 0 {
		time += strconv.Itoa(num/3600) + " 小时 "
		num %= 3600
	}
	if num/60 > 0 {
		time += strconv.Itoa(num/60) + " 分钟 "
		num %= 60
	}
	time += strconv.Itoa(num) + " 秒"
	return
}

func Interface2String(inter interface{}) string {
	switch inter := inter.(type) {
	case string:
		return inter
	case int:
		return fmt.Sprintf("%d", inter)
	case float64:
		return fmt.Sprintf("%f", inter)
	}
	return "Not Implemented"
}

func UnescapeHTML(x string) interface{} {
	return template.HTML(x)
}

func IntMax(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func GetUUID() string {
	code := uuid.New().String()
	code = strings.Replace(code, "-", "", -1)
	return code
}

const keyChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateKey() string {
	key := make([]byte, 48)
	for i := 0; i < 16; i++ {
		key[i] = keyChars[rand.Intn(len(keyChars))]
	}
	uuid_ := GetUUID()
	for i := 0; i < 32; i++ {
		c := uuid_[i]
		if i%2 == 0 && c >= 'a' && c <= 'z' {
			c = c - 'a' + 'A'
		}
		key[i+16] = c
	}
	return string(key)
}

func GetRandomString(length int) string {
	key := make([]byte, length)
	for i := 0; i < length; i++ {
		key[i] = keyChars[rand.Intn(len(keyChars))]
	}
	return string(key)
}

func GetTimestamp() int64 {
	return time.Now().Unix()
}

func GetTimeString() string {
	now := time.Now()
	return fmt.Sprintf("%s%d", now.Format("20060102150405"), now.UnixNano()%1e9)
}

func Max(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func GetOrDefault[T any](env string, defaultValue T) T {
	if viper.IsSet(env) {
		value := viper.Get(env)
		if v, ok := value.(T); ok {
			return v
		}
	}
	return defaultValue
}

func MessageWithRequestId(message string, id string) string {
	return fmt.Sprintf("%s (request id: %s)", message, id)
}

func String2Int(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return num
}

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func Contains[T comparable](value T, slice []T) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func Filter[T any](arr []T, f func(T) bool) []T {
	var res []T
	for _, v := range arr {
		if f(v) {
			res = append(res, v)
		}
	}
	return res
}

func GetModelsWithMatch(modelList *[]string, modelName string) string {
	for _, model := range *modelList {
		if strings.HasPrefix(modelName, strings.TrimRight(model, "*")) {
			return model
		}
	}
	return ""
}

func EscapeMarkdownText(text string) string {
	chars := []string{"_", "*", "[", "]", "(", ")", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!", "`"}
	for _, char := range chars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
}

func UnmarshalString[T interface{}](data string) (form T, err error) {
	err = json.Unmarshal([]byte(data), &form)
	return form, err
}

func Marshal[T interface{}](data T) string {
	res, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(res)
}
