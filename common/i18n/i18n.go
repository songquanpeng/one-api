package i18n

import (
	"embed"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed locales/*.json
var localesFS embed.FS

var (
	translations = make(map[string]map[string]string)
	defaultLang  = "en"
	ContextKey   = "i18n"
)

// Init loads all translation files from embedded filesystem
func Init() error {
	entries, err := localesFS.ReadDir("locales")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		langCode := strings.TrimSuffix(entry.Name(), ".json")
		content, err := localesFS.ReadFile("locales/" + entry.Name())
		if err != nil {
			return err
		}

		var translation map[string]string
		if err := json.Unmarshal(content, &translation); err != nil {
			return err
		}
		translations[langCode] = translation
	}

	return nil
}

func GetLang(c *gin.Context) string {
	rawLang, ok := c.Get(ContextKey)
	if !ok {
		return defaultLang
	}
	lang, _ := rawLang.(string)
	if lang != "" {
		return lang
	}
	return defaultLang
}

func Translate(c *gin.Context, message string) string {
	lang := GetLang(c)
	return translateHelper(lang, message)
}

func translateHelper(lang, message string) string {
	if trans, ok := translations[lang]; ok {
		if translated, exists := trans[message]; exists {
			return translated
		}
	}
	return message
}
