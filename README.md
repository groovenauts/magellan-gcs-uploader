# magellan-gcs-uploader

## Run local server

```
goapp serve
```

## Deploy

Specify GCP project id and api tokens (comma separated).

```
appcfg.py -A YOUR-PROJECT-ID -E STORAGE_BUCKET:YOUR-BUCKET-NAME -E API_TOKEN:XXXXXXXX update .
```

## Upload package to deploy via Google App Engine Admin API

Run the following command To gather source files and make manifest file. `v1` stands for the version of application.

```
./makepkg.sh v1
```

Upload source files and a manifest file to gcs.

```
gsutil cp -R pkg/v1 gs://your-gae-repository/magellan-gcs-uploader/v1
```
