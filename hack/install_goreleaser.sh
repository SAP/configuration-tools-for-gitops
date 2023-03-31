#!/bin/bash
set -e

install_to=${1}
expected_version=${2}

function cleanup() {
  rm -rf "${TMPDIR}"
}

export TMPDIR="$(mktemp -d)"
trap cleanup EXIT

if [[ "${install_to}" == "" ]]; then
  echo "[ERROR] no installation folder provided"
  exit 1
fi


RELEASES_URL="https://github.com/goreleaser/goreleaser/releases"
FILE_BASENAME="goreleaser"
LATEST="$(curl -sf https://goreleaser.com/static/latest)"

test -z "$expected_version" && expected_version="$LATEST"

test -z "$expected_version" && {
	echo "[ERROR] Unable to get goreleaser version." >&2
	exit 1
}



if [ -f "${install_to}/goreleaser" ]; then
  version_str=($(${install_to}/goreleaser --version | head -n 1))
  version="v${version_str[2]}"
fi

if [[ "${version}" == "${expected_version}" ]]; then
  echo "goreleaser version ${version} already present at \"${install_to}/goreleaser\""
  exit 0
fi
export VERSION="${expected_version}"

export TAR_FILE="$TMPDIR/${FILE_BASENAME}_$(uname -s)_$(uname -m).tar.gz"

(
	cd "$TMPDIR"
	echo "Downloading GoReleaser $VERSION..."
  arch=$(uname -m)
  if [[ "${arch}" == "aarch64" ]]; then
    arch="arm64"
  fi
  release_url="$RELEASES_URL/download/$VERSION/${FILE_BASENAME}_$(uname -s)_${arch}.tar.gz"

	curl -sfLo "$TAR_FILE" "$release_url"
	curl -sfLo "checksums.txt" "$RELEASES_URL/download/$VERSION/checksums.txt"
)

tar -xf "$TAR_FILE" -C "$TMPDIR"

mv "$TMPDIR/goreleaser" "${install_to}/goreleaser"

