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
	tplGithubToken base.TplName = "user/settings/github_token"
)

// GithubToken gets the GitHub token settings page
func GithubToken(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings.github_token")
	ctx.Data["PageIsSettingsGithubToken"] = true

	// Find existing token
	token := new(forgebridge.UserGithubToken)
	has, err := forgebridge.GetUserGithubToken(ctx.Doer.ID, token)
	if err != nil {
		ctx.ServerError("GetUserGithubToken", err)
		return
	}

	if has {
		ctx.Data["HasGithubToken"] = true
		ctx.Data["ShareForSearch"] = token.ShareForSearch
		ctx.Data["ShareForCloneInfo"] = token.ShareForCloneInfo

		// Decrypt and mask for UI
		decrypted, err := forgebridge.DecryptToken(token.TokenEncrypted)
		if err == nil {
			ctx.Data["MaskedGithubToken"] = forgebridge.MaskToken(decrypted)
		} else {
			ctx.Data["MaskedGithubToken"] = "Error decrypting"
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

	ctx.HTML(http.StatusOK, tplGithubToken)
}

// GithubTokenPost handles updating the GitHub token
func GithubTokenPost(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings.github_token")
	ctx.Data["PageIsSettingsGithubToken"] = true

	if ctx.FormBool("delete_token") {
		if err := forgebridge.DeleteUserGithubToken(ctx.Doer.ID); err != nil {
			ctx.ServerError("DeleteUserGithubToken", err)
			return
		}
		ctx.Flash.Success(ctx.Tr("settings.github_token_deleted"))
		ctx.Redirect(setting.AppSubURL + "/user/settings/github_token")
		return
	}

	newToken := ctx.FormString("github_token")
	shareSearch := ctx.FormBool("share_for_search")
	shareClone := ctx.FormBool("share_for_clone_info")

	token := new(forgebridge.UserGithubToken)
	has, err := forgebridge.GetUserGithubToken(ctx.Doer.ID, token)
	if err != nil {
		ctx.ServerError("GetUserGithubToken", err)
		return
	}

	if newToken != "" {
		encrypted, err := forgebridge.EncryptToken(newToken)
		if err != nil {
			ctx.ServerError("EncryptToken", err)
			return
		}
		token.TokenEncrypted = encrypted
	}

	token.ShareForSearch = shareSearch
	token.ShareForCloneInfo = shareClone
	token.UserID = ctx.Doer.ID

	if has {
		if err := forgebridge.UpdateUserGithubToken(token); err != nil {
			ctx.ServerError("UpdateUserGithubToken", err)
			return
		}
	} else {
		if newToken == "" {
			ctx.Flash.Error(ctx.Tr("settings.github_token_required"))
			ctx.Redirect(setting.AppSubURL + "/user/settings/github_token")
			return
		}
		token.CreatedUnix = timeutil.TimeStampNow()
		if err := forgebridge.InsertUserGithubToken(token); err != nil {
			ctx.ServerError("InsertUserGithubToken", err)
			return
		}
	}

	ctx.Flash.Success(ctx.Tr("settings.github_token_updated"))
	ctx.Redirect(setting.AppSubURL + "/user/settings/github_token")
}

// === END CUSTOM: forge-bridge ===
