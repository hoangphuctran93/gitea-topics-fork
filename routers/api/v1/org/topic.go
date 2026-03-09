// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package org

import (
	"net/http"

	"code.gitea.io/gitea/models/db"
	repo_model "code.gitea.io/gitea/models/repo"
	"code.gitea.io/gitea/routers/api/v1/utils"
	"code.gitea.io/gitea/services/context"

	"xorm.io/builder"
)

// topicWithCount represents a topic with its usage count
type topicWithCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ListOrgTopics returns all unique topics across repos in an organization
func ListOrgTopics(ctx *context.APIContext) {
	// swagger:operation GET /orgs/{org}/topics organization orgListTopics
	// ---
	// summary: List all topics used by repositories in an organization
	// produces:
	//   - application/json
	// parameters:
	// - name: org
	//   in: path
	//   description: name of the organization
	//   type: string
	//   required: true
	// - name: prefix
	//   in: query
	//   description: filter topics by prefix (e.g. "cat:" or "dom:")
	//   type: string
	// - name: sort
	//   in: query
	//   description: sort by "count" (default) or "name"
	//   type: string
	// - name: limit
	//   in: query
	//   description: number of results to return
	//   type: integer
	// - name: page
	//   in: query
	//   description: page number of results to return (1-based)
	//   type: integer
	// responses:
	//   "200":
	//     description: "TopicListResponse"
	//     schema:
	//       type: object
	//       properties:
	//         topics:
	//           type: array
	//           items:
	//             type: object
	//             properties:
	//               name:
	//                 type: string
	//               count:
	//                 type: integer
	//         total:
	//           type: integer
	//   "404":
	//     "$ref": "#/responses/notFound"

	prefix := ctx.FormString("prefix")
	sortBy := ctx.FormString("sort")
	listOpts := utils.GetListOptions(ctx)

	orgID := ctx.Org.Organization.ID

	baseCond := builder.Eq{"repository.owner_id": orgID}
	if prefix != "" {
		baseCond["topic.name LIKE"] = prefix + "%"
	}

	orderBy := "COUNT(repo_topic.repo_id) DESC"
	if sortBy == "name" {
		orderBy = "topic.name ASC"
	}

	// Get topics with count
	var topics []topicWithCount
	sess := db.GetEngine(ctx).Table("topic").
		Join("INNER", "repo_topic", "repo_topic.topic_id = topic.id").
		Join("INNER", "repository", "repository.id = repo_topic.repo_id").
		Where(builder.Eq{"repository.owner_id": orgID}).
		GroupBy("topic.id, topic.name").
		OrderBy(orderBy).
		Select("topic.name, COUNT(repo_topic.repo_id) as count")

	if prefix != "" {
		sess = sess.And(builder.Like{"topic.name", prefix + "%"})
	}

	if listOpts.PageSize > 0 {
		sess = sess.Limit(listOpts.PageSize, (listOpts.Page-1)*listOpts.PageSize)
	}

	if err := sess.Find(&topics); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	// Get total count of distinct topics
	total, err := db.GetEngine(ctx).Table("topic").
		Join("INNER", "repo_topic", "repo_topic.topic_id = topic.id").
		Join("INNER", "repository", "repository.id = repo_topic.repo_id").
		Where(builder.Eq{"repository.owner_id": orgID}).
		GroupBy("topic.id").
		Count(new(repo_model.Topic))
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	if topics == nil {
		topics = make([]topicWithCount, 0)
	}

	ctx.SetTotalCountHeader(total)
	ctx.JSON(http.StatusOK, map[string]any{
		"topics": topics,
		"total":  total,
	})
}
