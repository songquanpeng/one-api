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
	encodedString := "ewogICJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsCiAgInByb2plY3RfaWQiOiAibW9uaWNhLWRldi0zOTI2MDkiLAogICJwcml2YXRlX2tleV9pZCI6ICIzODlmMTJmYjkyNjkwMDRhOTgzZGU3MmM3NWMzZmQ0MWQxZjQ1ODI3IiwKICAicHJpdmF0ZV9rZXkiOiAiLS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tXG5NSUlFdmdJQkFEQU5CZ2txaGtpRzl3MEJBUUVGQUFTQ0JLZ3dnZ1NrQWdFQUFvSUJBUURvZkJHaGUwNFVZRUxYXG5jbTdiczdiT0dJWlNsRCtwaGI4czk0dGpva0drN0lPOGhmM2VlQW5jTnpOS2k0SHpCZnk3TS81SVM4ekw4eittXG5GRG9vNjdpdUFvbjlLb1Z6Y0FSdFQzRkQ0cCtCZWRSL2lXanJESy9GTW5kTW5DazNmckphTXdId0pJUkNKRnYwXG5xcUdhYzZxQ1NBMVpLRG8zbmxDYXpSaHBla1ArNHB6Z2ppV0RVOFRuTkRJcDBTakNVMGpzVXhycCtXbFVlSEd4XG5KR1cwd3U5QWxROXlUM2xESk11enBCTU5JU2U5MHVqSGh2NG0rTVFKT0NYOUlXbHJXTGp3dTlHdEZueUc4bVp2XG5yem9Tc2JDQ2F6K0xnTXh2YVlaM2hudFpjOFd6b3ZlUUozVUtxK21ucEFXZGRKTWhRU3VremNsMStQZnJjYWpKXG5Fa0xyYnFGVkFnTUJBQUVDZ2dFQURTR3E1a1hlZG1pc2hkcFZHQ3hKdkZEbXoyWEh4Y1hEODJDRkcwcllMZDVkXG5ING9xK1lTcXUrbFRTSmZpTG0ydFpZNk5nNXhpZERlb1pmTlpDS1F0NFloTHJvVFhGbHJpcVNEMld4a0RITzlhXG5lUnkwRkNpNmllY3NsV054c1l5Q3V2VU1IRG9YelZ1YjVSRTVRUTNjK1BCa2JwK0cwRXJ0THgwOERvTWxNWkdVXG5MdXVTS0FKRFJJVlk2dlo5RGpXcWVRRWt5VlhnbDBzN1Zlck03Q041RTk3Wm1qZExDL0NseUJlSXM1MXdkU3c3XG5VMHdmbkdaT2cxblBtL21hWmp0cUxwWUdCSW9aWStaQ3JlbWRRaGxJUWJXZTBENE1ycnR0QW1aMjJsZlhMcmJ2XG4rTFRhdGE0N0lseGh2V3V3YWRXM0FYdmkvdGpmcFBiV25PT1hhbGVqOFFLQmdRRDFta1JDVTNpd3d3cEluRi8xXG5KSnBLdXNva2w5QTdWOHBzM1R3TmhsVVpPYXRZMFhZODZhNk5Dcy9QSUtNZnB4UnREQTI1bDZMc3NCc01pdjlkXG5takdrcWNlOG9VM1VCbmZhSGhMczdzU0kzdWZoQmZBSmxaSjhtWXQyNzRIU2lKRlJLVlZKS2gwSVVwOG1BVEdlXG4razFDV3E5dkJwYUVBQS9HLzBRK1lVS0xoUUtCZ1FEeVU2S3lSVDk5OXpRODkrWFZMNytlampzeFljdlN6MFlFXG5Eb2FORGtMZitxU0dNbW16NTZiRy9DUVJvd0Q1Yit6d2QyajBCb3lyeVRHa1FmZlZ0cVRkRGorS3cyZEtibzZIXG5aWThtV0pxQmEvUlZzazFKY2MzS0dYM1RsZDBIUExJUUdPMmgzMzJFZDBTR2FZWGJWektsc1U2MlhNcGcxKzdyXG4yaVZwRko2ZmtRS0JnUUNHSHVsNXd2V2NxZFlhMHZKLy82NFdjeXppa05rUkh4OFhGald1T1JhTndQVjJlbVIwXG5YVFNLSjBaV21UOGJrUFZSbTR4L05uU3Robm91L2xUMys3VnljNWoweEsyb3hLTjh4SUdYUzhpZDZnUjgyTzQ5XG5mYVhTVDFOZTd1cFpXMlRvQ29kZGZoYityWWZsakM5WjN0eUVDTnZXNktVWGpxVVBDZVZ0bjFWa3RRS0JnRzJRXG40ejgza0Qya1NEcEkyK0pJZEp0OE04ZGdNSWhncjRlbUNiQTlnbjlERktDWXFySnRTenN0UmlHelVmMTJYZXRjXG5FbGhEbmRjT1lTT2pzQ3N4S2RuSlYzR21hRTEvTDNLSXVQRGRudjVsa1ZRdUNrUHE4T0V3SlhSRmptcDNSd3VBXG5PZkcyMjBuSm8zSWl4Q01vaWYzZzdYWUcvbnBMSi92NzVtNWNwRndCQW9HQkFNNlQzdVdaS3MvS3FiWEZCVXd5XG5Udkd6c2pRUkE0cGxqVXcyR2RwbEY0MWdacENmRUxialhSSE1FRnh5THZrWHp0UWFndnNWalFUYlNRaFRZMVFSXG5IQ2xiK3FTRHFlakowZy9ZNnlpbGM1NDFyMWwwRTNXQjY3THhNaFhRTWdiTUU4SlhmUllwNmozUzQ3cVUvK2JlXG5NSmRmTHQvbVZwckFmN0I4a3pFTzVXOCtcbi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS1cbiIsCiAgImNsaWVudF9lbWFpbCI6ICJjbGF1ZGUtZGV2QG1vbmljYS1kZXYtMzkyNjA5LmlhbS5nc2VydmljZWFjY291bnQuY29tIiwKICAiY2xpZW50X2lkIjogIjEwNzE2Njk3NjQ4MTIyOTEzMTcxMCIsCiAgImF1dGhfdXJpIjogImh0dHBzOi8vYWNjb3VudHMuZ29vZ2xlLmNvbS9vL29hdXRoMi9hdXRoIiwKICAidG9rZW5fdXJpIjogImh0dHBzOi8vb2F1dGgyLmdvb2dsZWFwaXMuY29tL3Rva2VuIiwKICAiYXV0aF9wcm92aWRlcl94NTA5X2NlcnRfdXJsIjogImh0dHBzOi8vd3d3Lmdvb2dsZWFwaXMuY29tL29hdXRoMi92MS9jZXJ0cyIsCiAgImNsaWVudF94NTA5X2NlcnRfdXJsIjogImh0dHBzOi8vd3d3Lmdvb2dsZWFwaXMuY29tL3JvYm90L3YxL21ldGFkYXRhL3g1MDkvY2xhdWRlLWRldiU0MG1vbmljYS1kZXYtMzkyNjA5LmlhbS5nc2VydmljZWFjY291bnQuY29tIiwKICAidW5pdmVyc2VfZG9tYWluIjogImdvb2dsZWFwaXMuY29tIgp9"
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
