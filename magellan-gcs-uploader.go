package main

import (
	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/storage"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"net/url"
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

func postBlocksFlow(ctx context.Context, blocks_url, blocks_api_token, gcs_url string, timestamp time.Time, r *http.Request) error {
	values := url.Values{}
	values.Set("api_token", blocks_api_token)
	values.Set("gcs_url", gcs_url)
	values.Set("target_time", timestamp.Format("2006-01-02T15:04:05.999999Z07:00"))
	for k, v := range r.Form {
		if k == "content" || k == "key" {
			continue
		}
		if v[0] != "" {
			values.Set(k, v[0])
		}
	}
	client := urlfetch.Client(ctx)
	res, err := client.PostForm(blocks_url, values)
	if err == nil {
		defer res.Body.Close()
	}

	return err
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
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
		response.Message = err.Error()
		code = 500
		return
	}
	err = writer.Close()
	if err != nil {
		log.Criticalf(ctx, err.Error())
		response.Message = err.Error()
		code = 500
		return
	}

	gcs_url := "gs://" + bucket_name + "/" + filename
	timestamp := time.Now()

	dataset_id := os.Getenv("BIGQUERY_DATASET")
	table_id := os.Getenv("BIGQUERY_TABLE")
	if dataset_id != "" && table_id != "" {
		// Insert metadata to BigQuery
		bqclient, err := bigquery.NewClient(ctx, projectId)
		if err != nil {
			log.Criticalf(ctx, err.Error())
			response.Message = err.Error()
			code = 500
			return
		}
		table := bqclient.Dataset(dataset_id).Table(table_id)
		uploader := table.Uploader()
		record := &BigQueryRecord{}
		record.Row = make(map[string](bigquery.Value))
		record.Row["gcs_url"] = gcs_url
		record.Row["timestamp"] = timestamp
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
			response.Message = err.Error()
			code = 500
			return
		}
	}

	blocks_url := os.Getenv("BLOCKS_URL")
	blocks_api_token := os.Getenv("BLOCKS_API_TOKEN")
	if blocks_url != "" && blocks_api_token != "" {
		err = postBlocksFlow(ctx, blocks_url, blocks_api_token, gcs_url, timestamp, r)
		if err != nil {
			log.Criticalf(ctx, err.Error())
			response.Message = err.Error()
			code = 500
			return
		}
	}

	response.Success = true
	response.Message = "ok"
	code = 200
	return
}
