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

### Environment Variables

| Name | Required | Description |
|------|----------|-------------|
| `STORAGE_BUCKET` | o | Google Cloud Storage Bucket name which files to be stored into. |
| `API_TOKEN` | o | API Tokens (comma separated) to be compared with request's key param for authorization |
| `BIGQUERY_DATASET` | x | BigQuery Dataset Id which insert metadata into. |
| `BIGQUERY_TABLE`| x | BigQuery Table Id which insert metadata into. |
| `BIGQUERY_COLUMNS` | x | Additional parameters (comma separated) insert into BigQuery table as metadata. |
| `BLOCKS_URL` | x | Hook URL to invoke BLOCKS flow. |
| `BLOCKS_API_TOKEN` | x | API Tokent for BLOCKS Board. |

## Upload package to deploy via Google App Engine Admin API

Run the following command To gather source files and make manifest file. `v1` stands for the version of application.

```
./makepkg.sh v1
```

Upload source files and a manifest file to gcs.

```
gsutil cp -R pkg/v1 gs://your-gae-repository/magellan-gcs-uploader/v1
```
