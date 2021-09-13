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
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	t.Run("successful", func(t *testing.T) {
		t.Parallel()

		backend, reqStorage := getTestBackend(t, true)

		testConfigRead(t, backend, reqStorage, nil)

		conf := map[string]interface{}{
			"base_url": "https://my.gitlab.com",
			"token":    "mytoken",
		}

		testConfigUpdate(t, backend, reqStorage, conf, NoTTLWarning("max_ttl"))

		expected := map[string]interface{}{
			"base_url": "https://my.gitlab.com",
			"max_ttl":  int64(0),
		}

		testConfigRead(t, backend, reqStorage, expected)

		conf["base_url"] = "https://another.gitlab.com"
		testConfigUpdate(t, backend, reqStorage, conf)

		expected["base_url"] = "https://another.gitlab.com"
		testConfigRead(t, backend, reqStorage, expected)
	})

	t.Run("max ttl", func(t *testing.T) {
		t.Parallel()

		backend, reqStorage := getTestBackend(t, true)

		testConfigRead(t, backend, reqStorage, nil)

		conf := map[string]interface{}{
			"base_url": "https://my.gitlab.com",
			"token":    "mytoken",
			"max_ttl":  fmt.Sprintf("%dh", 30*24),
		}

		testConfigUpdate(t, backend, reqStorage, conf)

		expected := map[string]interface{}{
			"base_url": "https://my.gitlab.com",
			"max_ttl":  int64(30 * 24 * 3600),
		}

		testConfigRead(t, backend, reqStorage, expected)

		// Try seconds
		conf["max_ttl"] = fmt.Sprintf("%ds", 7*24*3600)
		testConfigUpdate(t, backend, reqStorage, conf)

		expected["max_ttl"] = int64(7 * 24 * 3600)
		testConfigRead(t, backend, reqStorage, expected)

		// Try less than 24 hours
		conf["max_ttl"] = fmt.Sprintf("%ds", 12*3600)
		testConfigUpdate(t, backend, reqStorage, conf, LT24HourTTLWarning("max_ttl"))

		testConfigRead(t, backend, reqStorage, expected)
	})
}

func testConfigUpdate(t *testing.T, b logical.Backend, s logical.Storage, d map[string]interface{}, warnings ...string) {
	t.Helper()
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      pathPatternConfig,
		Data:      d,
		Storage:   s,
	})
	require.NoError(t, err)
	require.False(t, resp.IsError())

	for _, warning := range warnings {
		require.Contains(t, resp.Warnings, warning, "it should expect a warning",
			"expected_warning", warning, "actual_warnings", resp.Warnings)
	}
}

func testConfigRead(t *testing.T, b logical.Backend, s logical.Storage, expected map[string]interface{}) {
	t.Helper()
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      pathPatternConfig,
		Storage:   s,
	})

	require.NoError(t, err)

	if resp == nil && expected == nil {
		return
	}

	require.False(t, resp.IsError())
	assert.Equal(t, len(expected), len(resp.Data), "read data mismatch")
	assert.Equal(t, expected, resp.Data, "expected %v, actual: %v", expected, resp.Data)

	if t.Failed() {
		t.FailNow()
	}
}
