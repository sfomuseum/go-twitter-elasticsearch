# go-twitter-elasticsearch

Go package for indexing a Twitter backup archive (JSON) in Elasticsearch.

## Important

This is work in progress. Documentation to follow.

## Tools

### es-twitter-index

```
> go run -mod vendor cmd/es-twitter-index/main.go -h
  -append-all
	Enable all -append related flags.
  -append-timestamp
	Append a Unix timestamp to each post. (default true)
  -append-unshortened-urls
	Append unshortened URLs to each post. (default true)
  -elasticsearch-endpoint string
    			  The URL of your Elasticsearch endpoint. (default "http://localhost:9200")
  -elasticsearch-index string
    		       The name of your Elasticsearch index. (default "twitter")
  -workers int
    	   The number of concurrent indexers (default 2)
```

## Notes

* This assumes Elasticsearch 7.x