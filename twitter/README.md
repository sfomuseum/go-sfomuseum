# go-sfomuseum-twitter

Go package for working with Twitter archives.

## Tools

To build binary versions of these tools run the `cli` Makefile target. For example:

```
$> make cli
go build -mod vendor -o bin/emit cmd/emit/main.go
go build -mod vendor -o bin/pointers cmd/pointers/main.go
go build -mod vendor -o bin/trim cmd/trim/main.go
go build -mod vendor -o bin/unshorten cmd/unshorten/main.go
```

### emit

```
$> ./bin/emit -h
Usage of ./bin/emit:
  -append-all -append-
    	Enable all the -append- flags.
  -append-timestamp created
    	Append a created property containing a Unix timestamp derived from the `created_at` property.
  -append-urls unshortened_url
    	Append a unshortened_url property for each `entities.urls.(n)` property.
  -format-json
    	Format JSON output for each record.
  -json
    	Emit a JSON list.
  -null
    	Emit to /dev/null
  -query value
    	One or more {PATH}={REGEXP} parameters for filtering records.
  -query-mode string
    	Specify how query filtering should be evaluated. Valid modes are: ALL, ANY (default "ALL")
  -stdout
    	Emit to STDOUT (default true)
  -trim-prefix
    	Trim default tweet.js JavaScript prefix. (default true)
  -tweets-uri tweets.js
    	A valid gocloud.dev/blob URI to your tweets.js file.
```	

For example:

```
./bin/emit \
	-append-all \
	-json \
	-format-json \
	-tweets-uri file:///usr/local/twitter/data/tweet.js
[
{
  "created_at": "Mon Sep 19 19:21:04 +0000 2011",
  "display_text_range": ["0", "88"],
  "entities": {
    "hashtags": [],
    "symbols": [],
    "urls": [
      {
        "display_url": "bit.ly/q8nobK",
        "expanded_url": "http://bit.ly/q8nobK",
        "indices": ["68", "88"],
        "url": "http://t.co/aGv43tHf",
        "unshortened_url": "https://www.flysfo.com/web/page/sfo_museum/exhibitions/terminal1_exhibitions/B3_archive/robert_apte/01.html"
      }
    ],
    "user_mentions": []
  },
  "favorite_count": "0",
  "favorited": false,
  "full_text": "Is anyone else hot? How about an Antarctic iceberg to cool you off: http://t.co/aGv43tHf",
  "id": "115868023763632128",
  "id_str": "115868023763632128",
  "lang": "en",
  "possibly_sensitive": false,
  "retweet_count": "0",
  "retweeted": false,
  "source": "\u003ca href=\"https://about.twitter.com/products/tweetdeck\" rel=\"nofollow\"\u003eTweetDeck\u003c/a\u003e",
  "truncated": false,
  "created": 1316460064
}
...and so on
]
```

#### Inline queries

You can also specify inline queries by passing a `-query` parameter which is a string in the format of:

```
{PATH}={REGULAR EXPRESSION}
```

Paths follow the dot notation syntax used by the [tidwall/gjson](https://github.com/tidwall/gjson) package and regular expressions are any valid [Go language regular expression](https://golang.org/pkg/regexp/). Successful path lookups will be treated as a list of candidates and each candidate's string value will be tested against the regular expression's [MatchString](https://golang.org/pkg/regexp/#Regexp.MatchString) method.

For example:

```
> ./bin/emit \
	-json \
	-format-json \
	-query 'full_text=\bSFO\b' \
	-tweets-uri file:///usr/local/twitter/data/tweet.js
	
[{
  "created_at": "Tue Sep 20 18:30:53 +0000 2011",
  "display_text_range": ["0", "140"],
  "entities": {
    "hashtags": [],
    "symbols": [],
    "urls": [],
    "user_mentions": [
      {
        "id": "16408759",
        "id_str": "16408759",
        "indices": ["3", "10"],
        "name": "San Francisco International Airport (SFO) ✈️",
        "screen_name": "flySFO"
      }
    ]
  },
  "favorite_count": "0",
  "favorited": false,
  "full_text": "RT @flySFO: Nothing better than listening 2 an audio tour while waiting to board your flight. Introducing SFO's T2 iTunes Art Tour! http ...",
  "id": "116217783158714368",
  "id_str": "116217783158714368",
  "lang": "en",
  "retweet_count": "0",
  "retweeted": false,
  "source": "\u003ca href=\"https://about.twitter.com/products/tweetdeck\" rel=\"nofollow\"\u003eTweetDeck\u003c/a\u003e",
  "truncated": false
},
{
  "created_at": "Tue Sep 20 19:26:16 +0000 2011",
  "display_text_range": ["0", "140"],
  "entities": {
    "hashtags": [
      {
        "indices": ["76", "80"],
        "text": "SFO"
      }
    ],
    "symbols": [],
    "urls": [
      {
        "display_url": "fb.me/OdC3kRWA",
        "expanded_url": "http://fb.me/OdC3kRWA",
        "indices": ["120", "140"],
        "url": "http://t.co/6AOJbcIq"
      }
    ],
    "user_mentions": [
      {
        "id": "20019271",
        "id_str": "20019271",
        "indices": ["8", "16"],
        "name": "Jetwerk",
        "screen_name": "Jetwerk"
      }
    ]
  },
  "favorite_count": "0",
  "favorited": false,
  "full_text": "Thx! RT @Jetwerk: Have you checked out the TV in the Antenna Age exhibit in #SFO T3. Walk down memory lane. Very cool...http://t.co/6AOJbcIq",
  "id": "116231720801538048",
  "id_str": "116231720801538048",
  "lang": "en",
  "possibly_sensitive": false,
  "retweet_count": "0",
  "retweeted": false,
  "source": "\u003ca href=\"https://about.twitter.com/products/tweetdeck\" rel=\"nofollow\"\u003eTweetDeck\u003c/a\u003e",
  "truncated": false
}
...and so on
```

You can pass multiple `-query` parameters:

```
$> ./bin/emit \
	-json \
	-format-json \
	-query 'full_text=\bSFO\b' \
	-query 'full_text=\bJFK\b' \
	-tweets-uri file:///usr/local/twitter/data/tweet.js 

[{
  "created_at": "Wed Oct 03 18:53:02 +0000 2012",
  "display_text_range": ["0", "143"],
  "entities": {
    "hashtags": [],
    "symbols": [],
    "urls": [],
    "user_mentions": [
      {
        "id": "3042711",
        "id_str": "3042711",
        "indices": ["17", "28"],
        "name": "Mark Graham",
        "screen_name": "MarkGraham"
      }
    ]
  },
  "favorite_count": "0",
  "favorited": false,
  "full_text": "Woot! Thanks! RT @MarkGraham: Today is SFO-\u0026gt;JFK. The \"Deities in Stone\",Hindu Temple Architecture exhibit, at the United terminal is amazing",
  "id": "253568360359546881",
  "id_str": "253568360359546881",
  "lang": "en",
  "retweet_count": "0",
  "retweeted": false,
  "source": "\u003ca href=\"https://about.twitter.com/products/tweetdeck\" rel=\"nofollow\"\u003eTweetDeck\u003c/a\u003e",
  "truncated": false
}]
```

The default query mode is to ensure that all queries match but you can also specify that only one or more queries need to match by passing the `-query-mode ANY` flag:

```
$> ./bin/emit \
	-json \
	-format-json \
	-query 'full_text=\bSFO\b' \
	-query 'full_text=\bJFK\b' \
	-query-mode ANY \	       
	-tweets-uri file:///usr/local/twitter/data/tweet.js 

[{
  "created_at": "Wed May 25 00:20:46 +0000 2016",
  "display_text_range": ["0", "136"],
  "entities": {
    "hashtags": [
      {
        "indices": ["0", "10"],
        "text": "OnThisDay"
      }, 
      {
        "indices": ["37", "47"],
        "text": "Worldport"
      }, 
      {
        "indices": ["67", "71"],
        "text": "JFK"
      }, 
      {
        "indices": ["75", "79"],
        "text": "NYC"
      }
    ],
    "media": [
      {
        "display_url": "pic.twitter.com/F7H5ge7hX0",
        "expanded_url": "https://twitter.com/SFOMuseum/status/735264308389502976/photo/1",
        "id": "735264306720186368",
        "id_str": "735264306720186368",
        "indices": ["113", "136"],
        "media_url": "http://pbs.twimg.com/media/CjQu_coUoAAfGpb.jpg",
        "media_url_https": "https://pbs.twimg.com/media/CjQu_coUoAAfGpb.jpg",
        "sizes": {
          "large": {
            "h": "2000",
            "resize": "fit",
            "w": "1923"
          },
          "medium": {
            "h": "1200",
            "resize": "fit",
            "w": "1154"
          },
          "small": {
            "h": "680",
            "resize": "fit",
            "w": "654"
          },
          "thumb": {
            "h": "150",
            "resize": "crop",
            "w": "150"
          }
        },
        "type": "photo",
        "url": "https://t.co/F7H5ge7hX0"
      }
    ],
    "symbols": [],
    "urls": [],
    "user_mentions": []
  },
  "extended_entities": {
    "media": [
      {
        "display_url": "pic.twitter.com/F7H5ge7hX0",
        "expanded_url": "https://twitter.com/SFOMuseum/status/735264308389502976/photo/1",
        "id": "735264306720186368",
        "id_str": "735264306720186368",
        "indices": ["113", "136"],
        "media_url": "http://pbs.twimg.com/media/CjQu_coUoAAfGpb.jpg",
        "media_url_https": "https://pbs.twimg.com/media/CjQu_coUoAAfGpb.jpg",
        "sizes": {
          "large": {
            "h": "2000",
            "resize": "fit",
            "w": "1923"
          },
          "medium": {
            "h": "1200",
            "resize": "fit",
            "w": "1154"
          },
          "small": {
            "h": "680",
            "resize": "fit",
            "w": "654"
          },
          "thumb": {
            "h": "150",
            "resize": "crop",
            "w": "150"
          }
        },
        "type": "photo",
        "url": "https://t.co/F7H5ge7hX0"
      }, 
      {
        "display_url": "pic.twitter.com/F7H5ge7hX0",
        "expanded_url": "https://twitter.com/SFOMuseum/status/735264308389502976/photo/1",
        "id": "735264307491930117",
        "id_str": "735264307491930117",
        "indices": ["113", "136"],
        "media_url": "http://pbs.twimg.com/media/CjQu_fgUgAUmT2m.jpg",
        "media_url_https": "https://pbs.twimg.com/media/CjQu_fgUgAUmT2m.jpg",
        "sizes": {
          "large": {
            "h": "2000",
            "resize": "fit",
            "w": "1928"
          },
          "medium": {
            "h": "1200",
            "resize": "fit",
            "w": "1157"
          },
          "small": {
            "h": "680",
            "resize": "fit",
            "w": "656"
          },
          "thumb": {
            "h": "150",
            "resize": "crop",
            "w": "150"
          }
        },
        "type": "photo",
        "url": "https://t.co/F7H5ge7hX0"
      }
    ]
  },
  "favorite_count": "8",
  "favorited": false,
  "full_text": "#OnThisDay in 1960, the Pan American #Worldport terminal opened at #JFK in #NYC. Did you ever fly out Worldport? https://t.co/F7H5ge7hX0",
  "id": "735264308389502976",
  "id_str": "735264308389502976",
  "lang": "en",
  "possibly_sensitive": false,
  "retweet_count": "5",
  "retweeted": false,
  "source": "\u003ca href=\"http://twitter.com\" rel=\"nofollow\"\u003eTwitter Web Client\u003c/a\u003e",
  "truncated": false
}
...and so on
]
```

### pointers

Export pointers (for users and hashtags) from a `tweets.json` file (produced by the `trim` tool).
	
```
$> ./bin/pointers -h
Usage of ./bin/pointers:
  -hashtags
    	Export hash tags in tweets. (default true)
  -mentions
    	Export users mentioned in tweets. (default true)
  -trim-prefix
    	Trim default tweet.js JavaScript prefix. (default true)
  -tweets-uri tweets.js
    	A valid gocloud.dev/blob URI to your tweets.js file.
```

For example:

```
$> ./bin/pointers -tweets-uri file:///usr/local/twitter/data/tweet.js
property,value
user,genavnews
user,ladyfleur
user,JuanPDeAnda
user,sfo1977
user,787FirstClass
user,WNYCculture
user,Mairin_
user,ManuBarack
user,LMBernhard
user,OneBrownGirl
user,jtroll
user,spatrizi
user,patriciadenni20
...
tag,JimLund
tag,MoodLighting
tag,giants
tag,republicairlines
tag,T2
tag,eyecandy
tag,Wildenhain
```

### trim

Trim JavaScript boilerplate from a `tweets.js` file in order to make it valid JSON. Outputs to `STDOUT`.

```
./bin/trim -h
Usage of ./bin/trim:
  -tweets-uri tweets.js
    	A valid gocloud.dev/blob URI to your tweets.js file.
```	

For example:

```
$> ./bin/trim -tweets /usr/local/twitter/data/tweet.js > /usr/local/twitter/data/tweet.json
```

### Unshorten

Expand all the URLs in a `tweets.js` file. Outputs a JSON dictionary to `STDOUT`.

```
./bin/unshorten -h
Usage of ./bin/unshorten:
  -progress
    	Display progress information
  -qps int
    	Number of (unshortening) queries per second (default 10)
  -seed string
    	Pre-fill the unshortening cache with data in this file
  -timeout int
    	Maximum number of seconds of for an unshorterning request (default 30)
  -trim-prefix
    	Trim default tweet.js JavaScript prefix. (default true)
  -tweets-uri tweets.js
    	A valid gocloud.dev/blob URI to your tweets.js file.
```

For example:

```
$> ./bin/unshorten -progress -tweets-uri file:///usr/local/twitter/data/tweet.js
2020/10/06 11:37:02 Head "http://gowal.la/p/hHG2": dial tcp: lookup gowal.la: no such host
2020/10/06 11:37:08 Head "http://danceonmarket.com/events/": dial tcp: lookup danceonmarket.com: no such host
2020/10/06 11:37:11 3431 of 3461 URLs left to unshorten (from 9759 tweets)
...and so on
```

## See also

* https://github.com/sfomuseum/go-url-unshortener
* https://github.com/aaronland/go-json-query