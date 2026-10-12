package management

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

func TestBuildAuthFileEntryRedactsProxyCredentials(t *testing.T) {
	h := NewHandlerWithoutConfigFilePath(&config.Config{}, nil)
	auth := &coreauth.Auth{
		ID:       "antigravity-user@example.com.json",
		FileName: "antigravity-user@example.com.json",
		Provider: "antigravity",
		ProxyURL: "socks5://proxy-user:proxy-password@proxy.example:1080/path",
		Attributes: map[string]string{
			"path": "/tmp/antigravity-user@example.com.json",
		},
	}

	entry := h.buildAuthFileEntry(auth, authFileQuotaContext{})
	if got := entry["proxy_address"]; got != "socks5://proxy.example:1080" {
		t.Fatalf("proxy_address = %#v, want %q", got, "socks5://proxy.example:1080")
	}
	if got := entry["proxy_configured"]; got != true {
		t.Fatalf("proxy_configured = %#v, want true", got)
	}
	if _, ok := entry["proxy_url"]; ok {
		t.Fatal("proxy_url must not be included in the list response")
	}
	if got := entry["proxy_address"].(string); got == "proxy-user" || got == "proxy-password" {
		t.Fatalf("proxy credentials leaked in proxy_address: %q", got)
	}
}

func TestBuildAuthFileEntryMarksDirectProxy(t *testing.T) {
	h := NewHandlerWithoutConfigFilePath(&config.Config{}, nil)
	auth := &coreauth.Auth{
		ID:       "antigravity-user@example.com.json",
		FileName: "antigravity-user@example.com.json",
		Provider: "antigravity",
		ProxyURL: "direct",
		Attributes: map[string]string{
			"path": "/tmp/antigravity-user@example.com.json",
		},
	}

	entry := h.buildAuthFileEntry(auth, authFileQuotaContext{})
	if got := entry["proxy_configured"]; got != false {
		t.Fatalf("proxy_configured = %#v, want false", got)
	}
	if _, ok := entry["proxy_address"]; ok {
		t.Fatal("direct proxy must not include proxy_address")
	}
}
