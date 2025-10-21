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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientFail(t *testing.T) {
	t.Parallel()
	t.Run("no config", func(t *testing.T) {
		t.Parallel()

		c, err := NewClient(nil)
		require.Error(t, err, "nil config should thrown an error when retrieving Gitlab client")
		assert.Nil(t, c, "NewClient should return nil client on error")
	})

	t.Run("empty config", func(t *testing.T) {
		t.Parallel()

		config := &ConfigStorageEntry{}
		c, err := NewClient(config)
		require.Error(t, err, "NewClient should return an error if config is missing auth")
		assert.Nil(t, c, "NewClient should return nil client on error")
	})
}

func TestValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		client   *GitlabClient
		asserter assert.BoolAssertionFunc
	}{
		{
			name: "valid",
			client: &GitlabClient{
				expiration: time.Now().Add(clientTTL),
			},
			asserter: assert.True,
		},
		{
			name: "expired",
			client: &GitlabClient{
				expiration: time.Now().Add(-1 * time.Minute),
			},
			asserter: assert.False,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.asserter(t, test.client.Valid())
		})
	}
}

type mockGitlabClient struct{}

var _ Client = &mockGitlabClient{}

func (ac *mockGitlabClient) Valid() bool {
	return true
}

//	func (ac *mockGitlabClient) ListProjectAccessToken(id int) ([]*PAT, error) {
//		return nil, nil
//	}
func (ac *mockGitlabClient) CreateProjectAccessToken(_ *BaseTokenStorageEntry, _ *time.Time) (*PAT, error) {
	return nil, nil
}

// func (ac *mockGitlabClient) RevokeProjectAccessToken(tokenStorage *BaseTokenStorageEntry) error {
// 	return nil
// }
