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
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
)

type TokenStorageEntry struct {
	BaseTokenStorage BaseTokenStorageEntry
	ExpiresAt        *time.Time `json:"expires_at" structs:"expires_at" mapstructure:"expires_at,omitempty"`
}

type BaseTokenStorageEntry struct {
	// `json:"" structs:"" mapstructure:""`
	ID     int      `json:"id" structs:"id" mapstructure:"id"`
	Name   string   `json:"name" structs:"name" mapstructure:"name"`
	Scopes []string `json:"scopes" structs:"scopes" mapstructure:"scopes"`
	// AccessLevel int `json:"access_level" structs:"access_level" mapstructure:"access_level,omitempty"`
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

func (BaseTokenStorage *BaseTokenStorageEntry) assertValid() error {
	var err *multierror.Error
	if BaseTokenStorage.ID <= 0 {
		err = multierror.Append(err, errors.New("id is empty or invalid"))
	}
	if BaseTokenStorage.Name == "" {
		err = multierror.Append(err, errors.New("name is empty"))
	}
	if len(BaseTokenStorage.Scopes) == 0 {
		err = multierror.Append(err, errors.New("scopes are empty"))
	} else if e := validateScopes(BaseTokenStorage.Scopes); e != nil {
		err = multierror.Append(err, e)
	}

	return err.ErrorOrNil()
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
	// if accessLevelRaw, ok := data.GetOk("access_level"); ok {
	// 	tokenStorage.AccessLevel = accessLevelRaw.(string)
	// }
}
