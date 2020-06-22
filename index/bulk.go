package index

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	es "github.com/elastic/go-elasticsearch/v7"
	// "github.com/elastic/go-elasticsearch/v7/estransport"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/sfomuseum/go-twitter-elasticsearch/document"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"runtime"
	"time"
)

const FLAG_ES_ENDPOINT string = "elasticsearch-endpoint"
const FLAG_ES_INDEX string = "elasticsearch-index"
const FLAG_WORKERS string = "workers"

const FLAG_APPEND_TIMESTAMP string = "append-timestamp"
const FLAG_APPEND_UNSHORTENED_URLS string = "append-unshortened-urls"
const FLAG_APPEND_ALL string = "append-all"

func NewBulkIndexerFlagSet(ctx context.Context) (*flag.FlagSet, error) {

	fs := flagset.NewFlagSet("bulk")

	fs.String(FLAG_ES_ENDPOINT, "http://localhost:9200", "The URL of your Elasticsearch endpoint.")
	fs.String(FLAG_ES_INDEX, "twitter", "The name of your Elasticsearch index.")
	fs.Int(FLAG_WORKERS, runtime.NumCPU(), "The number of concurrent indexers")

	fs.Bool(FLAG_APPEND_TIMESTAMP, true, "Append a Unix timestamp to each post.")
	fs.Bool(FLAG_APPEND_UNSHORTENED_URLS, true, "Append unshortened URLs to each post.")
	fs.Bool(FLAG_APPEND_ALL, false, "Enable all -append related flags.")

	// debug := fs.Bool("debug", false, "...")

	return fs, nil
}

func RunBulkIndexerWithFlagSet(ctx context.Context, fs *flag.FlagSet) (*esutil.BulkIndexerStats, error) {

	es_endpoint, err := lookup.StringVar(fs, FLAG_ES_ENDPOINT)

	if err != nil {
		return nil, err
	}

	es_index, err := lookup.StringVar(fs, FLAG_ES_INDEX)

	if err != nil {
		return nil, err
	}

	workers, err := lookup.IntVar(fs, FLAG_WORKERS)

	if err != nil {
		return nil, err
	}

	append_timestamp, err := lookup.BoolVar(fs, FLAG_APPEND_TIMESTAMP)

	if err != nil {
		return nil, err
	}

	append_unshortened_urls, err := lookup.BoolVar(fs, FLAG_APPEND_UNSHORTENED_URLS)

	if err != nil {
		return nil, err
	}

	append_all, err := lookup.BoolVar(fs, FLAG_APPEND_ALL)

	if err != nil {
		return nil, err
	}

	if append_all {

		append_timestamp = true
		append_unshortened_urls = true
	}

	retry := backoff.NewExponentialBackOff()

	es_cfg := es.Config{
		Addresses: []string{es_endpoint},

		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retry.Reset()
			}
			return retry.NextBackOff()
		},
		MaxRetries: 5,
	}

	/*

		if debug {

			es_logger := &estransport.ColorLogger{
				Output:             os.Stdout,
				EnableRequestBody:  true,
				EnableResponseBody: true,
			}

			es_cfg.Logger = es_logger
		}

	*/

	es_client, err := es.NewClient(es_cfg)

	if err != nil {
		return nil, err
	}

	_, err = es_client.Indices.Create(es_index)

	if err != nil {
		return nil, err
	}

	// https://github.com/elastic/go-elasticsearch/blob/master/_examples/bulk/indexer.go

	bi_cfg := esutil.BulkIndexerConfig{
		Index:         es_index,
		Client:        es_client,
		NumWorkers:    workers,
		FlushInterval: 30 * time.Second,
	}

	bi, err := esutil.NewBulkIndexer(bi_cfg)

	if err != nil {
		return nil, err
	}

	IndexPath := func(ctx context.Context, path string) error {

		// TO DO: READ FROM GOCLOUD BUCKET

		fh, err := os.Open(path)

		if err != nil {
			return err
		}

		defer fh.Close()

		var posts []interface{}

		dec := json.NewDecoder(fh)
		err = dec.Decode(&posts)

		if err != nil {
			return err
		}

		for _, tw := range posts {

			tw_body, err := json.Marshal(tw)

			if err != nil {
				msg := fmt.Sprintf("Failed to marshal %s, %v", path, err)
				return errors.New(msg)
			}

			id_rsp := gjson.GetBytes(tw_body, "id")

			if !id_rsp.Exists() {
				return errors.New("Can't find ID")
			}

			doc_id := id_rsp.String()

			if append_timestamp {
				tw_body, err = document.AppendCreatedAtTimestamp(ctx, tw_body)

				if err != nil {
					return err
				}
			}

			if append_unshortened_urls {
				tw_body, err = document.AppendUnshortenedURLs(ctx, tw_body)

				if err != nil {
					return err
				}
			}

			// log.Println(string(enc_f))

			bulk_item := esutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: doc_id,
				Body:       bytes.NewReader(tw_body),

				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					// log.Printf("Indexed %s\n", path)
				},

				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Printf("ERROR: Failed to index %s, %s", path, err)
					} else {
						log.Printf("ERROR: Failed to index %s, %s: %s", path, res.Error.Type, res.Error.Reason)
					}
				},
			}

			err = bi.Add(ctx, bulk_item)

			if err != nil {
				log.Printf("Failed to schedule %s, %v", path, err)
				return nil
			}
		}

		return nil
	}

	paths := fs.Args()

	t1 := time.Now()

	for _, path := range paths {

		err := IndexPath(ctx, path)

		if err != nil {
			return nil, err
		}
	}

	err = bi.Close(ctx)

	if err != nil {
		return nil, err
	}

	log.Printf("Processed files in %v\n", time.Since(t1))

	stats := bi.Stats()
	return &stats, nil
}
