// Copyright  2021 Masahiro Yoshida
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
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccRevoke(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (short)")
	}
	req, backend := newGitlabAccEnv(t)

	ID := envAsInt("GITLAB_PROJECT_ID", 1)

	d := map[string]interface{}{
		"id":     ID,
		"name":   "vault-test-revoke",
		"scopes": []string{"read_api"},
	}

	t.Run("success", func(t *testing.T) {
		resp, err := testIssueToken(t, backend, req, d)
		require.NoError(t, err)
		require.False(t, resp.IsError())
		assert.NotEmpty(t, resp.Data["id"], "no id returned")

		mustRevoke(t, backend, req.Storage, ID, resp.Data["id"].(int))
	})

	t.Run("non-existing", func(t *testing.T) {
		resp, err := testIssueToken(t, backend, req, d)
		require.NoError(t, err)
		require.False(t, resp.IsError())
		assert.NotEmpty(t, resp.Data["id"], "no id returned")

		mustRevoke(t, backend, req.Storage, ID, resp.Data["id"].(int))

		resp, err = testRevoke(t, backend, req.Storage, ID, resp.Data["id"].(int))
		require.NoError(t, err)
		require.True(t, resp.IsError())
	})

	t.Run("revoke with invalid parameters", func(t *testing.T) {
		resp, err := testRevoke(t, backend, req.Storage, 0, 0)
		require.NoError(t, err)
		require.True(t, resp.IsError())
	})
}
func testRevoke(t *testing.T, b logical.Backend, s logical.Storage, id, tokenID int) (*logical.Response, error) {
	t.Helper()
	data := map[string]interface{}{
		"id":       id,
		"token_id": tokenID,
	}

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      pathPatternRevoke,
		Data:      data,
		Storage:   s,
	})
	return resp, err

}

func mustRevoke(t *testing.T, b logical.Backend, s logical.Storage, id, tokenID int) {
	resp, err := testRevoke(t, b, s, id, tokenID)
	require.NoError(t, err)
	require.Nil(t, resp)
}
