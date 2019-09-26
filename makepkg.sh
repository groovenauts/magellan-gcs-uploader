#!/bin/bash

DIR=$(dirname $0)
DIR=$(realpath ${DIR})

VERSION="$1"
if [ "${VERSION}" = "" ]
then
  echo "Usage: $0 version"
  exit 1
fi

export GOPATH=${DIR}

echo "set GOPATH=${GOPATH}"
echo "go get -d"

go get -d

APP_BASE=$(dirname "$0")

SRCS=magellan-gcs-uploader.go
PKGDIR="${APP_BASE}/pkg/${VERSION}"
MANIFEST="${PKGDIR}/_manifest"

mkdir -p "${PKGDIR}"

echo -n > "${MANIFEST}"

for f in $(find src -name "*.go" -a -not -name "*_test.go" | sed -e 's/src\///g'); do
  originalpath="src/${f}"
  destf="_gopath/src/${f}"
  mkdir -p "$(dirname "${PKGDIR}/${destf}")"
  cp "${originalpath}" "${PKGDIR}/${destf}"
  (cd "${PKGDIR}" && sha1sum "${destf}") >> "${MANIFEST}"
done

for filepath in ${SRCS}
do
  mkdir -p "$(dirname "${PKGDIR}/${filepath}")"
  cp "${filepath}" "${PKGDIR}/${filepath}"
  (cd "${PKGDIR}" && sha1sum "${filepath}") >> "${MANIFEST}"
done

mkdir -p "${PKGDIR}/_gopath"
echo -n 'github.com/groovenauts/magellan-gcs-uploader' > "${PKGDIR}/_gopath/main-package-path"
(cd "${PKGDIR}" && sha1sum "_gopath/main-package-path") >> "${MANIFEST}"
