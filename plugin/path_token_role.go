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
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

var roleTokenSchema = map[string]*framework.FieldSchema{
	"role_name": {
		Type:        framework.TypeString,
		Description: "Role name",
	},
}

func (b *GitlabBackend) pathRoleTokenCreate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	gc, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return logical.ErrorResponse("failed to obtain gitlab client - %s", err.Error()), nil
	}

	roleName := data.Get("role_name").(string)
	// get the role by name
	role, err := getRoleEntry(ctx, req.Storage, roleName)
	if role == nil || err != nil {
		return logical.ErrorResponse(fmt.Sprintf("Role name '%s' not recognised", roleName)), nil
	}

	expiresAt := time.Now().UTC().Add(role.TokenTTL)
	b.Logger().Debug("generating access token for a role", "role_name", role.RoleName, "expires_at", expiresAt)
	d, err := role.BaseTokenStorage.createAccessToken(gc, expiresAt)
	if err != nil {
		return logical.ErrorResponse("Failed to create a token - " + err.Error()), nil
	}
	return &logical.Response{Data: d}, nil
}

// set up the paths for the roles within vault
func pathRoleToken(b *GitlabBackend) []*framework.Path {
	paths := []*framework.Path{
		{
			Pattern: fmt.Sprintf("%s/%s", pathPatternToken, framework.GenericNameRegex("role_name")),
			Fields:  roleTokenSchema,
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.CreateOperation: &framework.PathOperation{

					Callback: b.pathRoleTokenCreate,
					Summary:  "Create an access token based on a predefined role",
					Examples: roleTokenExamples,
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathRoleTokenCreate,
				},
			},
			HelpSynopsis:    pathRoleTokenHelpSyn,
			HelpDescription: pathRoleTokenHelpDesc,
		},
	}

	return paths
}

const pathRoleTokenHelpSyn = `Generate an access token for a given project/group based on a predefined role`
const pathRoleTokenHelpDesc = `
This path allows you to generate an access token based on a predefined role. You must create a role beforehand in /roles/ path,
whose parameters are used to generate an access token.
`

var roleTokenExamples = []framework.RequestExample{
	{
		Description: "Create an access token based on a predefined role",
		Data: map[string]interface{}{
			"role_name": "MyRole",
		},
	},
}
