package jwt

import (
	"testing"
	"time"

	"kloggerx-server/config"
)

func setupTestConfig() {
	config.Cfg = &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-jwt-secret-for-unit-testing",
			Expire: 2 * time.Hour,
		},
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	setupTestConfig()

	tests := []struct {
		name     string
		userID   uint
		username string
		role     string
	}{
		{"admin user", 1, "admin", "admin"},
		{"normal user", 42, "zhangsan", "user"},
		{"editor user", 100, "editor01", "editor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.username, tt.role)
			if err != nil {
				t.Fatalf("GenerateToken failed: %v", err)
			}

			if token == "" {
				t.Fatal("generated token should not be empty")
			}

			// 解析 token
			claims, err := ParseToken(token)
			if err != nil {
				t.Fatalf("ParseToken failed: %v", err)
			}

			if claims.UserID != tt.userID {
				t.Errorf("UserID mismatch: got %d, want %d", claims.UserID, tt.userID)
			}
			if claims.Username != tt.username {
				t.Errorf("Username mismatch: got %q, want %q", claims.Username, tt.username)
			}
			if claims.Role != tt.role {
				t.Errorf("Role mismatch: got %q, want %q", claims.Role, tt.role)
			}
		})
	}
}

func TestExpiredToken(t *testing.T) {
	// 配置极短过期时间
	config.Cfg = &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-jwt-secret-for-unit-testing",
			Expire: -1 * time.Hour, // 已过期
		},
	}

	token, err := GenerateToken(1, "test", "user")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = ParseToken(token)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestTamperedToken(t *testing.T) {
	setupTestConfig()

	token, err := GenerateToken(1, "admin", "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// 篡改 token（修改最后几个字符）
	tampered := token[:len(token)-4] + "XXXX"
	_, err = ParseToken(tampered)
	if err == nil {
		t.Error("expected error for tampered token, got nil")
	}
}

func TestInvalidToken(t *testing.T) {
	setupTestConfig()

	_, err := ParseToken("completely.invalid.token")
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}

	_, err = ParseToken("")
	if err == nil {
		t.Error("expected error for empty token, got nil")
	}
}

func TestWrongSecret(t *testing.T) {
	setupTestConfig()

	token, err := GenerateToken(1, "test", "user")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// 更换 secret
	config.Cfg.JWT.Secret = "different-secret-key-for-testing"

	_, err = ParseToken(token)
	if err == nil {
		t.Error("expected error when parsing with wrong secret, got nil")
	}
}
