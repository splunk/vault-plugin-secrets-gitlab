// Copyright 2021 Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitlabtoken

import (
	"errors"
	"fmt"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const (
	clientTTL = 30 * time.Minute
)

// Client makes API calls to GitLab.
type Client interface {
	// ListProjectAccessToken(int) ([]*PAT, error)
	CreateProjectAccessToken(e *BaseTokenStorageEntry, t *time.Time) (*PAT, error)
	// RevokeProjectAccessToken(*BaseTokenStorageEntry) error
	Valid() bool
}

// GitlabClient calls the GitLab API and implements the provided Client interface.
type GitlabClient struct {
	client     *gitlab.Client
	expiration time.Time
}

var _ Client = &GitlabClient{}

// NewClient returns a new GitLab Client and any error if occurs.
func NewClient(config *ConfigStorageEntry) (*GitlabClient, error) {
	if config == nil {
		return nil, errors.New("gitlab backend configuration has not been set up")
	}

	gc := &GitlabClient{
		expiration: time.Now().Add(clientTTL),
	}

	opt := gitlab.WithBaseURL(config.BaseURL)
	if config.Token == "" {
		return nil, errors.New("token isn't configured")
	}

	c, err := gitlab.NewClient(config.Token, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitlab client iwht endpoint %s: %w", config.BaseURL, err)
	}

	gc.client = c

	return gc, nil
}

// Valid returns true if the client is not expired.
func (gc *GitlabClient) Valid() bool {
	return gc != nil && time.Now().Before(gc.expiration)
}

// CreateProjectAccessToken returns a new Project Access Token and/or any error.
func (gc *GitlabClient) CreateProjectAccessToken(tokenStorage *BaseTokenStorageEntry, expiresAt *time.Time) (*PAT, error) {
	opt := gitlab.CreateProjectAccessTokenOptions{
		Name:   &tokenStorage.Name,
		Scopes: &tokenStorage.Scopes,
	}
	if expiresAt != nil {
		expiration := gitlab.ISOTime(*expiresAt)
		opt.ExpiresAt = &expiration
	}

	if tokenStorage.AccessLevel != 0 {
		opt.AccessLevel = (*gitlab.AccessLevelValue)(&tokenStorage.AccessLevel)
	}

	pat, _, err := gc.client.ProjectAccessTokens.CreateProjectAccessToken(tokenStorage.ID, &opt)
	if err != nil {
		return nil, err
	}

	return pat, nil
}

// func (gc *gitlabClient) RevokeProjectAccessToken(tokenStorage *BaseTokenStorageEntry) error {
// 	return nil
// }
