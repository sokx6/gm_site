package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
	"time"
)

// LskyClient 兰空图床客户端
type LskyClient struct {
	baseURL  string
	email    string
	password string
	token    string
	tokenMu  sync.RWMutex

	httpClient *http.Client
}

// tokenResponse 兰空 GET token 响应
type tokenResponse struct {
	Status bool `json:"status"`
	Data   struct {
		Token string `json:"token"`
	} `json:"data"`
}

// uploadResponse 兰空上传响应
type uploadResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Links struct {
			URL string `json:"url"`
		} `json:"links"`
	} `json:"data"`
}

// NewLskyClient 创建兰空客户端实例，httpClient 超时 30s
func NewLskyClient(baseURL, email, password string) *LskyClient {
	return &LskyClient{
		baseURL:  baseURL,
		email:    email,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ensureToken 确保 token 有效，如果为空则请求新的 token
func (c *LskyClient) ensureToken() error {
	// 快速路径：读锁检查
	c.tokenMu.RLock()
	if c.token != "" {
		c.tokenMu.RUnlock()
		return nil
	}
	c.tokenMu.RUnlock()

	// 写锁获取 token
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// 双重检查：可能其他 goroutine 已获取
	if c.token != "" {
		return nil
	}

	body := map[string]string{
		"email":    c.email,
		"password": c.password,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal token request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/tokens",
		"application/json",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return fmt.Errorf("request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request returned status %d", resp.StatusCode)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return fmt.Errorf("decode token response: %w", err)
	}

	// 即使 tr.Status 为 false，只要有 token 就继续
	if tr.Data.Token == "" {
		return fmt.Errorf("empty token in response")
	}

	c.token = tr.Data.Token
	return nil
}

// UploadImage 上传图片到兰空图床，返回图片 URL
func (c *LskyClient) UploadImage(file multipart.File, header *multipart.FileHeader) (lskyURL string, err error) {
	return c.uploadImage(file, header, true)
}

// uploadImage 内部上传方法，allowRetry 控制是否在 401 时重试
func (c *LskyClient) uploadImage(file multipart.File, header *multipart.FileHeader, allowRetry bool) (string, error) {
	if err := c.ensureToken(); err != nil {
		return "", fmt.Errorf("ensure token: %w", err)
	}

	// 读取当前 token
	c.tokenMu.RLock()
	token := c.token
	c.tokenMu.RUnlock()

	// 将文件指针移到开头
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("seek file: %w", err)
		}
	}

	// 构造 multipart/form-data 请求
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close writer: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/upload", &buf)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	// 401 时清空 token 并重试一次
	if resp.StatusCode == http.StatusUnauthorized && allowRetry {
		c.tokenMu.Lock()
		c.token = ""
		c.tokenMu.Unlock()

		// 将文件指针移回开头后重试
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
		return c.uploadImage(file, header, false)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var ur uploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&ur); err != nil {
		return "", fmt.Errorf("decode upload response: %w", err)
	}

	if !ur.Status {
		msg := ur.Message
		if msg == "" {
			msg = "unknown error"
		}
		return "", fmt.Errorf("upload failed: %s", msg)
	}

	return ur.Data.Links.URL, nil
}
