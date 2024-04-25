package vertex

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/songquanpeng/one-api/relay/meta"
	"io"
	"net/http"
	"time"
)

type Credentials struct {
	PrivateKey   string
	PrivateKeyID string
	ClientEmail  string
}

// ServiceAccount holds the credentials and scopes required for token generation
type ServiceAccount struct {
	Cred   *Credentials
	Scopes string
}

var scopes = "https://www.googleapis.com/auth/cloud-platform"

// createSignedJWT creates a Signed JWT from service account credentials
func (sa *ServiceAccount) createSignedJWT() (string, error) {
	if sa.Cred == nil {
		return "", fmt.Errorf("credentials are nil")
	}

	issuedAt := time.Now()
	expiresAt := issuedAt.Add(time.Hour)

	claims := &jwt.MapClaims{
		"iss":   sa.Cred.ClientEmail,
		"sub":   sa.Cred.ClientEmail,
		"aud":   "https://www.googleapis.com/oauth2/v4/token",
		"iat":   issuedAt.Unix(),
		"exp":   expiresAt.Unix(),
		"scope": scopes,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = sa.Cred.PrivateKeyID
	token.Header["alg"] = "RS256"
	token.Header["typ"] = "JWT"

	// 解析 PEM 编码的私钥
	block, _ := pem.Decode([]byte(sa.Cred.PrivateKey))
	if block == nil {
		return "", errors.New("failed to decode PEM block containing private key")
	}

	// 解析 RSA 私钥
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("private key is not of type RSA")
	}

	signedToken, err := token.SignedString(rsaPrivateKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// getToken uses the signed JWT to obtain an access token
func (sa *ServiceAccount) getToken(ctx context.Context) (string, error) {
	signedJWT, err := sa.createSignedJWT()
	if err != nil {
		return "", err
	}

	return exchangeJwtForAccessToken(ctx, signedJWT)
}

// exchangeJwtForAccessToken exchanges a Signed JWT for a Google OAuth Access Token.
func exchangeJwtForAccessToken(ctx context.Context, signedJWT string) (string, error) {
	authURL := "https://www.googleapis.com/oauth2/v4/token"
	params := map[string]string{
		"grant_type": "urn:ietf:params:oauth:grant-type:jwt-bearer",
		"assertion":  signedJWT,
	}

	jsonData, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	// Extract the access token from the response
	accessToken, ok := data["access_token"].(string)
	if !ok {
		return "", err // You might want to return a more specific error here
	}

	return accessToken, nil
}

func getToken(ctx context.Context, meta *meta.Meta) (string, error) {
	// todo 每次请求都要换次token？？？
	encodedString := ""
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		return "", err
	}
	m := make(map[string]string)
	err = json.Unmarshal(decodedBytes, &m)
	if err != nil {
		return "", err
	}

	sa := &ServiceAccount{
		Cred: &Credentials{
			PrivateKey:   m["private_key"],
			PrivateKeyID: m["private_key_id"],
			ClientEmail:  m["client_email"],
		},
		Scopes: scopes,
	}
	return sa.getToken(ctx)
}
