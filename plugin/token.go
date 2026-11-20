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
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
)

var errAccessLevelInvalid = errors.New("invalid access level")
var errAccessLevelNotPermitted = errors.New("access level not permitted")

// Supported token target kinds; omitted kinds retain the project default.
const (
	TokenTypeProject = "project"
	TokenTypeGroup   = "group"
	tokenTypeField   = "token_type"
)

// TokenStorageEntry represents the token metadata stored in vault.
type TokenStorageEntry struct {
	BaseTokenStorage BaseTokenStorageEntry
	ExpiresAt        *time.Time `json:"expires_at" structs:"expires_at" mapstructure:"expires_at,omitempty"`
}

// BaseTokenStorageEntry represents the base token metadata stored in vault.
type BaseTokenStorageEntry struct {
	// `json:"" structs:"" mapstructure:""`
	ID          int      `json:"id" structs:"id" mapstructure:"id"`
	Name        string   `json:"name" structs:"name" mapstructure:"name"`
	Scopes      []string `json:"scopes" structs:"scopes" mapstructure:"scopes"`
	AccessLevel int      `json:"access_level" structs:"access_level" mapstructure:"access_level,omitempty"`
	TokenType   string   `json:"token_type,omitempty" structs:"token_type" mapstructure:"token_type,omitempty"`
}

// tokenType preserves project behavior for roles stored before target kinds existed.
func (baseTokenStorage *BaseTokenStorageEntry) tokenType() string {
	if baseTokenStorage.TokenType == "" {
		return TokenTypeProject
	}

	return baseTokenStorage.TokenType
}

func (tokenStorage *TokenStorageEntry) assertValid(maxTTL time.Duration, allowOwnerLevel bool) error {
	var err *multierror.Error

	e := tokenStorage.BaseTokenStorage.assertValid(allowOwnerLevel)
	if e != nil {
		err = multierror.Append(err, e)
	}

	if maxTTL > time.Duration(0) && tokenStorage.ExpiresAt != nil {
		maxExpiresAt := time.Now().UTC().Add(maxTTL)
		if maxExpiresAt.Before(*tokenStorage.ExpiresAt) {
			errMsg := fmt.Sprintf("Requested expires_at '%v' exceeds configured maximum ttl of '%v's. Expires at or before '%v'",
				*tokenStorage.ExpiresAt, int64(maxTTL/time.Second), maxExpiresAt)
			err = multierror.Append(err, errors.New(errMsg))
		}
	}

	return err.ErrorOrNil()
}

func (baseTokenStorage *BaseTokenStorageEntry) assertValid(allowOwnerLevel bool) error {
	var err *multierror.Error
	if kind := baseTokenStorage.tokenType(); kind != TokenTypeProject && kind != TokenTypeGroup {
		err = multierror.Append(err, fmt.Errorf("invalid token_type %q: expected project or group", kind))
	}

	if baseTokenStorage.ID <= 0 {
		err = multierror.Append(err, errors.New("id is empty or invalid"))
	}

	if baseTokenStorage.Name == "" {
		err = multierror.Append(err, errors.New("name is empty"))
	}

	if len(baseTokenStorage.Scopes) == 0 {
		err = multierror.Append(err, errors.New("scopes are empty"))
	} else {
		e := validateScopes(baseTokenStorage.Scopes)
		if e != nil {
			err = multierror.Append(err, e)
		}
	}

	// check validity of access level. allowed values are:
	// 0(zero value), 10, 20, 30, 40 and 50 (when allowOwnerLevel flag is set)
	maxAllowedAccessLevel := 40
	if allowOwnerLevel {
		maxAllowedAccessLevel = 50
	}

	requestedAccessLevel := baseTokenStorage.AccessLevel

	if requestedAccessLevel%10 != 0 {
		err = multierror.Append(err, errAccessLevelInvalid)
	} else if requestedAccessLevel > maxAllowedAccessLevel || requestedAccessLevel < 0 {
		err = multierror.Append(err, errAccessLevelNotPermitted)
	}

	return err.ErrorOrNil()
}

func (tokenStorage *TokenStorageEntry) retrieve(data *framework.FieldData) {
	tokenStorage.BaseTokenStorage.retrieve(data)

	if expiresAtRaw, ok := data.GetOk("expires_at"); ok {
		t, _ := expiresAtRaw.(time.Time)
		tokenStorage.ExpiresAt = &t
	}
}

func (baseTokenStorage *BaseTokenStorageEntry) retrieve(data *framework.FieldData) {
	if kind, ok := data.GetOk(tokenTypeField); ok {
		baseTokenStorage.TokenType, _ = kind.(string)
	}

	if idRaw, ok := data.GetOk("id"); ok {
		baseTokenStorage.ID, _ = idRaw.(int)
	}

	if nameRaw, ok := data.GetOk("name"); ok {
		baseTokenStorage.Name, _ = nameRaw.(string)
	}

	if scopesRaw, ok := data.GetOk("scopes"); ok {
		baseTokenStorage.Scopes, _ = scopesRaw.([]string)
	}

	if accessLevelRaw, ok := data.GetOk("access_level"); ok {
		baseTokenStorage.AccessLevel, _ = accessLevelRaw.(int)
	}
}
