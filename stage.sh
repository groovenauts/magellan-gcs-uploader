#! /bin/bash

set -eo pipefail

app_name=$(basename -s .git $(git config --get remote.origin.url))
app_root=$(git rev-parse --show-toplevel)
project_id=$(gcloud config get-value project)

bucket="$1"
version="$2"

if [ -z "${bucket}" -o -z "${version}" ]; then
  echo "Usage: $0 bucket version"
  exit 1
fi

gcs_version_path=gs://${bucket}/${app_name}/${version}

echo ">>> app_name: ${app_name}"
echo ">>> app_root: ${app_root}"
echo ">>> project_id: ${project_id}"
echo ">>> bucket: ${bucket}"
echo ">>> version: ${version}"
echo ">>> gcs_version_path: ${gcs_version_path}"

# check if already uploaded
if gsutil -q stat ${gcs_version_path}/_manifest; then
  echo "Abort. Manifest file ${gcs_version_path}/_manifest is already uploaded."
  exit 1
fi

# create temporary directory and clean up on exit
tmpdir=$(mktemp -d)
stagedir1=$tmpdir/stage1
stagedir2=$tmpdir/stage2
mkdir ${stagedir1} ${stagedir2}
trap "echo '>>> Cleanup'; rm -rf ${tmpdir}" EXIT
echo ">>> tmpdir: ${tmpdir}"
echo ">>> stagedir1: ${stagedir1}"
echo ">>> stagedir2: ${stagedir2}"

# load configuration file
if [ ! -f $(dirname $0)/stage-config.sh ]; then
  echo "stage-config.sh not found."
  exit 1
fi

. $(dirname $0)/stage-config.sh

if [ -z "${runtime}" ]; then
  echo "runtime is not specified in stage-config.sh"
  exit 1
fi

if [ -n "${app_dir}" ]; then
  app_dir=${app_root}/${app_dir}
else
  app_dir=${app_root}
fi

echo "runtime: ${runtime}"
echo "app_dir: ${app_dir}"

# create app.yaml in temporary directory
echo "runtime: ${runtime}" > ${tmpdir}/app.yaml
echo ">>> runtime: ${runtime}"

# "goNNN" -> "N.NN"
go_version=${runtime#go}
go_version=${go_version::1}.${go_version:1}
echo ">>> go_version: ${go_version}"

# check if go-app-stager is installed
go_app_stager=$(gcloud info --format='value(installation.sdk_root)')/platform/google_appengine/go-app-stager
if [ ! -x "${go_app_stager}" ]; then
  echo "go-app-stager not found. Run \"gcloud components install app-engine-go\" to install go-app-stager."
  exit 1
fi

# copy files into stage directory using go-app-stager
echo ">>> Copy files by ${go_app_stager} into stagedir1"
${go_app_stager} -go-version=${go_version} ${tmpdir}/app.yaml ${app_dir} ${stagedir1}

# prepare upload files
cd ${stagedir1}
echo ">>> Copy files with \"gcloud meta list-files-for-upload\" into stagedir2"
gcloud meta list-files-for-upload | sort > ${tmpdir}/files-for-upload
cat ${tmpdir}/files-for-upload | while read f; do
  install -p -D ${f} ${stagedir2}/${f}
done

# upload files to Cloud Storage
cd ${stagedir2}
echo ">>> Upload files to cloud storage"
gsutil -m -q rsync -r . ${gcs_version_path}

# upload manifest
echo ">>> Generate and upload menifest file to cloud storage"
sha1sum $(cat ${tmpdir}/files-for-upload) > ${tmpdir}/_manifest
echo ">>> * ${tmpdir}/_manifest -> ${gcs_version_path}/_manifest"
gsutil -q cp ${tmpdir}/_manifest ${gcs_version_path}/_manifest

# list uploaded objects
echo ">>> List objects in ${gcs_version_path}/"
gsutil ls -l -r ${gcs_version_path}/

# show manifest
echo ">>> Show manifest content in ${gcs_version_path}/_manifest"
gsutil cat ${gcs_version_path}/_manifest
