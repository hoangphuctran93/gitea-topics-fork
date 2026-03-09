// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgebridge

import (
	"regexp"
	"strings"
)

// === BEGIN CUSTOM: forge-bridge ===
// Author: hoangphuctran93

var (
	reGitHubClassic = regexp.MustCompile(`^ghp_[a-zA-Z0-9]{36}$`)
	reGitHubFine    = regexp.MustCompile(`^github_pat_[a-zA-Z0-9]{22}_[a-zA-Z0-9]{59}$`)
	reGiteaSHA1     = regexp.MustCompile(`^[a-fA-F0-9]{40}$`)
)

// ValidateTokenFormat checks if the token matches the expected format for the platform
func ValidateTokenFormat(platform, token string) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}

	switch platform {
	case "github":
		// GitHub tokens can be classic (ghp_...) or fine-grained (github_pat_...)
		if strings.HasPrefix(token, "ghp_") {
			return reGitHubClassic.MatchString(token)
		}
		if strings.HasPrefix(token, "github_pat_") {
			return reGitHubFine.MatchString(token)
		}
		return false
	case "gitea":
		// Gitea.com personal access tokens are typically 40-character hex SHA-1
		return reGiteaSHA1.MatchString(token)
	default:
		// Unknown platform, allow for now or keep strict?
		// For safety, we keep it strict to known platforms.
		return false
	}
}

// === END CUSTOM: forge-bridge ===
