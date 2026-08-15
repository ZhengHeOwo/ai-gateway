package upstream

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientChatCompletion(t *testing.T) {
	var gotMethod, gotPath, gotAuth, gotContentType string
	var gotBody []byte

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer fakeServer.Close()

	client, err := NewClient(fakeServer.URL, "test-key", 30*time.Second)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	resp, err := client.ChatCompletion(context.Background(), []byte(`{"role":"user"}`))
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if gotMethod != http.MethodPost {
		t.Errorf("want: %v, got: %v", http.MethodPost, gotMethod)
	}
	if gotPath != "/v1/chat/completions" {
		t.Errorf("want: %v, got: %v", "/v1/chat/completions", gotPath)
	}
	if gotAuth != "Bearer "+"test-key" {
		t.Errorf("want: %v, got: %v", "bearer "+"test-key", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Errorf("want: %v, got: %v", "application/json", gotContentType)
	}
	if string(gotBody) != `{"role":"user"}` {
		t.Errorf("want: %v, got: %v", `{"role":"user"}`, string(gotBody))
	}

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("want: %v, got: %v", http.StatusTooManyRequests, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("读取响应体失败: %v", err)
	}

	if string(body) != `{"error":"rate limited"}` {
		t.Fatalf("want: %v, got: %v", `{"error":"rate limited"}`, string(body))
	}
}
