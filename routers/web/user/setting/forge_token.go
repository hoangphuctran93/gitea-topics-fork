// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

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
	tplForgeToken base.TplName = "user/settings/forge_token"
)

// ForgeToken gets the forge token settings page
func ForgeToken(ctx *context.Context) {
	platform := ctx.Params("platform")
	if platform == "" {
		platform = "github"
	}
	platformName := "GitHub"
	if platform == "gitea" {
		platformName = "Gitea.com"
	}
	ctx.Data["Title"] = ctx.Tr("settings.forge_token", platformName)
	ctx.Data["PageIsSettingsForgeToken"] = true
	ctx.Data["ForgePlatform"] = platform
	ctx.Data["ForgePlatformName"] = platformName

	// Find existing token
	token := new(forgebridge.UserForgeToken)
	has, err := forgebridge.GetUserForgeToken(ctx.Doer.ID, platform, token)
	if err != nil {
		ctx.ServerError("GetUserForgeToken", err)
		return
	}

	if has {
		ctx.Data["HasForgeToken"] = true
		ctx.Data["ShareForSearch"] = token.ShareForSearch
		ctx.Data["ShareForCloneInfo"] = token.ShareForCloneInfo
		ctx.Data["ForgeUsername"] = token.ForgeUsername

		// Decrypt and mask for UI
		decrypted, err := forgebridge.DecryptToken(token.TokenEncrypted)
		if err == nil {
			ctx.Data["MaskedForgeToken"] = forgebridge.MaskToken(decrypted)
		} else {
			ctx.Data["MaskedForgeToken"] = "Error decrypting"
		}
	} else {
		// Calculate quota if no token
		quota := new(forgebridge.UserTokenQuota)
		hasQ, err := forgebridge.GetTodayUserQuota(ctx.Doer.ID, quota)
		if err != nil {
			ctx.ServerError("GetTodayUserQuota", err)
			return
		}

		limit := 50 // Default
		used := 0
		if hasQ {
			limit = quota.RequestsLimit
			used = quota.RequestsUsed
		}

		ctx.Data["QuotaLimit"] = limit
		ctx.Data["QuotaUsed"] = used
		ctx.Data["QuotaPercentage"] = float64(used) / float64(limit) * 100
	}

	ctx.HTML(http.StatusOK, tplForgeToken)
}

// ForgeTokenPost handles updating the forge token
func ForgeTokenPost(ctx *context.Context) {
	platform := ctx.Params("platform")
	if platform == "" {
		platform = "github"
	}
	platformName := "GitHub"
	if platform == "gitea" {
		platformName = "Gitea.com"
	}
	ctx.Data["Title"] = ctx.Tr("settings.forge_token", platformName)
	ctx.Data["PageIsSettingsForgeToken"] = true

	if ctx.FormBool("delete_token") {
		if err := forgebridge.DeleteUserForgeToken(ctx.Doer.ID, platform); err != nil {
			ctx.ServerError("DeleteUserForgeToken", err)
			return
		}
		ctx.Flash.Success(ctx.Tr("settings.forge_token_deleted", platformName))
		ctx.Redirect(setting.AppSubURL + "/user/settings/forge_token/" + platform)
		return
	}

	newToken := ctx.FormString("forge_token")
	shareSearch := ctx.FormBool("share_for_search")
	shareClone := ctx.FormBool("share_for_clone_info")
	forgeUsername := ctx.FormString("forge_username")

	token := new(forgebridge.UserForgeToken)
	has, err := forgebridge.GetUserForgeToken(ctx.Doer.ID, platform, token)
	if err != nil {
		ctx.ServerError("GetUserForgeToken", err)
		return
	}

	if newToken != "" {
		if !forgebridge.ValidateTokenFormat(platform, newToken) {
			ctx.Flash.Error(ctx.Tr("settings.forge_token_invalid_format", platformName))
			ctx.Redirect(setting.AppSubURL + "/user/settings/forge_token/" + platform)
			return
		}
		encrypted, err := forgebridge.EncryptToken(newToken)
		if err != nil {
			ctx.ServerError("EncryptToken", err)
			return
		}
		token.TokenEncrypted = encrypted
	}

	token.Platform = platform
	token.ShareForSearch = shareSearch
	token.ShareForCloneInfo = shareClone
	token.ForgeUsername = forgeUsername
	token.UserID = ctx.Doer.ID

	if has {
		if err := forgebridge.UpdateUserForgeToken(token); err != nil {
			ctx.ServerError("UpdateUserForgeToken", err)
			return
		}
	} else {
		if newToken == "" {
			ctx.Flash.Error(ctx.Tr("settings.forge_token_required", platformName))
			ctx.Redirect(setting.AppSubURL + "/user/settings/forge_token/" + platform)
			return
		}
		token.CreatedUnix = timeutil.TimeStampNow()
		if err := forgebridge.InsertUserForgeToken(token); err != nil {
			ctx.ServerError("InsertUserForgeToken", err)
			return
		}
	}

	ctx.Flash.Success(ctx.Tr("settings.forge_token_updated", platformName))
	ctx.Redirect(setting.AppSubURL + "/user/settings/forge_token/" + platform)
}

// === END CUSTOM: forge-bridge ===
