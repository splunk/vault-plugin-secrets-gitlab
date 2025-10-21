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
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"
)

// RoleStorageEntry represents the role metadata stored in vault.
type RoleStorageEntry struct {
	// `json:"" structs:"" mapstructure:""`
	RoleName string `json:"role_name" structs:"role_name" mapstructure:"role_name"`
	// The TTL for your token
	TokenTTL         time.Duration `json:"token_ttl" structs:"token_ttl" mapstructure:"token_ttl"`
	BaseTokenStorage BaseTokenStorageEntry
}

func (role *RoleStorageEntry) assertValid(maxTTL time.Duration, allowOwnerLevel bool) error {
	var err *multierror.Error

	e := role.BaseTokenStorage.assertValid(allowOwnerLevel)
	if e != nil {
		err = multierror.Append(err, e)
	}

	if maxTTL > time.Duration(0) {
		if role.TokenTTL > maxTTL {
			errMsg := fmt.Sprintf("Requested token ttl '%v' exceeds configured maximum ttl of '%v's. ",
				role.TokenTTL, int64(maxTTL/time.Second))
			err = multierror.Append(err, errors.New(errMsg))
		}
	}

	return err.ErrorOrNil()
}

func (role *RoleStorageEntry) retrieve(data *framework.FieldData) {
	role.BaseTokenStorage.retrieve(data)

	ttlRaw, ok := data.GetOk("token_ttl")

	ttl, _ := ttlRaw.(int)
	if ok && ttl > 0 {
		role.TokenTTL = time.Duration(ttl) * time.Second
	} else if role.TokenTTL == time.Duration(0) {
		tokenTTL, _ := roleSchema["token_ttl"].Default.(int)
		role.TokenTTL = time.Duration(tokenTTL) * time.Second
	}
}

// save saves a role to storage.
func (role *RoleStorageEntry) save(ctx context.Context, storage logical.Storage) error {
	entry, err := logical.StorageEntryJSON(fmt.Sprintf("%s/%s", pathPatternRoles, role.RoleName), role)
	if err != nil {
		return err
	}

	return storage.Put(ctx, entry)
}

// Get or create the basic lock for the role name.
func (b *GitlabBackend) roleLock(roleName string) *locksutil.LockEntry {
	return locksutil.LockForKey(b.roleLocks, roleName)
}

// deleteRoleEntry will remove the role with specified name from storage.
func deleteRoleEntry(ctx context.Context, storage logical.Storage, roleName string) error {
	if roleName == "" {
		return errors.New("missing role name")
	}

	return storage.Delete(ctx, fmt.Sprintf("%s/%s", pathPatternRoles, roleName))
}

// getRoleEntry fetches a role from the storage.
func getRoleEntry(ctx context.Context, storage logical.Storage, roleName string) (*RoleStorageEntry, error) {
	var result RoleStorageEntry

	entry, err := storage.Get(ctx, fmt.Sprintf("%s/%s", pathPatternRoles, roleName))
	if err != nil {
		return nil, err
	} else if entry == nil {
		return nil, nil
	}

	err = entry.DecodeJSON(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// listRoleEntries gets all the roles.
func listRoleEntries(ctx context.Context, storage logical.Storage) ([]string, error) {
	roles, err := storage.List(ctx, pathPatternRoles+"/")
	if err != nil {
		return nil, err
	}

	return roles, nil
}
