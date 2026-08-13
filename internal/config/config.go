package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultUpstreamTimeout = 300 * time.Second
	defaultListenAddress   = ":8080"
)

type Config struct {
	ListenAddress string
	Upstream      UpstreamConfig
}

type UpstreamConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

func Load() (Config, error) {
	var errs []error

	listenAddress, err := loadListenAddress()
	if err != nil {
		errs = append(errs, err)
	}

	baseURLValue, err := requiredEnv("UPSTREAM_BASE_URL")
	if err != nil {
		errs = append(errs, err)
	}

	baseURL, err := normalizeBaseURL(baseURLValue)
	if err != nil {
		errs = append(errs, err)
	}

	apiKey, err := requiredEnv("UPSTREAM_API_KEY")
	if err != nil {
		errs = append(errs, err)
	}

	timeout, err := loadUpstreamTimeout()
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}

	return Config{
		ListenAddress: listenAddress,
		Upstream: UpstreamConfig{
			BaseURL: baseURL,
			APIKey:  apiKey,
			Timeout: timeout,
		},
	}, nil
}

func loadListenAddress() (string, error) {
	listenAddress, exists := os.LookupEnv("LISTEN_ADDRESS")
	if !exists {
		return defaultListenAddress, nil
	}

	if listenAddress = strings.TrimSpace(listenAddress); listenAddress == "" {
		return "", fmt.Errorf(" LISTEN_ADDRESS 不能为空")
	}

	return listenAddress, nil
}

func requiredEnv(name string) (string, error) {
	value, exists := os.LookupEnv(name)

	value = strings.TrimSpace(value)

	if !exists || value == "" {
		return "", fmt.Errorf("%s 是必填项", name)
	}

	return value, nil
}

func normalizeBaseURL(value string) (string, error) {
	if !strings.Contains(value, "://") {
		return "", fmt.Errorf("URL 缺少协议")
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf(" URL 解析失败: %w", err)
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf(" URL 使用错误协议: %s", parsed.Scheme)
	}

	if parsed.Hostname() == "" {
		return "", fmt.Errorf(" URL 缺少主机名")
	}

	if parsed.User != "" {
		return "", fmt.Errorf(" URL 不得包含User信息: %s", parsed.User)
	}

	if parsed.RawQuery != "" || parsed.ForceQuery {
		return "", fmt.Errorf(" URL 不得包含查询信息: %s", parsed.RawQuery)
	}

	if parsed.Fargment != "" {
		return "", fmt.Errorf(" URL 不得包含片段: %s", parsed.Fargment)
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawPath = strings.TrimRight(parsed.RawPath, "/")

	return parsed.String(), nil
}

func loadUpstreamTimeout() (time.Duration, error) {
	value, exists := os.LookupEnv("UPSTREAM_TIMEOUT")
	if !exists {
		return defaultUpstreamTimeout, nil
	}

	if value = strings.TrimSpace("value"); value == "" {
		return 0, fmt.Errorf("配置超时不得为零")
	}

	timeout, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("配置超时解析失败: %w", err)
	}

	if timeout <= 0 {
		return 0, fmt.Errorf("配置超时 %v 必须大于零", timeout)
	}

	return timeout, nil
}
