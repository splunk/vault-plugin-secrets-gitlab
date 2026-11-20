package gitlabtoken

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccessTokenAPI verifies both kinds use their own API and preserve native expiry.
func TestAccessTokenAPI(t *testing.T) {
	for _, tokenType := range []string{"", TokenTypeProject, TokenTypeGroup} {
		for _, withExpiry := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/expiry=%t", tokenType, withExpiry), func(t *testing.T) {
				kind := tokenType
				if kind == "" {
					kind = TokenTypeProject
				}

				var expiresAt *time.Time

				if withExpiry {
					expiry := time.Now().UTC().Add(48 * time.Hour)
					expiresAt = &expiry
				}

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodPost, r.Method)
					assert.Equal(t, "/api/v4/"+kind+"s/500/access_tokens", r.URL.Path)
					assert.Equal(t, "backend-token", r.Header.Get("Private-Token"))

					var payload map[string]any
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
					assert.Equal(t, "ci-reader", payload["name"])
					assert.EqualValues(t, 20, payload["access_level"])
					assert.Equal(t, []any{"read_api"}, payload["scopes"])

					if withExpiry {
						assert.Equal(t, expiresAt.Format(time.DateOnly), payload["expires_at"])
					} else {
						assert.NotContains(t, payload, "expires_at")
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					_, _ = fmt.Fprint(w, `{"id":123,"name":"ci-reader","token":"issued-token","scopes":["read_api"],"access_level":20,"expires_at":"2030-01-02"}`)
				}))
				defer server.Close()

				client, err := NewClient(&ConfigStorageEntry{BaseURL: server.URL, Token: "backend-token"})
				require.NoError(t, err)
				token, err := client.CreateAccessToken(&BaseTokenStorageEntry{ID: 500, Name: "ci-reader", TokenType: tokenType, Scopes: []string{"read_api"}, AccessLevel: 20}, expiresAt)
				require.NoError(t, err)
				assert.Equal(t, "issued-token", token.Token)
				assert.Equal(t, "2030-01-02", time.Time(*token.ExpiresAt).Format(time.DateOnly))
			})
		}
	}
}

// TestGroupTokenIssuance covers direct and persisted role issuance through Vault.
func TestGroupTokenIssuance(t *testing.T) {
	for _, viaRole := range []bool{false, true} {
		t.Run(fmt.Sprintf("role=%t", viaRole), func(t *testing.T) {
			expiry := time.Now().UTC().Add(48 * time.Hour)
			requests := 0

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++

				assert.Equal(t, "/api/v4/groups/500/access_tokens", r.URL.Path)

				var payload map[string]any
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
				assert.Equal(t, expiry.Format(time.DateOnly), payload["expires_at"])
				assert.EqualValues(t, 20, payload["access_level"])
				assert.Equal(t, []any{"read_api"}, payload["scopes"])
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = fmt.Fprintf(w, `{"id":123,"name":"ci-reader","token":"issued-token","scopes":["read_api"],"access_level":20,"expires_at":%q}`, expiry.Format(time.DateOnly))
			}))
			defer server.Close()

			backend, storage := getTestBackend(t, false)
			testConfigUpdate(t, backend, storage, map[string]any{"base_url": server.URL, "token": "backend-token"})

			data := map[string]any{"id": 500, "name": "ci-reader", "token_type": TokenTypeGroup, "scopes": []string{"read_api"}, "access_level": 20}
			path := "token"

			if viaRole {
				data["token_ttl"] = "48h"
				mustRoleCreate(t, backend, storage, "reader", data)

				path = "token/reader"
				data = nil
			} else {
				data["expires_at"] = expiry.Format(time.RFC3339)
			}

			resp, err := backend.HandleRequest(context.Background(), &logical.Request{Operation: logical.UpdateOperation, Path: path, Data: data, Storage: storage})
			require.NoError(t, err)
			require.False(t, resp.IsError(), "%v", resp.Data)
			assert.Equal(t, "issued-token", resp.Data["token"])
			assert.Equal(t, "ci-reader", resp.Data["name"])
			returnedExpiry, ok := resp.Data["expires_at"].(time.Time)
			require.True(t, ok)
			assert.Equal(t, expiry.Format(time.DateOnly), returnedExpiry.Format(time.DateOnly))
			assert.Equal(t, 1, requests)
		})
	}
}

// TestGroupRoleRoundTrip verifies partial updates and explicit kind changes persist.
func TestGroupRoleRoundTrip(t *testing.T) {
	backend, storage := getTestBackend(t, false)
	testConfigUpdate(t, backend, storage, map[string]any{"base_url": "https://gitlab.example.com", "token": "backend-token"})
	mustRoleCreate(t, backend, storage, "reader", map[string]any{"id": 500, "name": "ci-reader", "token_type": TokenTypeGroup, "scopes": []string{"read_api"}})

	for _, update := range []map[string]any{{"name": "renamed"}, {"token_type": TokenTypeProject}} {
		resp, err := backend.HandleRequest(context.Background(), &logical.Request{Operation: logical.UpdateOperation, Path: "roles/reader", Data: update, Storage: storage})
		require.NoError(t, err)
		require.False(t, resp.IsError())
		read, err := testRoleRead(t, backend, storage, "reader")
		require.NoError(t, err)

		expected := TokenTypeGroup
		if update["token_type"] != nil {
			expected = TokenTypeProject
		}

		assert.Equal(t, expected, read.Data["token_type"])
		assert.Equal(t, "renamed", read.Data["name"])
	}

	legacy := &logical.StorageEntry{Key: "roles/legacy", Value: []byte(`{"role_name":"legacy","BaseTokenStorage":{"id":500,"name":"legacy","scopes":["read_api"]}}`)}
	require.NoError(t, storage.Put(context.Background(), legacy))
	read, err := testRoleRead(t, backend, storage, "legacy")
	require.NoError(t, err)
	assert.Equal(t, TokenTypeProject, read.Data["token_type"])
}

// TestGroupTokenRejectsInvalidRequests ensures validation happens before any GitLab call.
func TestGroupTokenRejectsInvalidRequests(t *testing.T) {
	backend, storage := getTestBackend(t, false)
	testConfigUpdate(t, backend, storage, map[string]any{"base_url": "https://gitlab.invalid", "token": "backend-token", "max_ttl": "48h"})

	for _, path := range []string{"token", "roles/reader"} {
		for _, invalid := range []map[string]any{{"token_type": "personal"}, {"id": -1}, {"access_level": 31}, {"expires_at": time.Now().UTC().Add(96 * time.Hour).Format(time.RFC3339), "token_ttl": "96h"}} {
			data := map[string]any{"id": 500, "name": "ci-reader", "token_type": TokenTypeGroup, "scopes": []string{"read_api"}}
			maps.Copy(data, invalid)
			// Only pass fields supported by the endpoint being validated.
			if path == "token" {
				delete(data, "token_ttl")
			} else {
				delete(data, "expires_at")
			}

			resp, err := backend.HandleRequest(context.Background(), &logical.Request{Operation: logical.UpdateOperation, Path: path, Data: data, Storage: storage})
			require.NoError(t, err)
			require.True(t, resp.IsError(), "%s: %v", path, resp.Data)
		}
	}
}

// TestGroupTokenAPIErrors ensures forbidden requests never produce credentials.
func TestGroupTokenAPIErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = fmt.Fprint(w, `{"message":"403 Forbidden"}`)
	}))
	defer server.Close()

	backend, storage := getTestBackend(t, false)
	testConfigUpdate(t, backend, storage, map[string]any{"base_url": server.URL, "token": "backend-token"})

	data := map[string]any{"id": 500, "name": "ci-reader", "token_type": TokenTypeGroup, "scopes": []string{"read_api"}}
	mustRoleCreate(t, backend, storage, "reader", data)

	for _, path := range []string{"token", "token/reader"} {
		resp, err := backend.HandleRequest(context.Background(), &logical.Request{Operation: logical.UpdateOperation, Path: path, Data: data, Storage: storage})
		require.NoError(t, err)
		require.True(t, resp.IsError())
		assert.NotContains(t, resp.Data, "token")
		assert.Contains(t, resp.Data["error"], "403")
	}
}

// TestStoredUnknownTokenKindCannotFallBack ensures corrupted roles cannot issue project tokens.
func TestStoredUnknownTokenKindCannotFallBack(t *testing.T) {
	backend, storage := getTestBackend(t, false)
	testConfigUpdate(t, backend, storage, map[string]any{"base_url": "https://gitlab.invalid", "token": "backend-token"})

	role := &RoleStorageEntry{RoleName: "invalid", TokenTTL: 24 * time.Hour, BaseTokenStorage: BaseTokenStorageEntry{ID: 500, Name: "invalid", Scopes: []string{"read_api"}, TokenType: "personal"}}
	require.NoError(t, role.save(context.Background(), storage))
	resp, err := backend.HandleRequest(context.Background(), &logical.Request{Operation: logical.UpdateOperation, Path: "token/invalid", Storage: storage})
	require.NoError(t, err)
	require.True(t, resp.IsError())
	assert.Contains(t, resp.Data["error"], `invalid token_type "personal"`)
}
