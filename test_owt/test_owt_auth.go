package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func testOWTAuth() {
	fmt.Println("Starting OWT auth test...")
	// OWT 服务配置
	serviceID := "698486aae7bc7a15d1643ee8" // 使用配置文件中的 superserviceID
	serviceKey := "BipVGJFNgCX9uR4s+UAtdUvxOBiq9hrjcbPL7FHgBcxE0heJ3EoCQPirtPLJALbiJ8canAzGm9wXi/7G7slUXnNaB/C94DGvSXfXKOkR17g2OnVpxnub1Z1BOvXPh9aBD3/a456Nl0vMqb6zslP/NMLcYxcVg9RzukYV/V19RJw="
	realm := "http://marte3.dit.upm.es" // 使用 OWT 服务期望的 realm
	baseURL := "https://owt.yujj888.top"

	fmt.Printf("ServiceID: %s\n", serviceID)
	fmt.Printf("ServiceKey length: %d\n", len(serviceKey))
	fmt.Printf("Realm: %s\n", realm)
	fmt.Printf("Base URL: %s\n", baseURL)

	// 生成认证头 - 传递 HTTP 方法和请求路径
	method := "POST"
	path := "/v1/rooms"
	fmt.Printf("Method: %s\n", method)
	fmt.Printf("Path: %s\n", path)

	fmt.Println("Generating auth header...")
	authHeader := getOWTAuthorizationHeader(serviceID, serviceKey, realm, method, path)
	if authHeader == "" {
		fmt.Println("Error: Failed to generate auth header")
		return
	}
	fmt.Println("\nGenerated auth header:")
	fmt.Println(authHeader)

	// 测试创建房间
	fmt.Println("Testing create room...")
	testCreateRoom(baseURL, authHeader)
	fmt.Println("Test completed")
}

func main() {
	testOWTAuth()
}

func getOWTAuthorizationHeader(serviceID, serviceKey, realm, method, path string) string {
	nonce := rand.Intn(99999)
	authTimestamp := time.Now().Unix() // 使用秒级时间戳

	// 解码 ServiceKey
	serviceKeyDecoded, err := base64.StdEncoding.DecodeString(serviceKey)
	if err != nil {
		fmt.Printf("ServiceKey 解码失败: %v\n", err)
		return ""
	}

	// 计算签名 - 只使用 realm、timestamp 和 nonce
	signData := fmt.Sprintf("%s,%d,%d", realm, authTimestamp, nonce)
	h := hmac.New(sha256.New, serviceKeyDecoded)
	h.Write([]byte(signData))
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 构建认证头 - 不使用双引号包围值
	authHeader := fmt.Sprintf("MAuth realm=%s,mauth_signature_method=HMAC_SHA256,mauth_serviceid=%s,mauth_cnonce=%d,mauth_timestamp=%d,mauth_signature=%s",
		realm, serviceID, nonce, authTimestamp, sign)

	// 打印认证参数
	fmt.Println("Auth parameters:")
	fmt.Printf("realm: %s\n", realm)
	fmt.Printf("serviceID: %s\n", serviceID)
	fmt.Printf("nonce: %d\n", nonce)
	fmt.Printf("timestamp: %d\n", authTimestamp)
	fmt.Printf("method: %s\n", method)
	fmt.Printf("path: %s\n", path)
	fmt.Printf("signData: %s\n", signData)
	fmt.Printf("signature: %s\n", sign)
	fmt.Printf("serviceKey length: %d\n", len(serviceKey))
	fmt.Printf("serviceKeyDecoded length: %d\n", len(serviceKeyDecoded))

	return authHeader
}

func testCreateRoom(baseURL, authHeader string) {
	fmt.Println("\nTesting create room...")

	// 创建请求
	reqURL := fmt.Sprintf("%s/v1/rooms", baseURL)
	req, err := http.NewRequest("POST", reqURL, bytes.NewBufferString(`{"name": "测试房间"}`))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	// 发送请求
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	buffer := make([]byte, 1024)
	n, _ := resp.Body.Read(buffer)
	responseBody := string(buffer[:n])

	// 打印响应
	fmt.Printf("Response status code: %d\n", resp.StatusCode)
	fmt.Printf("Response body: %s\n", responseBody)
}
