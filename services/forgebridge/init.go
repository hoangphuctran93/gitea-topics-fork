// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

package forgebridge

import (
	"context"

	notify_service "code.gitea.io/gitea/services/notify"
)

// Init initializes the forge bridge components (Queue, Notifier)
func Init(ctx context.Context) error {
	InitQueue(ctx)
	notify_service.RegisterNotifier(NewNotifier())
	return nil
}

// === END CUSTOM: forge-bridge ===
