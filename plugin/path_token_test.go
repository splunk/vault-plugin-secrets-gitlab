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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

func TestAccToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (short)")
	}
	req, backend := newGitlabAccEnv(t)

	ID := envAsInt("GITLAB_PROJECT_ID", 1)

	t.Run("successfully create", func(t *testing.T) {
		d := map[string]interface{}{
			"id":         ID,
			"name":       "vault-test",
			"scopes":     []string{"read_api"},
			"token_type": "project",
		}
		resp, err := testIssueToken(t, backend, req, d)
		require.NoError(t, err)
		fmt.Println(resp.Error())
		require.False(t, resp.IsError())

		assert.NotEmpty(t, resp.Data["token"], "no token returned")
		assert.NotEmpty(t, resp.Data["id"], "no id returned")
		assert.Empty(t, resp.Data["expires_at"], "default is never(nil) for expires_at")
	})

	t.Run("successfully create with expiration", func(t *testing.T) {
		e := time.Now().Add(time.Hour * 24)
		d := map[string]interface{}{
			"id":         ID,
			"name":       "vault-test-expires",
			"scopes":     []string{"read_api"},
			"expires_at": e.Unix(),
			"token_type": "project",
		}
		resp, err := testIssueToken(t, backend, req, d)
		require.NoError(t, err)
		require.False(t, resp.IsError())

		assert.NotEmpty(t, resp.Data["token"], "no token returned")
		assert.NotEmpty(t, resp.Data["id"], "no id returned")
		assert.Contains(t, resp.Data["expires_at"].(time.Time).String(), e.Format("2006-01-02"))
	})

	t.Run("successfully create with access level", func(t *testing.T) {
		e := time.Now().Add(time.Hour * 24)
		d := map[string]interface{}{
			"id":           ID,
			"name":         "vault-test-access-level",
			"scopes":       []string{"read_api"},
			"access_level": 30,
			"expires_at":   e.Unix(),
			"token_type":   "project",
		}
		resp, err := testIssueToken(t, backend, req, d)
		require.NoError(t, err)
		require.False(t, resp.IsError())

		assert.NotEmpty(t, resp.Data["token"], "no token returned")
		assert.NotEmpty(t, resp.Data["id"], "no id returned")
		assert.NotEmpty(t, resp.Data["access_level"], "no access_level returned")
		assert.Contains(t, resp.Data["expires_at"].(time.Time).String(), e.Format("2006-01-02"))

		assert.Equal(t, gitlab.AccessLevelValue(30), resp.Data["access_level"])

	})

	t.Run("validation failure", func(t *testing.T) {
		t.Parallel()
		d := map[string]interface{}{
			"id": -1,
		}
		resp, err := testIssueToken(t, backend, req, d)
		require.NoError(t, err)
		require.True(t, resp.IsError())

		require.Contains(t, resp.Data["error"], "id is empty or invalid")
		require.Contains(t, resp.Data["error"], "name is empty")
		require.Contains(t, resp.Data["error"], "scopes are empty")
		require.Contains(t, resp.Data["error"], "token_type must be either")
	})

	t.Run("exceeding max token lifetime", func(t *testing.T) {
		t.Parallel()

		conf := map[string]interface{}{
			"max_ttl": fmt.Sprintf("%dh", 7*24), // 7 days
		}

		testConfigUpdate(t, backend, req.Storage, conf)

		e := time.Now().Add(time.Hour * 14 * 24)
		d := map[string]interface{}{
			"id":         ID,
			"name":       "vault-test-exceeding-lifetime",
			"scopes":     []string{"read_api"},
			"expires_at": e.Unix(),
			"token_type": "project",
		}
		resp, err := testIssueToken(t, backend, req, d)
		require.NoError(t, err)
		require.True(t, resp.IsError())
	})

}

// create the token given the parameters
func testIssueToken(t *testing.T, b logical.Backend, req *logical.Request, data map[string]interface{}) (*logical.Response, error) {
	req.Operation = logical.CreateOperation
	req.Path = pathPatternToken
	req.Data = data

	resp, err := b.HandleRequest(context.Background(), req)

	return resp, err
}
