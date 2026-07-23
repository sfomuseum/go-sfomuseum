# go-sfomuseum-twitter-publish

## Work in progress

Documentation to follow.

## Tools

```
$> make cli
go build -mod vendor -o bin/publish cmd/publish/main.go
```

### publish

```
> ./bin/publish -h
Usage of ./bin/publish:
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-index URI (default "repo://")
  -iterator-source string
        ...  		   
  -reader-uri string
    	A valid whosonfirst/go-reader URI
  -trim-prefix
    	Trim default tweet.js JavaScript prefix. (default true)
  -tweets-uri tweets.js
    	A valid gocloud.dev/blob URI to your tweets.js file.
  -writer-uri string
    	A valid whosonfirst/go-writer URI
```

Import (or update) Twitter messages contained in a `tweet.js` file produced by the Twitter export data process in to a `sfomuseum-data-twitter` repository.

For example:

```
$> bin/publish \
	-reader-uri fs:///usr/local/data/sfomuseum-data-twitter/data \
	-writer-uri fs:///usr/local/data/sfomuseum-data-twitter/data \
	-tweets-uri file:///usr/local/data/twitter/data/tweet.js \
	-iterator-uri repo:// \
	-iterator-source /usr/local/data/sfomuseum-data-twitter/data
```

## See also

* https://github.com/sfomuseum/go-sfomuseum-twitter
