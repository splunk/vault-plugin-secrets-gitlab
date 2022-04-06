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

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathRole(t *testing.T) {
	a := assert.New(t)
	backend, storage := getTestBackend(t, false)

	conf := map[string]interface{}{
		"base_url": "http://randomhost",
		"token":    "gibberish",
	}
	testConfigUpdate(t, backend, storage, conf)

	data := map[string]interface{}{
		"id":           1,
		"name":         "role-test",
		"scopes":       []string{"api", "read_repository"},
		"access_level": 30,
		"token_type":   "project",
	}
	t.Run("successful", func(t *testing.T) {
		roleName := "successful"
		resp, err := testRoleRead(t, backend, storage, roleName)
		require.NoError(t, err, "non-existing role should not return error")
		require.Nil(t, resp, "non-existing role should return nil response")

		mustRoleCreate(t, backend, storage, roleName, data)

		resp, err = testRoleRead(t, backend, storage, roleName)
		require.NoError(t, err, "existing role should not return error")
		require.False(t, resp.IsError())

		a.Equal(roleName, resp.Data["role_name"])
		a.Equal("role-test", resp.Data["name"])
		a.Equal(1, resp.Data["id"])
		a.Equal([]string{"api", "read_repository"}, resp.Data["scopes"])
		a.Equal(30, resp.Data["access_level"])

		mustRoleDelete(t, backend, storage, roleName)
	})

	t.Run("successful with ttl", func(t *testing.T) {
		conf["max_ttl"] = fmt.Sprintf("%dh", 7*24)
		testConfigUpdate(t, backend, storage, conf)

		roleName := "successful-ttl"
		data["token_ttl"] = 3 * 24 * 3600
		mustRoleCreate(t, backend, storage, roleName, data)

		resp, err := testRoleRead(t, backend, storage, roleName)
		require.NoError(t, err, "existing role should not return error")
		require.False(t, resp.IsError())

		a.Equal(int64(3*24*3600), resp.Data["token_ttl"])

		mustRoleDelete(t, backend, storage, roleName)
	})

	t.Run("delete non-existing", func(t *testing.T) {
		roleName := "non-existing"
		resp, err := testRoleDelete(t, backend, storage, roleName)
		require.NoError(t, err, "non-existing role should not return error")
		require.Nil(t, resp)
	})

	t.Run("validation failure", func(t *testing.T) {
		roleName := "validation-failure"
		d := map[string]interface{}{
			"id":           -1,
			"token_ttl":    fmt.Sprintf("%dh", 30*24),
			"access_level": 31,
			"token_type":   "foo",
		}
		resp, err := testRoleCreate(t, backend, storage, roleName, d)
		require.NoError(t, err)
		require.True(t, resp.IsError())

		require.Contains(t, resp.Data["error"], "id is empty or invalid")
		require.Contains(t, resp.Data["error"], "name is empty")
		require.Contains(t, resp.Data["error"], "scopes are empty")
		require.Contains(t, resp.Data["error"], "exceeds configured maximum ttl")
		require.Contains(t, resp.Data["error"], "invalid access level")
		require.Contains(t, resp.Data["error"], "token_type must be either")
	})
}

func TestPathRoleList(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	backend, storage := getTestBackend(t, false)
	conf := map[string]interface{}{
		"base_url": "http://randomhost",
		"token":    "gibberish",
	}
	testConfigUpdate(t, backend, storage, conf)
	data := map[string]interface{}{
		"id":         1,
		"name":       "role-test",
		"scopes":     []string{"api", "read_repository"},
		"token_type": "project",
	}

	var listResp map[string]interface{}

	resp, err := testRoleList(t, backend, storage)
	require.NoError(t, err)
	require.False(t, resp.IsError())

	err = mapstructure.Decode(resp.Data, &listResp)
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Empty(t, resp.Data, "no role to list should return nil data")

	roleName1 := "test_list_role1"
	roleName2 := "test_list_role2"
	mustRoleCreate(t, backend, storage, roleName1, data)
	mustRoleCreate(t, backend, storage, roleName2, data)

	resp, err = testRoleList(t, backend, storage)
	require.NoError(t, err)
	require.False(t, resp.IsError())
	err = mapstructure.Decode(resp.Data, &listResp)
	require.NoError(t, err)
	returnedRoles := listResp["keys"].([]string)
	a.Len(returnedRoles, 2, "incorrect number of roles")
	a.Equal(roleName1, returnedRoles[0], "incorrect path set")
	a.Equal(roleName2, returnedRoles[1], "incorrect path set")

	mustRoleDelete(t, backend, storage, roleName2)
	resp, err = testRoleList(t, backend, storage)
	require.NoError(t, err)
	require.False(t, resp.IsError())
	err = mapstructure.Decode(resp.Data, &listResp)
	require.NoError(t, err)
	returnedRoles = listResp["keys"].([]string)
	a.Len(returnedRoles, 1, "incorrect number of roles")
	a.Equal(roleName1, returnedRoles[0], "incorrect path set")
}

func testRoleCreate(t *testing.T, b logical.Backend, s logical.Storage, roleName string, data map[string]interface{}) (*logical.Response, error) {
	t.Helper()
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.CreateOperation,
		Path:      fmt.Sprintf("%s/%s", pathPatternRoles, roleName),
		Data:      data,
		Storage:   s,
	})
	return resp, err
}

func mustRoleCreate(t *testing.T, b logical.Backend, s logical.Storage, roleName string, data map[string]interface{}) {
	t.Helper()
	resp, err := testRoleCreate(t, b, s, roleName, data)
	require.NoError(t, err)
	require.False(t, resp.IsError())
}

func testRoleRead(t *testing.T, b logical.Backend, s logical.Storage, roleName string) (*logical.Response, error) {
	t.Helper()
	data := map[string]interface{}{
		"role_name": roleName,
	}

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      fmt.Sprintf("%s/%s", pathPatternRoles, roleName),
		Data:      data,
		Storage:   s,
	})
	return resp, err
}

func testRoleDelete(t *testing.T, b logical.Backend, s logical.Storage, roleName string) (*logical.Response, error) {
	t.Helper()
	data := map[string]interface{}{
		"role_name": roleName,
	}

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      fmt.Sprintf("%s/%s", pathPatternRoles, roleName),
		Data:      data,
		Storage:   s,
	})
	return resp, err
}

func mustRoleDelete(t *testing.T, b logical.Backend, s logical.Storage, roleName string) {
	resp, err := testRoleDelete(t, b, s, roleName)
	require.NoError(t, err)
	require.Nil(t, resp)
}

func testRoleList(t *testing.T, b logical.Backend, s logical.Storage) (*logical.Response, error) {
	t.Helper()
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ListOperation,
		Path:      pathPatternRoles,
		Storage:   s,
	})
	return resp, err
}
