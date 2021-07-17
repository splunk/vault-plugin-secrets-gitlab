#!/bin/bash

set -euox pipefail


: ${GITLAB_URL:?unset}

export VAULT_ADDR="http://localhost:8200"
export VAULT_TOKEN=root

setup_vault() {
  plugin=vault-plugin-secrets-gitlab
  existing=$(vault secrets list -format json | jq -r '."gitlab/"')
  if [ "$existing" == "null" ]; then

    # in CI, current container bind mount is private, preventing nested bind mounts
    # instead, copy plugin in to vault container and reload
    vault plugin list secret | grep -q gitlab
    if [ $? -ne 0 ]; then
      echo "Plugin missing from dev plugin dir /vault/plugins... registering manually."
      sha=$(shasum -a 256 plugins/$plugin | cut -d' ' -f1)
      # if plugin is missing, it is assumed this is a CI environment and vault is running in a container
      docker cp plugins/$plugin vault:/vault/plugins
      vault plugin register -sha256=$sha secret $plugin
    fi

    echo "Enabling vault gitlab plugin..."
    vault secrets enable -path=gitlab $plugin

  else
    echo
    echo  "Plugin enabled on path 'gitlab/':"
    echo "$existing" | jq
  fi

  vault write gitlab/config base_url=$GITLAB_URL token=$GITLAB_TOKEN
}

setup_vault >&2

# eval output for local use
echo export VAULT_ADDR=\"$VAULT_ADDR\"\;
echo export VAULT_TOKEN=\"$VAULT_TOKEN\"\;
