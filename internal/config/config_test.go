package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

const testTimeout = 10 * time.Second
const secret = "DO_NOT_LEAK_THIS_SECRET"

type testLoadStruct struct {
	name       string
	baseURL    string
	timeout    string
	address    string
	apiKey     string
	wantConfig Config
	wantErr    bool
}

func TestConfigLoad(t *testing.T) {
	tests := []testLoadStruct{
		{
			name:    "正确情况",
			baseURL: "http://test.api.com",
			timeout: "10s",
			address: ":8787",
			apiKey:  "test-key",
			wantConfig: Config{
				ListenAddress: ":8787",
				Upstream: UpstreamConfig{
					BaseURL: "http://test.api.com",
					APIKey:  "test-key",
					Timeout: testTimeout,
				},
			},
			wantErr: false,
		},
		{
			name:    "BaseURL末尾有多个斜杠",
			baseURL: "http://test.api.com//",
			timeout: "10s",
			address: ":8787",
			apiKey:  "test-key",
			wantConfig: Config{
				ListenAddress: ":8787",
				Upstream: UpstreamConfig{
					BaseURL: "http://test.api.com",
					APIKey:  "test-key",
					Timeout: testTimeout,
				},
			},
			wantErr: false,
		},
		{
			name:    "APIKey前后包含空格但不是空白值",
			baseURL: "http://test.api.com",
			timeout: "10s",
			address: ":8787",
			apiKey:  "  test-key  ",
			wantConfig: Config{
				ListenAddress: ":8787",
				Upstream: UpstreamConfig{
					BaseURL: "http://test.api.com",
					APIKey:  "  test-key  ",
					Timeout: testTimeout,
				},
			},
			wantErr: false,
		},
		{
			name:    "缺失非必填配置",
			baseURL: "http://test.api.com",
			apiKey:  "test-key",
			wantConfig: Config{
				ListenAddress: defaultListenAddress,
				Upstream: UpstreamConfig{
					BaseURL: "http://test.api.com",
					APIKey:  "test-key",
					Timeout: defaultUpstreamTimeout,
				},
			},
			wantErr: false,
		},
		{
			name:       "URL缺少Scheme",
			baseURL:    "test.api.com",
			timeout:    "10s",
			address:    ":8787",
			apiKey:     "test-key",
			wantConfig: Config{},
			wantErr:    true,
		},
		{
			name:       "URL使用非HTTP/HTTPS协议",
			baseURL:    "zttp://test.api.com",
			timeout:    "10s",
			address:    ":8787",
			apiKey:     "test-key",
			wantConfig: Config{},
			wantErr:    true,
		},
		{
			name:       "URL缺少Host",
			baseURL:    "http://",
			timeout:    "10s",
			address:    ":8787",
			apiKey:     "test-key",
			wantConfig: Config{},
			wantErr:    true,
		},
		{
			name:       "Timeout无法解析",
			baseURL:    "http://test.api.com",
			timeout:    "10",
			address:    ":8787",
			apiKey:     "test-key",
			wantConfig: Config{},
			wantErr:    true,
		},
		{
			name:       "Timeout等于0",
			baseURL:    "http://test.api.com",
			timeout:    "0s",
			address:    ":8787",
			apiKey:     "test-key",
			wantConfig: Config{},
			wantErr:    true,
		},
		{
			name:       "Timeout小于0",
			baseURL:    "http://test.api.com",
			timeout:    "-1s",
			address:    ":8787",
			apiKey:     "test-key",
			wantConfig: Config{},
			wantErr:    true,
		},
		{
			name:       "缺失BaseURL",
			timeout:    "10s",
			address:    ":8787",
			apiKey:     "test-key",
			wantConfig: Config{},
			wantErr:    true,
		},
		{
			name:       "缺失APIKey",
			baseURL:    "http://test.api.com",
			timeout:    "10s",
			address:    ":8787",
			wantConfig: Config{},
			wantErr:    true,
		},
		{
			name:       "BaseURL存在Query",
			baseURL:    "http://test.api.com?" + secret,
			timeout:    "10s",
			address:    ":8787",
			apiKey:     "test-key",
			wantConfig: Config{},
			wantErr:    true,
		},
		{
			name:       "BaseURL存在Fragment",
			baseURL:    "http://test.api.com#" + secret,
			timeout:    "10s",
			address:    ":8787",
			apiKey:     "test-key",
			wantConfig: Config{},
			wantErr:    true,
		},
		{
			name:       "BaseURL存在User",
			baseURL:    "http://user:" + secret + "@test.api.com",
			timeout:    "10s",
			address:    ":8787",
			apiKey:     "test-key",
			wantConfig: Config{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleRun(t, tt)
		})
	}
}

func ruleRun(t *testing.T, tt testLoadStruct) {
	t.Helper()

	switch tt.name {
	case "缺失非必填配置":
		t.Setenv("UPSTREAM_BASE_URL", tt.baseURL)
		t.Setenv("UPSTREAM_API_KEY", tt.apiKey)
	case "缺失BaseURL":
		t.Setenv("UPSTREAM_TIMEOUT", tt.timeout)
		t.Setenv("LISTEN_ADDRESS", tt.address)
		t.Setenv("UPSTREAM_API_KEY", tt.apiKey)
	case "缺失APIKey":
		t.Setenv("UPSTREAM_BASE_URL", tt.baseURL)
		t.Setenv("UPSTREAM_TIMEOUT", tt.timeout)
		t.Setenv("LISTEN_ADDRESS", tt.address)
	default:
		t.Setenv("UPSTREAM_BASE_URL", tt.baseURL)
		t.Setenv("UPSTREAM_TIMEOUT", tt.timeout)
		t.Setenv("LISTEN_ADDRESS", tt.address)
		t.Setenv("UPSTREAM_API_KEY", tt.apiKey)

	}

	result, err := Load()
	if tt.wantErr {
		if err == nil {
			t.Fatal("期望出错, 实际为nil")
		}

		empty := Config{}
		if !reflect.DeepEqual(result, empty) {
			t.Fatalf("want: %v, got: %v", empty, result)
		}

		if strings.Contains(err.Error(), secret) {
			t.Fatalf("错误 %v 不得携带敏感信息: %v", err, secret)
		}

		return
	}

	if err != nil {
		t.Fatalf("期望正确, 实际Err: %v", err)
	}

	if !reflect.DeepEqual(result, tt.wantConfig) {
		t.Fatalf("want: %v, got: %v", tt.wantConfig, result)
	}

}
