// Copyright  2021 Splunk Inc.
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
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateScopes(t *testing.T) {
	t.Parallel()

	t.Run("valid_scopes", func(t *testing.T) {
		t.Parallel()

		validScopes := []string{"api", "read_api",
			"read_registry", "write_registry",
			"read_repository", "write_repository"}
		err := validateScopes(validScopes)
		require.NoError(t, err, "not expecting error: %s", err)
	})

	t.Run("invalid_scopes", func(t *testing.T) {
		t.Parallel()

		invalidScopes := []string{"something", "invalid"}
		err := validateScopes(invalidScopes)
		require.Error(t, err, "expecting error")

		if merr, ok := err.(*multierror.Error); ok {
			assert.Len(t, merr.Errors, 2, "expecting %d errors, got %s", 2, len(merr.Errors))
		}
		assert.Contains(t, err.Error(), "scope 'something' is not allowed")
	})
}
