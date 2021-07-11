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
)

type TokenStorageEntry struct {
	// `json:"" structs:"" mapstructure:""`
	ID        int        `json:"id" structs:"id" mapstructure:"id"`
	Name      string     `json:"name" structs:"name" mapstructure:"name"`
	Scopes    []string   `json:"scopes" structs:"scopes" mapstructure:"scopes"`
	ExpiresAt *time.Time `json:"expires_at" structs:"expires_at" mapstructure:"expires_at,omitempty"`
	// AccessLevel int `json:"access_level" structs:"access_level" mapstructure:"access_level,omitempty"`
}

func (tokenStorage TokenStorageEntry) assertValid(maxTokenLifetime time.Duration) error {
	var err *multierror.Error
	if tokenStorage.ID <= 0 {
		err = multierror.Append(err, errors.New("id is empty or invalid"))
	}
	if tokenStorage.Name == "" {
		err = multierror.Append(err, errors.New("name is empty"))
	}
	if len(tokenStorage.Scopes) == 0 {
		err = multierror.Append(err, errors.New("scopes are empty"))
	} else if e := validateScopes(tokenStorage.Scopes); e != nil {
		err = multierror.Append(err, e)
	}

	if maxTokenLifetime > time.Duration(0) {
		maxExpiresAt := time.Now().UTC().Add(maxTokenLifetime)
		if maxExpiresAt.Before(*tokenStorage.ExpiresAt) {
			errMsg := fmt.Sprintf("Requested expires_at '%v' exceeds configured maximum token lifetime of '%v'. Expires at or before '%v'",
				*tokenStorage.ExpiresAt, maxTokenLifetime, maxExpiresAt)
			err = multierror.Append(err, errors.New(errMsg))
		}
	}

	return err.ErrorOrNil()
}