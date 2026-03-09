// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package admin

import (
	"net/http"

	"code.gitea.io/gitea/models/forgebridge"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/timeutil"
	"code.gitea.io/gitea/services/context"
)

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

const (
	tplGithubTokens base.TplName = "admin/github_token"
)

// GithubTokens shows the admin panel for GitHub tokens pool
func GithubTokens(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.github_tokens")
	ctx.Data["PageIsAdminGithubTokens"] = true

	tokens, err := forgebridge.GetAllAdminTokens()
	if err != nil {
		ctx.ServerError("GetAllAdminTokens", err)
		return
	}

	type displayToken struct {
		ID                 int64
		Label              string
		MaskedToken        string
		RequestCount       int64
		RateLimitRemaining int
		IsActive           bool
	}

	displayTokens := make([]displayToken, 0, len(tokens))
	for _, t := range tokens {
		decrypted, err := forgebridge.DecryptToken(t.TokenEncrypted)
		masked := "Error"
		if err == nil {
			masked = forgebridge.MaskToken(decrypted)
		}

		displayTokens = append(displayTokens, displayToken{
			ID:                 t.ID,
			Label:              t.Label,
			MaskedToken:        masked,
			RequestCount:       t.RequestCount,
			RateLimitRemaining: t.RateLimitRemaining,
			IsActive:           t.IsActive,
		})
	}

	ctx.Data["AdminTokens"] = displayTokens
	ctx.HTML(http.StatusOK, tplGithubTokens)
}

// GithubTokensPost handles admin adding/deleting tokens
func GithubTokensPost(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.github_tokens")
	ctx.Data["PageIsAdminGithubTokens"] = true

	// Delete logic
	if ctx.Req.URL.Path == setting.AppSubURL+"/admin/forgebridge/github_tokens/delete" {
		id := ctx.FormInt64("id")
		if id > 0 {
			if err := forgebridge.DeleteAdminGithubToken(id); err != nil {
				ctx.ServerError("DeleteAdminGithubToken", err)
				return
			}
			ctx.Flash.Success(ctx.Tr("admin.github_tokens_deleted"))
		}
		ctx.Redirect(setting.AppSubURL + "/admin/forgebridge/github_tokens")
		return
	}

	// Add logic
	label := ctx.FormString("label")
	rawToken := ctx.FormString("token")

	if label == "" || rawToken == "" {
		ctx.Flash.Error(ctx.Tr("admin.github_tokens_required"))
		ctx.Redirect(setting.AppSubURL + "/admin/forgebridge/github_tokens")
		return
	}

	encrypted, err := forgebridge.EncryptToken(rawToken)
	if err != nil {
		ctx.ServerError("EncryptToken", err)
		return
	}

	newToken := &forgebridge.AdminGithubToken{
		Label:              label,
		TokenEncrypted:     encrypted,
		IsActive:           true,
		RateLimitRemaining: 5000,
		CreatedUnix:        timeutil.TimeStampNow(),
	}

	if err := forgebridge.InsertAdminGithubToken(newToken); err != nil {
		ctx.ServerError("InsertAdminGithubToken", err)
		return
	}

	ctx.Flash.Success(ctx.Tr("admin.github_tokens_added"))
	ctx.Redirect(setting.AppSubURL + "/admin/forgebridge/github_tokens")
}

// === END CUSTOM: forge-bridge ===
