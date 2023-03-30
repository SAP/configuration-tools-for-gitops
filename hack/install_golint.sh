#!/usr/bin/env bash

set -e

url="https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"

function main() {

  install_to=${1}
  expected_version=${2}

  if [ -f "${install_to}/golangci-lint" ]; then
    version_str=($(${install_to}/golangci-lint --version))
    version="${version_str[3]}"
  fi

  if [[ "v${version}" != "${expected_version}" ]]; then
    echo reinstall
    curl -sSfL "${url}" \
      | sh -s -- -b "${install_to}" "${expected_version}"
  fi

}

main "${@}"

