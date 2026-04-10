package rtc

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"go.uber.org/zap"
)

// 获取owt公用auth头
func (r *RTC) getOWTAuthorizationHeader(method, path string) string {
	strBuilder := strings.Builder{}

	nonce := rand.Intn(99999)
	authTimestamp := time.Now().Unix()                                                 // 使用秒级时间戳
	realm := "http://marte3.dit.upm.es"                                                // OWT 服务期望的 realm
	serviceID := strings.TrimSpace(strings.Trim(r.ctx.GetConfig().OWT.ServiceID, `"`)) // 去除空格和引号
	serviceKey := r.ctx.GetConfig().OWT.ServiceKey

	// 构建认证头 - 确保参数值被双引号包围（MAuth 规范要求）
	strBuilder.WriteString("MAuth ")
	strBuilder.WriteString("realm=\"" + realm + "\",")
	strBuilder.WriteString("mauth_signature_method=\"HMAC_SHA256\",")
	strBuilder.WriteString("mauth_serviceid=" + serviceID + ",")
	strBuilder.WriteString("mauth_cnonce=\"" + fmt.Sprintf("%d", nonce) + "\",")
	strBuilder.WriteString("mauth_timestamp=\"" + fmt.Sprintf("%d", authTimestamp) + "\",")

	// 计算签名 - 使用标准的 HMAC-SHA256 实现
	// 注意：ServiceKey 是 base64 编码的，需要先解码
	serviceKeyDecoded, err := base64.StdEncoding.DecodeString(serviceKey)
	if err != nil {
		r.Error("ServiceKey 解码失败", zap.Error(err))
		return ""
	}
	// 计算签名 - 只使用 realm、timestamp 和 nonce
	signData := fmt.Sprintf("%s,%d,%d", realm, authTimestamp, nonce)
	h := hmac.New(sha256.New, serviceKeyDecoded)
	h.Write([]byte(signData))
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	strBuilder.WriteString("mauth_signature=\"" + sign + "\"")

	authHeader := strBuilder.String()
	r.Info("Generated auth header", zap.String("header", authHeader))
	r.Info("Auth parameters",
		zap.String("realm", realm),
		zap.String("serviceID", serviceID),
		zap.Int("serviceKeyLength", len(serviceKey)),
		zap.Int("serviceKeyDecodedLength", len(serviceKeyDecoded)),
		zap.String("nonce", fmt.Sprintf("%d", nonce)),
		zap.String("timestamp", fmt.Sprintf("%d", authTimestamp)),
		zap.String("signData", signData),
		zap.String("signature", sign))

	return authHeader
}

func (r *RTC) post(path string, bodyMap map[string]interface{}) ([]byte, error) {
	owtConfig := r.ctx.GetConfig().OWT
	r.Info("OWT config", zap.Any("config", owtConfig))
	baseURL := owtConfig.URL
	if baseURL == "" {
		baseURL = "https://api.yujj888.top"
		r.Info("OWT URL为空，使用默认值", zap.String("url", baseURL))
	}
	reqURL := fmt.Sprintf("%s%s", baseURL, path)
	r.Info("POST request", zap.String("url", reqURL), zap.Any("body", bodyMap))

	// 设置超时时间
	timeoutClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewBufferString(util.ToJson(bodyMap)))
	if err != nil {
		r.Error("创建请求失败", zap.Error(err))
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	auth := r.getOWTAuthorizationHeader("POST", path)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)

	resp, err := timeoutClient.Do(req)
	if err != nil {
		r.Error("发送请求失败", zap.Error(err), zap.String("url", reqURL))
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		resultBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			r.Error("读取响应失败", zap.Error(err))
			return nil, fmt.Errorf("读取响应失败: %v", err)
		}
		r.Info("请求成功", zap.String("url", reqURL), zap.String("response", string(resultBytes)))
		return resultBytes, nil
	}

	resultBytes, _ := ioutil.ReadAll(resp.Body)
	responseStr := string(resultBytes)
	r.Error("请求状态错误", zap.Int("status_code", resp.StatusCode), zap.String("response", responseStr), zap.String("url", reqURL))
	return nil, fmt.Errorf("请求状态错误！->%d, 响应: %s", resp.StatusCode, responseStr)
}

func (r *RTC) getRoomPresenterToken(roomID, uid string) (string, error) {
	return r.getRoomToken(roomID, uid, RolePresenter.String())
}

// 获取指定用户在房间的token
func (r *RTC) getRoomToken(roomID string, uid string, role string) (string, error) {
	resp, err := r.post(fmt.Sprintf("/v1/rooms/%s/tokens", roomID), map[string]interface{}{
		"room": roomID,
		"user": uid,
		"role": role,
		"preference": map[string]string{
			"isp":    "isp",
			"region": "region",
		},
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}
