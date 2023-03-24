#!/usr/bin/env bash

set -e

function main() {

  parse_parameters "${@}"

  data="{
    \"tag_name\": \"${GIT_TAG}\",
    \"target_commitish\": \"${GIT_COMMIT}\",
    \"name\": \"${GIT_TAG}\",
    \"body\": \"coco release ${VERSION}\",
    \"draft\": false,
    \"prerelease\": false
  }";
  resp=$(curl \
    -X POST \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Authorization: token ${GIT_PASSWORD}" \
    "${REPO_URL}/releases" \
    -d "${data}"
  );
  UPLOAD_URL=$(echo $resp | jq -r '.upload_url' | cut -f1 -d"{");

  upload "coco-linux-amd64.tar.gz"
  upload "coco-linux-arm64.tar.gz"
  upload "coco-darwin-amd64.tar.gz"
  upload "coco-darwin-arm64.tar.gz"

  upload "login-linux-amd64.tar.gz"
  upload "login-linux-arm64.tar.gz"
  upload "login-darwin-amd64.tar.gz"
  upload "login-darwin-arm64.tar.gz"
}

function upload() {
  local -r name="${1}"
  # from env
  token="${GIT_PASSWORD}"
  url="${UPLOAD_URL}"
  file="${BIN_LOCATION}/${name}"

  curl \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: application/octet-stream" \
    -H "Authorization: token ${token}" \
    --data-binary @"${file}" \
    "${url}?name=${name}";
}


function parse_parameters() {
  input_params="${@}"

  while [ "$#" -gt 0 ]
  do
    case "$1" in
    --git-tag)
      export GIT_TAG="${2}"
      shift
      ;;
    --git-commit)
      export GIT_COMMIT="${2}"
      shift
      ;;
    --version)
      export VERSION="${2}"
      shift
      ;;
    --repo-url)
      export REPO_URL="${2}"
      shift
      ;;
    --bin-location)
      export BIN_LOCATION="${2}"
      shift
      ;;
    --)
      break
      ;;
    -*)
      echo "Invalid option '$1'. Use --help to see the valid options" >&2
      exit 1
      ;;
    *)  
      break
      ;;
    esac
    shift
  done
  check_env_vars GIT_PASSWORD

  check_required_parameters GIT_TAG
  check_required_parameters GIT_COMMIT
  check_required_parameters VERSION
  check_required_parameters REPO_URL
  check_required_parameters BIN_LOCATION

}

function check_env_vars() {
  if [ -z "${!1}" ]; then
    error "environment variable \"${1}\" is not set"
    exit 1
  fi
}

function check_required_parameters() {
  if [ -z "${!1}" ]; then
    error "parameter \"${1}\" is not set"
    exit 1
  fi
}

function error() {
  echo -e "\033[1;31m[-] ${error_prefix} (${script_name}): $*\033[0m";
}

main "${@}"

