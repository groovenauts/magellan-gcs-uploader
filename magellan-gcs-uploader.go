package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/cloud/bigquery"
	"google.golang.org/cloud/storage"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	apiTokens []string
)

func init() {
	apiTokens = nil
	http.HandleFunc("/upload", postHandler)
}

func mustGetenv(ctx context.Context, k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Criticalf(ctx, "%s environment variable not set.", k)
	}
	return v
}

func verifyApiToken(token string) error {
	for _, x := range apiTokens {
		if x == token {
			return nil
		}
	}
	return errors.New("invalid api token.")
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type BigQueryRecord struct {
	Row map[string]bigquery.Value
}

func (r *BigQueryRecord) Save() (row map[string]bigquery.Value, insertID string, err error) {
	return r.Row, "", nil
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	projectId := appengine.AppID(ctx)
	response := Response{false, "something wrong."}
	code := 500

	defer func() {
		outjson, e := json.Marshal(response)
		if e != nil {
			log.Errorf(ctx, e.Error())
		}
		w.Header().Set("Content-Type", "application/json")
		if code == 200 {
			fmt.Fprint(w, string(outjson))
		} else {
			http.Error(w, string(outjson), code)
		}
	}()

	if r.Method != "POST" {
		response.Message = "only POST method method was accepted"
		code = 404
		return
	}

	// Check API Token
	api_key := r.FormValue("key")
	if apiTokens == nil {
		apiTokens = strings.Split(mustGetenv(ctx, "API_TOKEN"), ",")
	}
	err := verifyApiToken(api_key)
	if err != nil {
		response.Message = err.Error()
		code = 401
		return
	}

	// Upload file to GCS
	filename := r.FormValue("filename")
	content, err := base64.StdEncoding.DecodeString(r.FormValue("content"))
	if err != nil {
		response.Message = "content parameter: invalid base64 encoded."
		code = 400
		return
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Criticalf(ctx, err.Error())
		response.Message = err.Error()
		code = 500
		return
	}
	bucket_name := mustGetenv(ctx, "STORAGE_BUCKET")

	bucket := client.Bucket(bucket_name)

	object := bucket.Object(filename)
	writer := object.NewWriter(ctx)

	_, err = writer.Write(content)
	if err != nil {
		log.Criticalf(ctx, err.Error())
		code = 500
		return
	}
	err = writer.Close()
	if err != nil {
		log.Criticalf(ctx, err.Error())
		code = 500
		return
	}

	dataset_id := mustGetenv(ctx, "BIGQUERY_DATASET")
	table_id := mustGetenv(ctx, "BIGQUERY_TABLE")
	if dataset_id != "" && table_id != "" {
		// Insert metadata to BigQuery
		bqclient, err := bigquery.NewClient(ctx, projectId)
		if err != nil {
			log.Criticalf(ctx, err.Error())
			code = 500
			return
		}
		table := bqclient.OpenTable(projectId, dataset_id, table_id)
		uploader := table.NewUploader()
		record := &BigQueryRecord{}
		record.Row = make(map[string](bigquery.Value))
		record.Row["gcs_url"] = "gs://" + bucket_name + "/" + filename
		record.Row["timestamp"] = time.Now()
		columns := mustGetenv(ctx, "BIGQUERY_COLUMNS")
		field_names := strings.Split(columns, ",")
		for _, fn := range field_names {
			val := r.FormValue(fn)
			if val != "" {
				record.Row[fn] = val
			}
		}
		err = uploader.Put(ctx, record)
		if err != nil {
			log.Criticalf(ctx, err.Error())
			code = 500
			return
		}
	}

	response.Success = true
	response.Message = "ok"
	code = 200
	return
}
