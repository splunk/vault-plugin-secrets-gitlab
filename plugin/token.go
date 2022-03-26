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

var errInvalidAccessLevel = errors.New("invalid access level")

const (
	tokenTypeProject string = "project"
	tokenTypeGroup   string = "group"
)

type TokenStorageEntry struct {
	BaseTokenStorage BaseTokenStorageEntry
	ExpiresAt        *time.Time `json:"expires_at" structs:"expires_at" mapstructure:"expires_at,omitempty"`
}

type BaseTokenStorageEntry struct {
	// `json:"" structs:"" mapstructure:""`
	ID          int      `json:"id" structs:"id" mapstructure:"id"`
	Name        string   `json:"name" structs:"name" mapstructure:"name"`
	Scopes      []string `json:"scopes" structs:"scopes" mapstructure:"scopes"`
	AccessLevel int      `json:"access_level" structs:"access_level" mapstructure:"access_level,omitempty"`
	TokenType   string   `json:"token_type" struct:"token_type" mapstructure:"token_type,omitempty"`
}

func (tokenStorage *TokenStorageEntry) assertValid(maxTTL time.Duration) error {
	var err *multierror.Error
	if e := tokenStorage.BaseTokenStorage.assertValid(); e != nil {
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

func (baseTokenStorage *BaseTokenStorageEntry) assertValid() error {
	var err *multierror.Error
	if baseTokenStorage.ID <= 0 {
		err = multierror.Append(err, errors.New("id is empty or invalid"))
	}
	if baseTokenStorage.Name == "" {
		err = multierror.Append(err, errors.New("name is empty"))
	}
	if len(baseTokenStorage.Scopes) == 0 {
		err = multierror.Append(err, errors.New("scopes are empty"))
	} else if e := validateScopes(baseTokenStorage.Scopes); e != nil {
		err = multierror.Append(err, e)
	}

	// check validity of access level. allowed values are 0(zero value), 10, 20, 30 and 40
	if d := baseTokenStorage.AccessLevel / 10; d > 4 || d < 0 {
		err = multierror.Append(err, errInvalidAccessLevel)
	} else if baseTokenStorage.AccessLevel%10 != 0 {
		err = multierror.Append(err, errInvalidAccessLevel)
	}

	// no default type for access token
	if e := validateTokenType(baseTokenStorage.TokenType); err != nil {
		err = multierror.Append(err, e)
	}

	return err.ErrorOrNil()
}

func validateTokenType(t string) error {
	switch t {
	case tokenTypeGroup, tokenTypeProject:
		return nil
	default:
		return fmt.Errorf("token_type must be either %s or %s", tokenTypeProject, tokenTypeGroup)
	}
}

func (tokenStorage *TokenStorageEntry) retrieve(data *framework.FieldData) {
	tokenStorage.BaseTokenStorage.retrieve(data)
	if expiresAtRaw, ok := data.GetOk("expires_at"); ok {
		t := expiresAtRaw.(time.Time)
		tokenStorage.ExpiresAt = &t
	}
}

func (baseTokenStorage *BaseTokenStorageEntry) retrieve(data *framework.FieldData) {
	if idRaw, ok := data.GetOk("id"); ok {
		baseTokenStorage.ID = idRaw.(int)
	}
	if nameRaw, ok := data.GetOk("name"); ok {
		baseTokenStorage.Name = nameRaw.(string)
	}
	if scopesRaw, ok := data.GetOk("scopes"); ok {
		baseTokenStorage.Scopes = scopesRaw.([]string)
	}
	if accessLevelRaw, ok := data.GetOk("access_level"); ok {
		baseTokenStorage.AccessLevel = accessLevelRaw.(int)
	}
	if tokenType, ok := data.GetOk("token_type"); ok {
		baseTokenStorage.TokenType = tokenType.(string)
	}
}

// not right way to do this. use generic introduced in 1.18
func (baseTokenStorage *BaseTokenStorageEntry) createAccessToken(gc Client, expiresAt time.Time) (data map[string]interface{}, err error) {
	switch baseTokenStorage.TokenType {
	case tokenTypeGroup:
		gat, err := gc.CreateGroupAccessToken(baseTokenStorage, &expiresAt)
		if err != nil {
			err = fmt.Errorf("Failed to create a group token - " + err.Error())
		}
		data = groupTokenDetails(gat)
	case tokenTypeProject:
		pat, err := gc.CreateProjectAccessToken(baseTokenStorage, &expiresAt)
		if err != nil {
			err = fmt.Errorf("Failed to create a project token - " + err.Error())
		}
		data = projectTokenDetails(pat)
	}

	return
}
