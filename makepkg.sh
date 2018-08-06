#!/bin/bash

GCLOUD=$(which gcloud)
if [ "${GCLOUD}" = "" ]
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

API_VERSION="go1.9"
SDK_BASE=$(dirname $(dirname "${GCLOUD}"))
GOROOT="${SDK_BASE}/platform/google_appengine/goroot-1.9"
GO_APP_BUILDER="${SDK_BASE}/platform/google_appengine/goroot-1.9/bin/go-app-builder"

APP_BASE=$(dirname "$0")

SRCS=magellan-gcs-uploader.go
PKGDIR="${APP_BASE}/pkg/${VERSION}"
MANIFEST="${PKGDIR}/_manifest"

export GOOS="linux"
export GOARCH="amd64"

mkdir -p "${PKGDIR}"

echo -n > "${MANIFEST}"

for line in $(${GO_APP_BUILDER} -api_version ${API_VERSION} -app_base "${APP_BASE}" -arch 6 -print_extras -goroot "${GOROOT}" ${SRCS})
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
