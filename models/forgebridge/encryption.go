// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgebridge

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"code.gitea.io/gitea/modules/setting"
)

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

var (
	ErrDecryptionFailed = errors.New("failed to decrypt token")
	ErrEncryptionFailed = errors.New("failed to encrypt token")
)

// getSecretKey returns the 32-byte secret key derived from setting.SecretKey.
// If the configured key is not exactly 32 bytes, we pad or truncate it to ensure AES-256 size.
func getSecretKey() []byte {
	key := []byte(setting.SecretKey)
	if len(key) == 32 {
		return key
	}

	finalKey := make([]byte, 32)
	copy(finalKey, key)
	return finalKey
}

// EncryptToken encrypts a plaintext token using AES-256-GCM.
func EncryptToken(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(getSecretKey())
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// DecryptToken decrypts an encrypted token string using AES-256-GCM.
func DecryptToken(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	cipherText, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(getSecretKey())
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", ErrDecryptionFailed
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// MaskToken returns a masked version of the token for UI display (e.g. ghp_••••xxxx)
// Assumes standard GitHub token format, roughly 40 characters for PAT.
func MaskToken(token string) string {
	if len(token) <= 8 {
		return "••••••••"
	}
	prefix := token[:4]
	suffix := token[len(token)-4:]
	middleLen := len(token) - 8
	if middleLen > 16 {
		middleLen = 16
	}
	return prefix + strings.Repeat("•", middleLen) + suffix
}

// PlatformMaskHint returns a masked token hint based on the platform type.
// It does NOT require decryption for platforms with known prefixes (github, gitlab).
// For Gitea tokens (hex strings without fixed prefix), it attempts decryption
// to show first/last 4 chars as a visual hint for the user.
func PlatformMaskHint(platform, encryptedToken string) string {
	if encryptedToken == "" {
		return ""
	}

	switch platform {
	case "github":
		// GitHub PATs (20 char hint on UI)
		return "ghp_••••••••••••••••••••"
	case "gitea":
		// Gitea tokens: try to decrypt and show first4 + mask + last4
		decrypted, err := DecryptToken(encryptedToken)
		if err == nil && len(decrypted) >= 8 {
			return MaskToken(decrypted)
		}
		// Fallback if decryption fails
		return "••••••••••••••••••••"
	case "gitlab":
		// GitLab PATs
		return "glpat-••••••••••••••••••••"
	default:
		return "••••••••••••••••••••"
	}
}

// === END CUSTOM: forge-bridge ===
