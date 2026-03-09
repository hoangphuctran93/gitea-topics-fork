// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

// GithubTokenForm represents the form for updating a GitHub token
type GithubTokenForm struct {
	Token             string `binding:"MaxSize(255)"`
	ShareForSearch    bool
	ShareForCloneInfo bool
	DeleteToken       bool
}

// === END CUSTOM: forge-bridge ===
