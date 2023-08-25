#!/bin/bash
set -e

package="${1}"
testFunc="${2}"

if [[ "${package}" == "" ]]; then
	echo "[ERROR] environment variable \"package\" must be set" >&2
	exit 1
fi

cmd=(go test -timeout 30s -v -race)
if [[ "${testFunc}" != "" ]]; then
	cmd+=(-run "^${testFunc}$")
fi
cmd+=("github.com/SAP/configuration-tools-for-gitops/v2/${package}")
echo "${cmd[*]}"
eval $(echo "${cmd[*]}")
