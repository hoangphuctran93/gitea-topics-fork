// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package admin

import (
	"net/http"

	"code.gitea.io/gitea/models/forgebridge"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/templates"
	"code.gitea.io/gitea/modules/timeutil"
	"code.gitea.io/gitea/services/context"
)

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

const (
	tplForgeTokens templates.TplName = "admin/forge_tokens"
)

// ForgeTokens shows the admin panel for forge tokens pool
func ForgeTokens(ctx *context.Context) {
	platform := ctx.PathParam("platform")
	if platform == "" {
		platform = "github"
	}
	platformName := "GitHub"
	if platform == "gitea" {
		platformName = "Gitea.com"
	}
	ctx.Data["Title"] = ctx.Tr("admin.forge_tokens", platformName)
	ctx.Data["PageIsAdminForgeTokens"] = true
	ctx.Data["ForgePlatform"] = platform
	ctx.Data["ForgePlatformName"] = platformName

	tokens, err := forgebridge.GetAllAdminTokens(ctx, platform)
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
	ctx.HTML(http.StatusOK, tplForgeTokens)
}

// ForgeTokensPost handles admin adding/deleting tokens
func ForgeTokensPost(ctx *context.Context) {
	platform := ctx.PathParam("platform")
	if platform == "" {
		platform = "github"
	}
	platformName := "GitHub"
	if platform == "gitea" {
		platformName = "Gitea.com"
	}
	ctx.Data["Title"] = ctx.Tr("admin.forge_tokens", platformName)
	ctx.Data["PageIsAdminForgeTokens"] = true

	if ctx.FormBool("delete_token") {
		id := ctx.FormInt64("id")
		if err := forgebridge.DeleteAdminForgeToken(ctx, id); err != nil {
			ctx.ServerError("DeleteAdminForgeToken", err)
			return
		}
		ctx.Flash.Success(ctx.Tr("settings.forge_token_deleted", platformName))
		ctx.Redirect(setting.AppSubURL + "/admin/forge_tokens/" + platform)
		return
	}

	// Add logic
	label := ctx.FormString("label")
	rawToken := ctx.FormString("token")

	if label == "" || rawToken == "" {
		ctx.Flash.Error(ctx.Tr("admin.forge_tokens_required"))
		ctx.Redirect(setting.AppSubURL + "/admin/forge_tokens/" + platform)
		return
	}

	if !forgebridge.ValidateTokenFormat(platform, rawToken) {
		ctx.Flash.Error(ctx.Tr("settings.forge_token_invalid_format", platformName))
		ctx.Redirect(setting.AppSubURL + "/admin/forge_tokens/" + platform)
		return
	}

	encrypted, err := forgebridge.EncryptToken(rawToken)
	if err != nil {
		ctx.ServerError("EncryptToken", err)
		return
	}

	token := &forgebridge.AdminForgeToken{
		Platform:           platform,
		Label:              label,
		TokenEncrypted:     encrypted,
		IsActive:           true,
		RateLimitRemaining: 5000,
	}
	token.CreatedUnix = timeutil.TimeStampNow()

	if err := forgebridge.InsertAdminForgeToken(ctx, token); err != nil {
		ctx.ServerError("InsertAdminForgeToken", err)
		return
	}

	ctx.Flash.Success(ctx.Tr("settings.forge_token_updated", platformName))
	ctx.Redirect(setting.AppSubURL + "/admin/forge_tokens/" + platform)
}

// === END CUSTOM: forge-bridge ===
