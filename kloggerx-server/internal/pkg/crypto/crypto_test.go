package crypto

import (
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	// 空 key 应返回错误
	err := Init("")
	if err == nil {
		t.Fatal("expected error for empty key, got nil")
	}

	// 正好 32 字节的 key
	err = Init("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("expected no error for 32-byte key, got: %v", err)
	}

	// 非 32 字节的 key (使用 SHA-256 哈希)
	err = Init("short-key")
	if err != nil {
		t.Fatalf("expected no error for short key, got: %v", err)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	// 初始化
	if err := Init("test-key-for-unit-tests-32bytes!"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"short text", "hello"},
		{"chinese text", "你好世界"},
		{"special chars", "p@$$w0rd!#%&"},
		{"long text", strings.Repeat("a", 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(tt.input)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// 密文不应与原文相同（空字符串除外，密文肯定不为空）
			if tt.input != "" && encrypted == tt.input {
				t.Error("encrypted text should differ from plaintext")
			}

			// 密文不应为空
			if encrypted == "" && tt.input != "" {
				t.Error("encrypted text should not be empty for non-empty input")
			}

			// 解密
			decrypted, err := Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if decrypted != tt.input {
				t.Errorf("Decrypt mismatch: got %q, want %q", decrypted, tt.input)
			}
		})
	}
}

func TestDecryptInvalidData(t *testing.T) {
	if err := Init("test-key-for-unit-tests-32bytes!"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 无效 base64
	_, err := Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}

	// 有效 base64 但不是有效密文
	_, err = Decrypt("aGVsbG8=") // "hello" in base64
	if err == nil {
		t.Error("expected error for invalid ciphertext")
	}

	// 太短的数据
	_, err = Decrypt("YQ==") // "a" in base64
	if err == nil {
		t.Error("expected error for too-short ciphertext")
	}
}

func TestEncryptDifferentEachTime(t *testing.T) {
	if err := Init("test-key-for-unit-tests-32bytes!"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	plaintext := "same input every time"
	results := make(map[string]bool)

	for i := 0; i < 10; i++ {
		encrypted, err := Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encrypt failed on iteration %d: %v", i, err)
		}
		results[encrypted] = true
	}

	// 由于 nonce 随机，10 次加密结果应全部不同
	if len(results) != 10 {
		t.Errorf("expected 10 unique ciphertexts, got %d", len(results))
	}
}

func TestEncryptWithoutInit(t *testing.T) {
	// 重置 key
	encryptionKey = nil

	_, err := Encrypt("test")
	if err == nil {
		t.Error("expected error when key not initialized")
	}

	_, err = Decrypt("dGVzdA==")
	if err == nil {
		t.Error("expected error when key not initialized")
	}
}
