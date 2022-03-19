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

// schema for the role, this will map the fields coming in from the
// vault request field map
var roleSchema = map[string]*framework.FieldSchema{
	"role_name": {
		Type:        framework.TypeString,
		Description: "Role name",
	},
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
	"token_ttl": {
		Type:        framework.TypeDurationSecond,
		Description: "The TTL of the token",
		Default:     24 * 3600, // 24 hours, until it hits midnight UTC
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

func roleDetail(role *RoleStorageEntry) map[string]interface{} {
	return map[string]interface{}{
		"role_name":    role.RoleName,
		"id":           role.BaseTokenStorage.ID,
		"name":         role.BaseTokenStorage.Name,
		"scopes":       role.BaseTokenStorage.Scopes,
		"access_level": role.BaseTokenStorage.AccessLevel,
		"token_ttl":    int64(role.TokenTTL / time.Second),
		"token_type":   role.BaseTokenStorage.TokenType,
	}
}

func (b *GitlabBackend) pathRoleCreateUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	warnings := []string{}

	roleName := data.Get("role_name").(string)
	if roleName == "" {
		return logical.ErrorResponse("Role name not supplied"), nil
	}

	lock := b.roleLock(roleName)
	lock.RLock()
	defer lock.RUnlock()

	role, err := getRoleEntry(ctx, req.Storage, roleName)
	if err != nil {
		return logical.ErrorResponse("Error reading role"), nil
	}

	if role == nil {
		role = &RoleStorageEntry{
			RoleName: roleName,
		}
	}
	role.retrieve(data)
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return logical.ErrorResponse("failed to obtain gitlab config - %s", err.Error()), nil
	}
	if config == nil {
		return logical.ErrorResponse("gitlab backend configuration has not been set up"), nil
	}
	err = role.assertValid(config.MaxTTL)
	if err != nil {
		return logical.ErrorResponse("Failed to validate - " + err.Error()), nil
	}
	if role.TokenTTL == 0 {
		warnings = append(warnings, NoTTLWarning("token_ttl"))
	}

	if err := role.save(ctx, req.Storage); err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}
	b.Logger().Debug("successfully create role", "role_name", roleName, "id", role.BaseTokenStorage.ID,
		"name", role.BaseTokenStorage.Name, "scopes", role.BaseTokenStorage.Scopes, "token_type", role.BaseTokenStorage.TokenType)

	return &logical.Response{
		Data:     roleDetail(role),
		Warnings: warnings,
	}, nil
}

func (b *GitlabBackend) pathRoleRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleName := data.Get("role_name").(string)
	role, err := getRoleEntry(ctx, req.Storage, roleName)
	if err != nil {
		return logical.ErrorResponse("Error reading role"), err
	}

	if role == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: roleDetail(role),
	}, nil
}

func (b *GitlabBackend) pathRoleDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleName := data.Get("role_name").(string)
	if roleName == "" {
		return logical.ErrorResponse("Unable to remove, missing role name"), nil
	}

	lock := b.roleLock(roleName)
	lock.RLock()
	defer lock.RUnlock()

	// get the role to make sure it exists and to get the role id
	role, err := getRoleEntry(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, nil
	}

	if err := deleteRoleEntry(ctx, req.Storage, roleName); err != nil {
		return logical.ErrorResponse(fmt.Sprintf("Unable to remove role %s", roleName)), err
	}

	b.Logger().Debug("successfully deleted role", "role_name", roleName)
	return nil, nil
}

func (b *GitlabBackend) pathRoleList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roles, err := listRoleEntries(ctx, req.Storage)
	if err != nil {
		return logical.ErrorResponse("Error listing roles"), err
	}
	return logical.ListResponse(roles), nil
}

// set up the paths for the roles within vault
func pathRole(b *GitlabBackend) []*framework.Path {
	paths := []*framework.Path{
		{
			Pattern: fmt.Sprintf("%s/%s", pathPatternRoles, framework.GenericNameRegex("role_name")),
			Fields:  roleSchema,
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.pathRoleCreateUpdate,
					Summary:  "Create a role",
					Examples: roleExamples,
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathRoleCreateUpdate,
				},
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.pathRoleRead,
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.pathRoleDelete,
				},
			},
			HelpSynopsis:    pathRoleHelpSyn,
			HelpDescription: pathRoleHelpDesc,
		},
	}

	return paths
}

func pathRoleList(b *GitlabBackend) []*framework.Path {
	// Paths for listing role sets
	paths := []*framework.Path{
		{
			Pattern: fmt.Sprintf("%s?/?", pathPatternRoles),
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ListOperation: b.pathRoleList,
			},
			HelpSynopsis: pathListRoleHelpSyn,
		},
	}
	return paths
}

const pathRoleHelpSyn = `Create a role with parameters that are used to generate an access token.`
const pathRoleHelpDesc = `
This path allows you to create a role whose parameters will be used to generate an access token. 
You must supply a project/group id to generate a token for, a name, which will be used as a name field in Gitlab, 
and scopes for the generated project access token.
`

var roleExamples = []framework.RequestExample{
	{
		Description: "Create/update a role",
		Data: map[string]interface{}{
			"role_name": "MyProject1ReadRole",
			"id":        1,
			"name":      "MyProjectAccessToken",
			"scopes":    []string{"read_api", "read_repository"},
		},
	},
}

const pathListRoleHelpSyn = `List existing roles.`
