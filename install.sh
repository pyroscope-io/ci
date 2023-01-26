##!/bin/sh

set -e

VERSION="${1:-latest}"

# https://en.wikipedia.org/wiki/Uname
case $(uname -sm) in
"Linux x86_64") target="linux-amd64" ;;
"Linux aarch64") target="linux-arm64" ;;
*)
  echo "Error: unsupported os/arch: '$(uname -sm)'" 1>&2
  exit 1
  ;;
esac

if ! command -v curl >/dev/null; then
	echo "Error: 'curl' is required" 1>&2
	exit 1
fi

if ! command -v tar >/dev/null; then
	echo "Error: 'tar' is required" 1>&2
	exit 1
fi

downloadUrl="https://github.com/pyroscope-io/ci/releases/${VERSION}/download/pyroscope-ci-${target}.tar.gz"
curl --fail --location --progress-bar --output "pyroscope-ci.tar.gz" "$downloadUrl" 
tar -zxvf "pyroscope-ci.tar.gz" "pyroscope-ci"
rm "pyroscope-ci.tar.gz"
chmod +x "pyroscope-ci"

if ! command -v ./pyroscope-ci >/dev/null; then
	echo "An error has occurred: binary './pyroscope-ci' was not installed properly"
  exit 1
fi

echo "pyroscope-ci has been downloaded locally"
echo "run ./pyroscope-ci --help for more information."
