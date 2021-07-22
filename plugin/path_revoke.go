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

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// schema for the token, this will map the fields coming in from the
// vault request field map
var revokeSchema = map[string]*framework.FieldSchema{
	"id": {
		Type:        framework.TypeInt,
		Description: "Project ID to revoke a project access token for",
	},
	"token_id": {
		Type:        framework.TypeInt,
		Description: "The token id of the project access token",
	},
}

func (b *GitlabBackend) pathRevoke(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	gc, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return logical.ErrorResponse("failed to obtain gitlab client - %s", err.Error()), nil
	}

	var revokeStorage RevokeStorageEntry
	revokeStorage.retrieve(data)
	err = revokeStorage.assertValid()
	if err != nil {
		return logical.ErrorResponse("Failed to validate - " + err.Error()), nil
	}

	b.Logger().Debug("revoking access token", "id", revokeStorage.ID, "token_id", revokeStorage.TokenID)
	err = gc.RevokeProjectAccessToken(&revokeStorage)
	if err != nil {
		return logical.ErrorResponse("Failed to revoke a token - " + err.Error()), nil
	}
	return nil, nil
}

// set up the paths for the roles within vault
func pathRevoke(b *GitlabBackend) []*framework.Path {
	paths := []*framework.Path{
		{
			Pattern: pathPatternRevoke,
			Fields:  revokeSchema,
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.pathRevoke,
					Summary:  "Revoke a project access token",
					Examples: revokeExamples,
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.pathRevoke,
				},
			},
			HelpSynopsis:    pathRevokeHelpSyn,
			HelpDescription: pathRevokeHelpDesc,
		},
	}

	return paths
}

const pathRevokeHelpSyn = `Revoke a project access token for a given project with token id.`
const pathRevokeHelpDesc = `
This path allows you to revoke a project access token. You must supply a project id to revoke a token for, a token id.
`

var revokeExamples = []framework.RequestExample{
	{
		Description: "Create a project access token",
		Data: map[string]interface{}{
			"id":     1,
			"name":   "MyProjectAccessToken",
			"scopes": []string{"read_api", "read_repository"},
		},
	},
}
