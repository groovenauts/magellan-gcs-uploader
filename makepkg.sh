#!/bin/bash

APPCFG=$(which appcfg.py)
if [ "${APPCFG}" = "" ]
then
  echo "Cannot find appcfg.py in PATH. Please set environment variable."
  exit 1
fi

VERSION="$1"
if [ "${VERSION}" = "" ]
then
  echo "Usage: $0 version"
  exit 1
fi

SDK_BASE=$(dirname "${APPCFG}")
GO_APP_BUILDER="${SDK_BASE}/goroot/bin/go-app-builder"

APP_BASE=$(dirname "$0")

SRCS=magellan-log-collector.go
PKGDIR="${APP_BASE}/pkg/${VERSION}"
MANIFEST="${PKGDIR}/_manifest"

export GOOS="linux"
export GOARCH="amd64"

mkdir -p "${PKGDIR}"

echo -n > "${MANIFEST}"

for line in $(${GO_APP_BUILDER} -api_version go1 -app_base "${APP_BASE}" -arch 6 -print_extras -goroot "${SDK_BASE}/goroot" ${SRCS})
do
  filepath=$(echo "${line}" | cut -f 1 -d \|)
  originalpath=$(echo "${line}" | cut -f 2 -d \|)
  mkdir -p "$(dirname "${PKGDIR}/${filepath}")"
  cp "${originalpath}" "${PKGDIR}/${filepath}"
  (cd "${PKGDIR}" && sha1sum "${filepath}") >> "${MANIFEST}"
done

for filepath in ${SRCS}
do
  mkdir -p "$(dirname "${PKGDIR}/${filepath}")"
  cp "${filepath}" "${PKGDIR}/${filepath}"
  (cd "${PKGDIR}" && sha1sum "${filepath}") >> "${MANIFEST}"
done
