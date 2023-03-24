#!/bin/bash

# commit="641c39939ba19b33f5df9bbafdf8d4f0949e70e9"

# lock="I'm your testcluster."
# give_up="I was too late."

# msg="${give_up}"
# cluster="aws.eu-central-1.infra-tests"
# data="{
#   \"state\": \"failure\",
#   \"description\": \"${msg}\",
#   \"context\": \"lock/test/${cluster}\"
# }";
# # data="{
# #   \"state\": \"success\",
# #   \"description\": \"done\",
# #   \"context\": \"PR/intber-suite\"
# # }";
# curl -H "Content-Type: application/json" \
#   -H "Authorization: token ${GITHUB_TOKEN}" \
#   -X POST \
#   --data "${data}" \
#   "https://github.tools.sap/api/v3/repos/MLF/mlf-gitops/statuses/${commit}";


# resp=$(curl -H "Content-Type: application/json" \
#   -H "Authorization: token ${GITHUB_TOKEN}" \
#   -X GET \
#   "https://github.tools.sap/api/v3/repos/MLF/mlf-gitops/commits/${commit}/status" \
#   --silent);

# # echo $resp | jq '.'
# echo $resp | jq '.statuses[] | select( .context|test("lock/test/"))'


# scopes needed to move tags:
# public_repo, repo:status, repo_deployment
# curl \
#   -X PATCH \
#   -H "Accept: application/vnd.github.v3+json" \
#   -H "Authorization: token ${GITHUB_TOKEN}" \
#   https://github.tools.sap/api/v3/repos/MLF/mlf-gitops/git/refs/tags/test-tag \
#   -d '{"sha":"fee182a8ebb19e50ae46661547fb69b83abe219b"}'

repo_url="https://github.wdf.sap.corp/api/v3/repos/ICN-ML/aicore"
# tag="rel/2.16.0"
release_id="214842"
resp=$(curl \
  -H "Accept: application/vnd.github.v3+json" \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  "${repo_url}/releases/${release_id}"
)

# data="{
#   \"tag_name\": \"test1\",
#   \"target_commitish\": \"coco/file-generation\",
#   \"name\": \"test1\",
#   \"body\": \"test release\",
#   \"draft\": false,
#   \"prerelease\": false
# }"
# resp=$(curl \
#   -X POST \
#   -H "Accept: application/vnd.github.v3+json" \
#   -H "Authorization: token ${GITHUB_TOKEN}" \
#   "${repo_url}/releases" \
#   -d "${data}")

# echo $resp
# echo $resp | jq '.id'



# release_id="214842"
bin_location="${HOME}/src/github.com/configuration-tools-for-gitops/dist"


# upload_url="https://github.wdf.sap.corp/api/uploads/repos/ICN-ML/aicore"
# upload_url="https://github.wdf.sap.corp/api/uploads/repos/ICN-ML/aicore/releases/214842/assets{?name,label}"
upload_url=$(echo $resp | jq -r '.upload_url' | cut -f1 -d"{")
echo "$upload_url"
curl \
  -X POST \
  -H "Accept: application/vnd.github.v3+json" \
  -H "Content-Type: application/octet-stream" \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  --data-binary @"${bin_location}/coco-linux-amd64" \
  "${upload_url}?name=test"
  # ?name=coco-linux-amd64