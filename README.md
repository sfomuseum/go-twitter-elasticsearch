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

For example:

```
$> ./bin/es-twitter-index -append-all /usr/local/data/tweet.js
2020/10/02 12:02:06 http://gowal.la/p/hHG2 Head "http://gowal.la/p/hHG2": dial tcp: lookup gowal.la on 0.0.0.0:53: no such host
2020/10/02 12:02:38 http://fb.me/1rL5puFkV Head "http://danceonmarket.com/events/": dial tcp: lookup danceonmarket.com on 0.0.0.0:53: no such host
2020/10/02 12:05:58 http://wrd.cm/18JTa7d Head "http://oak.ctx.ly/r/k0fw": context deadline exceeded
2020/10/02 12:07:06 http://ow.ly/sfvv4 Head "https://fbcdn-sphotos-b-a.akamaihd.net/hphotos-ak-prn2/1492506_197934060399828_748773257_o.jpg": dial tcp: lookup fbcdn-sphotos-b-a.akamaihd.net on 0.0.0.0:53: no such host
2020/10/02 12:07:25 http://ow.ly/t9Wu6 Head "http://blog.nationalgeographic.org2014/01/30/shining-a-light-on-the-hidden-world-of-women-cartographers/": dial tcp: lookup blog.nationalgeographic.org2014 on 0.0.0.0:53: no such host
2020/10/02 12:08:18 http://ow.ly/uAWZH Head "https://blogs.sfweekly.com/exhibitionist/2014/03/sf_filters_top_10_san_francisc.php": x509: certificate is valid for *.fdncms.com, fdncms.com, not blogs.sfweekly.com
2020/10/02 12:08:52 http://buzz.mw/bppwg_f Head "http://buzz.mw/bppwg_f": dial tcp: lookup buzz.mw on 0.0.0.0:53: no such host
...
```

## Notes

* This assumes Elasticsearch 7.x