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
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// schema for the configuring Gitlab token plugin, this will map the fields coming in from the
// vault request field map
var configSchema = map[string]*framework.FieldSchema{
	"base_url": {
		Type:        framework.TypeString,
		Description: `gitlab base url`,
		Default:     "https://gitlab.com",
	},
	"token": {
		Type:        framework.TypeString,
		Description: `gitlab token that has permissions to generate project access tokens`,
	},
	// "max_ttl": {
	// 	Type:        framework.TypeDurationSecond,
	// 	Description: "Maximum time a token generated will be valid for. If <= 0, will use system default(3600).",
	// 	Default:     3600,
	// },
}

func (backend *GitlabBackend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	cfg, err := backend.getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, nil
	}

	return &logical.Response{

		Data: map[string]interface{}{
			"base_url": cfg.BaseURL,
		},
	}, nil
}

func (backend *GitlabBackend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	cfg, err := backend.getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &ConfigStorageEntry{}
	}

	baseURL, ok := data.GetOk("base_url")
	if ok {
		cfg.BaseURL = baseURL.(string)
	} else if cfg.BaseURL == "" {
		cfg.BaseURL = configSchema["base_url"].Default.(string)
	}

	if token, ok := data.GetOk("token"); ok {
		cfg.Token = token.(string)
	}

	// maxTTLRaw, ok := data.GetOk("max_ttl")
	// if ok && maxTTLRaw.(int) > 0 {
	// 	cfg.MaxTTL = time.Duration(maxTTLRaw.(int)) * time.Second
	// } else if cfg.MaxTTL == time.Duration(0) {
	// 	cfg.MaxTTL = time.Duration(configSchema["max_ttl"].Default.(int)) * time.Second
	// }

	entry, err := logical.StorageEntryJSON(pathPatternConfig, cfg)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}

func pathConfig(b *GitlabBackend) []*framework.Path {
	paths := []*framework.Path{
		{
			Pattern: pathPatternConfig,
			Fields:  configSchema,

			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.pathConfigRead,
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathConfigWrite,
					Examples: configExamples,
				},
			},

			HelpSynopsis:    pathConfigHelpSyn,
			HelpDescription: pathConfigHelpDesc,
		},
	}

	return paths
}

const pathConfigHelpSyn = `
Configure the Gitlab backend.
`

const pathConfigHelpDesc = `
The Gitlab backend requires credentials for creating a project access token.
This endpoint is used to configure those credentials as well as default values
for the backend in general.
`

var configExamples = []framework.RequestExample{
	{
		Description: "Create/update backend configuration",
		Data: map[string]interface{}{
			"base_url": "https://my.gitlab.com",
			"token":    "MyPersonalAccessToken",
		},
	},
}
