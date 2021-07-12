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
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// schema for the creation of the role, this will map the fields coming in from the
// vault request field map
var createAccessTokenSchema = map[string]*framework.FieldSchema{
	"id": {
		Type:        framework.TypeInt,
		Description: "Project ID to create a project access token for",
	},
	"name": {
		Type:        framework.TypeString,
		Description: "The name of the project access token",
	},
	"scopes": {
		Type:        framework.TypeCommaStringSlice,
		Description: "List of scopes",
	},
	"expires_at": {
		Type:        framework.TypeTime,
		Description: "The token expires at midnight UTC on that date",
	},
	// Not valid until gitlab 14.1
	// "access_level": {
	// 	Type:        framework.TypeInt,
	// 	Description: "access level of project access token",
	//  Default: accessLevelMaintainer,
	// },
}

func (b *GitlabBackend) pathTokenCreate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	tokenDetails := func(pat *PAT) map[string]interface{} {
		d := map[string]interface{}{
			"token":  pat.Token,
			"id":     pat.ID,
			"name":   pat.Name,
			"scopes": pat.Scopes,
		}
		if pat.ExpiresAt != nil {
			d["expires_at"] = time.Time(*pat.ExpiresAt)
		}
		return d
	}

	gc, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return logical.ErrorResponse("failed to obtain gitlab client - %s", err.Error()), nil
	}

	var tokenStorage TokenStorageEntry

	if idRaw, ok := data.GetOk("id"); ok {
		tokenStorage.ID = idRaw.(int)
	}
	if nameRaw, ok := data.GetOk("name"); ok {
		tokenStorage.Name = nameRaw.(string)
	}
	if scopesRaw, ok := data.GetOk("scopes"); ok {
		tokenStorage.Scopes = scopesRaw.([]string)
	}
	if expiresAtRaw, ok := data.GetOk("expires_at"); ok {
		t := expiresAtRaw.(time.Time)
		tokenStorage.ExpiresAt = &t
	}
	// if accessLevelRaw, ok := data.GetOk("access_level"); ok {
	// 	tokenStorage.AccessLevel = accessLevelRaw.(string)
	// }
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	err = tokenStorage.assertValid(config.MaxTTL)
	if err != nil {
		return logical.ErrorResponse("Failed to validate - " + err.Error()), nil
	}

	b.Logger().Debug("generating access token", "id", tokenStorage.ID, "name", tokenStorage.Name, "scopes", tokenStorage.Scopes)
	pat, err := gc.CreateProjectAccessToken(&tokenStorage)
	if err != nil {
		return logical.ErrorResponse("Failed to create a token - " + err.Error()), nil
	}
	return &logical.Response{Data: tokenDetails(pat)}, nil
}

// set up the paths for the roles within vault
func pathToken(b *GitlabBackend) []*framework.Path {
	paths := []*framework.Path{
		{
			Pattern: pathPatternToken,
			Fields:  createAccessTokenSchema,
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.CreateOperation: &framework.PathOperation{

					Callback: b.pathTokenCreate,
					Summary:  "Create a project access token",
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

const pathTokenHelpSyn = `Generate a project access token for a given project with token name, scopes.`
const pathTokenHelpDesc = `
This path allows you to generate a project access token for a given role. You must supply a name, which 
will be used as a name field in Gitlab, and scopes for the generated project access token.
`

var tokenExamples = []framework.RequestExample{
	{
		Description: "Create/update backend configuration",
		Data: map[string]interface{}{
			"id":     1,
			"name":   "MyProjectAccessToken",
			"scopes": []string{"read_api", "read_repository"},
		},
	},
}
