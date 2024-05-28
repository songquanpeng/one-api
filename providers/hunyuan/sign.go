package hunyuan

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/types"
	"strconv"
	"strings"
	"time"
)

func sha256hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

func hmacsha256(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	return string(hashed.Sum(nil))
}

func (p *HunyuanProvider) sign(body any, action, method string) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	service := "hunyuan"
	version := "2023-09-01"
	// region := ""
	host := strings.Replace(p.GetBaseURL(), "https://", "", 1)
	algorithm := "TC3-HMAC-SHA256"
	var timestamp = time.Now().Unix()

	secretId, secretKey, err := p.parseHunyuanConfig(p.Channel.Key)
	if err != nil {
		return nil, common.ErrorWrapper(err, "get_tunyuan_secret_failed", http.StatusInternalServerError)
	}

	// ************* 步骤 1：拼接规范请求串 *************
	contentType := "application/json; charset=utf-8"
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-tc-action:%s\n",
		contentType, host, strings.ToLower(action))
	signedHeaders := "content-type;host;x-tc-action"
	payloadJson, _ := json.Marshal(body)
	payloadStr := string(payloadJson)

	hashedRequestPayload := sha256hex(payloadStr)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		method,
		"/",
		"",
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload)

	// ************* 步骤 2：拼接待签名字符串 *************
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	hashedCanonicalRequest := sha256hex(canonicalRequest)
	string2sign := fmt.Sprintf("%s\n%d\n%s\n%s",
		algorithm,
		timestamp,
		credentialScope,
		hashedCanonicalRequest)

	// ************* 步骤 3：计算签名 *************
	secretDate := hmacsha256(date, "TC3"+secretKey)
	secretService := hmacsha256(service, secretDate)
	secretSigning := hmacsha256("tc3_request", secretService)
	signature := hex.EncodeToString([]byte(hmacsha256(string2sign, secretSigning)))

	// ************* 步骤 4：拼接 Authorization *************
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		secretId,
		credentialScope,
		signedHeaders,
		signature)

	// ************* 步骤 5：构造并发起请求 *************
	headers := map[string]string{
		"Host":           host,
		"X-TC-Action":    action,
		"X-TC-Version":   version,
		"X-TC-Timestamp": strconv.FormatInt(timestamp, 10),
		"Content-Type":   contentType,
		"Authorization":  authorization,
	}
	// if region != "" {
	// 	headers["X-TC-Region"] = region
	// }

	req, err := p.Requester.NewRequest(method, p.GetBaseURL(), p.Requester.WithBody(body), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "get_tunyuan_secret_failed", http.StatusInternalServerError)
	}

	return req, nil
}
