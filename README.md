# magellan-gcs-uploader

## Run local server

```
export GO111MODULE=on
go run magellan-gcs-uploader.go
```

## Deploy

Specify GCP project id and api tokens (comma separated).

```
export GO111MODULE=on
(Edit app.yaml to setup environment variables)
gcloud --project=YOUR-PROJECT-ID app deploy app.yaml
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

Run the following command to gather source files and make manifest file, and upload them to gcs.
`v1` stands for the version of application.

```
export GO111MODULE=on
./makepkg.sh your-gae-repository v1
```

This workflow is automated by [Release workflow](.github/workflows/release.yml),
triggered by tag push.

