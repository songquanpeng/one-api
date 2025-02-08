package message

import (
	"fmt"

	"github.com/songquanpeng/one-api/common/config"
)

// EmailTemplate 生成美观的 HTML 邮件内容
func EmailTemplate(title, content string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 20px; font-family: Arial, sans-serif; line-height: 1.6; background-color: #f4f4f4;">
    <div style="max-width: 600px; margin: 20px auto; padding: 30px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);">
        <div style="text-align: center; margin-bottom: 30px;">
            <h2 style="color: #333; margin: 0; font-size: 24px;">%s</h2>
        </div>
        <div style="color: #555; font-size: 16px;">
            %s
        </div>
        <div style="margin-top: 40px; padding-top: 20px; border-top: 1px solid #eee; color: #888; font-size: 14px; text-align: center;">
            <p style="margin: 5px 0;">此邮件由系统自动发送，请勿直接回复</p>
            <p style="margin: 5px 0;">%s</p>
        </div>
    </div>
</body>
</html>
`, title, content, config.SystemName)
}
