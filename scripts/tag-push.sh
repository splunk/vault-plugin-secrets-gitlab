#!/bin/bash 
set -euo pipefail

# should use ssh that's associated to deploy keys instead of user's PAT
URL=`git remote get-url origin | sed -e "s/https:\/\/gitlab-ci-token:.*@//g"`
git remote set-url origin "https://gitlab-ci-token:${GITLAB_TOKEN}@${URL}"

# tag should trigger a pipeline 
git push origin --tags 
