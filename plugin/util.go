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
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/go-multierror"
)

func validateScopes(scopes []string) error {
	var err *multierror.Error

	for _, scope := range scopes {
		switch scope {
		case "api", "read_api",
			"read_registry", "write_registry",
			"read_repository", "write_repository":
			continue
		default:
			err = multierror.Append(err, fmt.Errorf("scope '%s' is not allowed", scope))
		}
	}
	return err.ErrorOrNil()
}

func envOrDefault(key, d string) string {
	env := os.Getenv(key)
	if env == "" {
		return d
	}
	return env
}

func envAsInt(key string, d int) int {
	v := envOrDefault(key, "")
	if val, err := strconv.Atoi(v); err == nil {
		return val
	}

	return d
}
