package image

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"regexp"
	"strings"

	_ "golang.org/x/image/webp"
)

func GetImageSizeFromUrl(url string) (width int, height int, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	img, _, err := image.DecodeConfig(resp.Body)
	if err != nil {
		return
	}
	return img.Width, img.Height, nil
}

var (
	reg = regexp.MustCompile(`data:image/([^;]+);base64,`)
)

func GetImageSizeFromBase64(encoded string) (width int, height int, err error) {
	encoded = strings.TrimPrefix(encoded, "data:image/png;base64,")
	base64 := strings.NewReader(reg.ReplaceAllString(encoded, ""))
	img, _, err := image.DecodeConfig(base64)
	if err != nil {
		return
	}
	return img.Width, img.Height, nil
}

func GetImageSize(image string) (width int, height int, err error) {
	if strings.HasPrefix(image, "data:image/") {
		return GetImageSizeFromBase64(image)
	}
	return GetImageSizeFromUrl(image)
}
