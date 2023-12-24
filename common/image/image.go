package image

import (
	"bytes"
	"encoding/base64"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"regexp"
	"strings"
	"sync"

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

var readerPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Reader{}
	},
}

func GetImageSizeFromBase64(encoded string) (width int, height int, err error) {
	decoded, err := base64.StdEncoding.DecodeString(reg.ReplaceAllString(encoded, ""))
	if err != nil {
		return 0, 0, err
	}

	reader := readerPool.Get().(*bytes.Reader)
	defer readerPool.Put(reader)
	reader.Reset(decoded)

	img, _, err := image.DecodeConfig(reader)
	if err != nil {
		return 0, 0, err
	}

	return img.Width, img.Height, nil
}

func GetImageSize(image string) (width int, height int, err error) {
	if strings.HasPrefix(image, "data:image/") {
		return GetImageSizeFromBase64(image)
	}
	return GetImageSizeFromUrl(image)
}
