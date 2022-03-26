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
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// schema for the token, this will map the fields coming in from the
// vault request field map
var accessTokenSchema = map[string]*framework.FieldSchema{
	"id": {
		Type:        framework.TypeInt,
		Description: "Project/Group ID to create an access token for",
	},
	"name": {
		Type:        framework.TypeString,
		Description: "The name of the access token",
	},
	"scopes": {
		Type:        framework.TypeCommaStringSlice,
		Description: "List of scopes",
	},
	"expires_at": {
		Type:        framework.TypeTime,
		Description: "The token expires at midnight UTC on that date",
	},
	"access_level": {
		Type:        framework.TypeInt,
		Description: "access level of access token",
	},
	"token_type": {
		Type:        framework.TypeString,
		Description: "access token type",
	},
}

func projectTokenDetails(pat *PAT) map[string]interface{} {
	d := map[string]interface{}{
		"token":        pat.Token,
		"id":           pat.ID,
		"name":         pat.Name,
		"scopes":       pat.Scopes,
		"access_level": pat.AccessLevel,
	}
	if pat.ExpiresAt != nil {
		d["expires_at"] = time.Time(*pat.ExpiresAt)
	}
	return d
}
func groupTokenDetails(gat *GAT) map[string]interface{} {
	d := map[string]interface{}{
		"token":        gat.Token,
		"id":           gat.ID,
		"name":         gat.Name,
		"scopes":       gat.Scopes,
		"access_level": gat.AccessLevel,
	}
	if gat.ExpiresAt != nil {
		d["expires_at"] = time.Time(*gat.ExpiresAt)
	}
	return d
}

func (b *GitlabBackend) pathTokenCreate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	gc, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return logical.ErrorResponse("failed to obtain gitlab client - %s", err.Error()), nil
	}

	var tokenStorage TokenStorageEntry
	tokenStorage.retrieve(data)

	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return logical.ErrorResponse("failed to obtain gitlab config - %s", err.Error()), nil
	}
	if config == nil {
		return logical.ErrorResponse("gitlab backend configuration has not been set up"), nil
	}
	err = tokenStorage.assertValid(config.MaxTTL)
	if err != nil {
		return logical.ErrorResponse("Failed to validate - " + err.Error()), nil
	}

	b.Logger().Debug("generating access token", "id", tokenStorage.BaseTokenStorage.ID,
		"name", tokenStorage.BaseTokenStorage.Name, "scopes", tokenStorage.BaseTokenStorage.Scopes)

	d, err := tokenStorage.BaseTokenStorage.createAccessToken(gc, *tokenStorage.ExpiresAt)
	if err != nil {
		return logical.ErrorResponse("Failed to create a token - " + err.Error()), nil
	}
	return &logical.Response{Data: d}, nil
}

// set up the paths for the roles within vault
func pathToken(b *GitlabBackend) []*framework.Path {
	paths := []*framework.Path{
		{
			Pattern: pathPatternToken,
			Fields:  accessTokenSchema,
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.CreateOperation: &framework.PathOperation{

					Callback: b.pathTokenCreate,
					Summary:  "Create an access token",
					Examples: tokenExamples,
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathTokenCreate,
				},
			},
			HelpSynopsis:    pathTokenHelpSyn,
			HelpDescription: pathTokenHelpDesc,
		},
	}

	return paths
}

const pathTokenHelpSyn = `Generate an access token for a given project/group with token name, scopes.`
const pathTokenHelpDesc = `
This path allows you to generate an access token. You must supply a project/group id to generate a token for, a name, which 
will be used as a name field in Gitlab, and scopes for the generated project access token.
`

var tokenExamples = []framework.RequestExample{
	{
		Description: "Create an access token",
		Data: map[string]interface{}{
			"id":     1,
			"name":   "MyAccessToken",
			"scopes": []string{"read_api", "read_repository"},
		},
	},
}
