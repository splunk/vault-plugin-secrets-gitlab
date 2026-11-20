package gitlabtoken

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRolePayloads exercises project and group role payloads against the backend.
//
//nolint:funlen
func TestRolePayloads(t *testing.T) {
	fixture, err := os.ReadFile("testdata/role-payloads.json")
	require.NoError(t, err)

	var resources []struct {
		Path string         `json:"path"`
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(fixture, &resources))

	for _, resource := range resources {
		if !strings.HasPrefix(resource.Path, "roles/") {
			continue
		}

		t.Run(resource.Path, func(t *testing.T) {
			kind, ok := resource.Data["token_type"].(string)
			require.True(t, ok)

			calls := 0

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++

				assert.Equal(t, "/api/v4/"+kind+"s/500/access_tokens", r.URL.Path)

				var payload map[string]any
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
				assert.Equal(t, time.Now().UTC().Add(24*time.Hour).Format(time.DateOnly), payload["expires_at"])
				assert.Equal(t, resource.Data["name"], payload["name"])
				assert.Equal(t, resource.Data["scopes"], payload["scopes"])
				assert.Equal(t, resource.Data["access_level"], payload["access_level"])
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = fmt.Fprintf(w, `{"id":123,"name":%q,"token":"issued-token","scopes":["read_api"],"access_level":20,"expires_at":%q}`, payload["name"], payload["expires_at"])
			}))
			defer server.Close()

			backend, storage := getTestBackend(t, false)
			testConfigUpdate(t, backend, storage, map[string]any{"base_url": server.URL, "token": "backend-token"})

			roleName := strings.TrimPrefix(resource.Path, "roles/")
			mustRoleCreate(t, backend, storage, roleName, resource.Data)
			resp, err := backend.HandleRequest(context.Background(), &logical.Request{Operation: logical.UpdateOperation, Path: "token/" + roleName, Storage: storage})
			require.NoError(t, err)
			require.False(t, resp.IsError())
			assert.Equal(t, "issued-token", resp.Data["token"])
			assert.Equal(t, resource.Data["name"], resp.Data["name"])
			assert.Equal(t, 1, calls)

			if kind == TokenTypeGroup {
				resp, err = backend.HandleRequest(context.Background(), &logical.Request{Operation: logical.UpdateOperation, Path: "roles/" + roleName, Data: map[string]any{"token_type": TokenTypeProject}, Storage: storage})
				require.NoError(t, err)
				require.False(t, resp.IsError())

				kind = TokenTypeProject
				resp, err = backend.HandleRequest(context.Background(), &logical.Request{Operation: logical.UpdateOperation, Path: "token/" + roleName, Storage: storage})
				require.NoError(t, err)
				require.False(t, resp.IsError())
				assert.Equal(t, 2, calls)
			}
		})
	}
}
