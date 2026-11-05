// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"time"

	ghapi "github.com/google/go-github/v89/github"
)

// gitHubPublicRawClaims projects a GitHub user's public profile fields into a map
// using a strict static allowlist. Private account metrics, plans, security attributes,
// private counts, and unverified raw emails are excluded.
func gitHubPublicRawClaims(user *ghapi.User) map[string]any {
	if user == nil {
		return nil
	}

	raw := make(map[string]any)

	if user.ID != nil {
		raw["id"] = *user.ID
	}
	if user.Login != nil {
		raw["login"] = *user.Login
	}
	if user.NodeID != nil {
		raw["node_id"] = *user.NodeID
	}
	if user.AvatarURL != nil {
		raw["avatar_url"] = *user.AvatarURL
	}
	if user.GravatarID != nil {
		raw["gravatar_id"] = *user.GravatarID
	}
	if user.URL != nil {
		raw["url"] = *user.URL
	}
	if user.HTMLURL != nil {
		raw["html_url"] = *user.HTMLURL
	}
	if user.FollowersURL != nil {
		raw["followers_url"] = *user.FollowersURL
	}
	if user.FollowingURL != nil {
		raw["following_url"] = *user.FollowingURL
	}
	if user.GistsURL != nil {
		raw["gists_url"] = *user.GistsURL
	}
	if user.StarredURL != nil {
		raw["starred_url"] = *user.StarredURL
	}
	if user.SubscriptionsURL != nil {
		raw["subscriptions_url"] = *user.SubscriptionsURL
	}
	if user.OrganizationsURL != nil {
		raw["organizations_url"] = *user.OrganizationsURL
	}
	if user.ReposURL != nil {
		raw["repos_url"] = *user.ReposURL
	}
	if user.EventsURL != nil {
		raw["events_url"] = *user.EventsURL
	}
	if user.ReceivedEventsURL != nil {
		raw["received_events_url"] = *user.ReceivedEventsURL
	}
	if user.Type != nil {
		raw["type"] = *user.Type
	}
	if user.UserViewType != nil {
		raw["user_view_type"] = *user.UserViewType
	}
	if user.SiteAdmin != nil {
		raw["site_admin"] = *user.SiteAdmin
	}
	if user.Name != nil {
		raw["name"] = *user.Name
	}
	if user.Company != nil {
		raw["company"] = *user.Company
	}
	if user.Blog != nil {
		raw["blog"] = *user.Blog
	}
	if user.Location != nil {
		raw["location"] = *user.Location
	}
	if user.Hireable != nil {
		raw["hireable"] = *user.Hireable
	}
	if user.Bio != nil {
		raw["bio"] = *user.Bio
	}
	if user.TwitterUsername != nil {
		raw["twitter_username"] = *user.TwitterUsername
	}
	if user.PublicRepos != nil {
		raw["public_repos"] = *user.PublicRepos
	}
	if user.PublicGists != nil {
		raw["public_gists"] = *user.PublicGists
	}
	if user.Followers != nil {
		raw["followers"] = *user.Followers
	}
	if user.Following != nil {
		raw["following"] = *user.Following
	}
	if user.CreatedAt != nil {
		raw["created_at"] = user.CreatedAt.Time.UTC().Format(time.RFC3339)
	}
	if user.UpdatedAt != nil {
		raw["updated_at"] = user.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}

	return raw
}
