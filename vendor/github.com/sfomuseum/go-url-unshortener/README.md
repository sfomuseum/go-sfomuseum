# go-url-unshortener

Tools for resolving URLs that may have been "shortened".

## Install

You will need to have both `Go` (specifically [version 1.12](https://golang.org/dl/) or higher because we're using [Go modules](https://github.com/golang/go/wiki/Modules)) and the `make` programs installed on your computer. Assuming you do just type:

```
make tools
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Example

```
import (
       "context"
       "flag"
	"github.com/sfomuseum/go-url-unshortener"       
       "log"
       "time"
)

func main() {

	flag.Parse()

	timeout := time.Second / 30
	rate := time.Second / 10
	
	worker, err := unshortener.NewThrottledUnshortener(rate, timeout)

	if err != nil {
		log.Fatal(err)
	}

	cache, err := unshortener.NewCachedUnshortener(worker)

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, str_url := range flag.Args() {

		u, err := unshortener.UnshortenString(ctx, cache, str_url)

		if err != nil {
			log.Fatal(err)
		}

		log.Println(u.String())
	}
	
}	
```

## Interfaces

### Unshortener

```
type Unshortener interface {
	Unshorten(context.Context, *url.URL) (*url.URL, error)
}
```

## Tools

### unshorten

Unshorten one or more URLs from the command-line (or STDIN).

```
Usage of ./bin/unshorten:
  -progress
	Display progress information
  -qps int
       Number of (unshortening) queries per second (default 10)
  -seed string
    	Pre-fill the unshortening cache with data in this file
  -stdin
	Read URLs from STDIN
  -timeout int
    	   Maximum number of seconds of for an unshorterning request (default 30)
  -verbose
	Be chatty about what's going on
```

For example, let's say you wanted to unshorted all the `expanded_urls` URLs in a Twitter users `tweet.js` export file:

```
$> grep expanded_url /path/to/tweet.js \
	| awk '{ print $3 }' | sort | uniq | sed 's/^"//' | sed 's/"\,$//' \
	| ./bin/unshorten -stdin -verbose \
	| jq

2019/03/01 11:57:47 Head http://airwaysnews.com/blog/2016/06/09/wow-air-kicks-off-san-francisco-service/%22,: dial tcp: lookup airwaysnews.com: no such host
2019/03/01 11:57:48 http://4sq.com/RfgMa", becomes http://4sq.com/RfgMa%22,
2019/03/01 11:57:48 Head http://CNN.com",: dial tcp: lookup CNN.com",: no such host
2019/03/01 11:57:49 Head http://MahoBeachCam.com",: dial tcp: lookup MahoBeachCam.com",: no such host
2019/03/01 11:57:49 http://500px.com/photo/4751280", becomes https://500px.com:443/photo/4751280%22,
2019/03/01 11:57:49 http://1.usa.gov/1LbxdUe", becomes https://www.transportation.gov/fastlane/women-in-aviation-connect-engage-inspire
2019/03/01 11:57:49 Head http://airwaysnews.com/blog/2015/08/14/american-airlines-to-launch-new-trial-on-uniforms/%22,: dial tcp: lookup airwaysnews.com: no such host
2019/03/01 11:57:49 Head http://airwaysnews.com/blog/2016/04/01/airline-industry-announcements-for-april-1/%22,: dial tcp: lookup airwaysnews.com: no such host
2019/03/01 11:57:49 Head http://airwaysnews.com/blog/2016/04/27/museum-of-flight-completes-final-boeing-247d-flight/%22,: dial tcp: lookup airwaysnews.com: no such host
2019/03/01 11:57:50 http://bit.ly/1rbWi7G", becomes https://www.flysfo.com/museum/aviation-museum-library/collection?field_type_collection_tid_1=1027
2019/03/01 11:57:50 http://bit.ly/1950sConsumer", becomes http://bit.ly/1950sConsumer%22,
2019/03/01 11:57:50 http://bit.ly/16vo1lU", becomes https://www.flysfo.com/museum/exhibitions/souvenirs-tokens-travel07.html
2019/03/01 11:57:51 http://bit.ly/1NvfDZt", becomes https://www.flysfo.com/museum/exhibitions/classic-monsters-kirk-hammett-collection
2019/03/01 11:57:52 http://bit.ly/1TPvJ29", becomes https://www.flysfo.com/museum/about/employment
2019/03/01 11:57:52 http://bit.ly/1RVIYKt", becomes https://www.flysfo.com/museum/public-art-collection?nid=3292
2019/03/01 11:57:52 http://bit.ly/1U7sdhL", becomes https://www.flysfo.com/museum/aviation-museum-library/collection?field_type_collection_tid_1=1025
2019/03/01 11:57:52 http://1.usa.gov/KmedO3", becomes https://www.nga.gov/404status.html
2019/03/01 11:57:52 http://bit.ly/1WLCpRH", becomes https://www.flysfo.com/museum/aviation-museum-library/collection/10319
time passes...
{
  "http://1.usa.gov/1LbxdUe\",": "https://www.transportation.gov/fastlane/women-in-aviation-connect-engage-inspire",
  "http://1.usa.gov/KmedO3\",": "https://www.nga.gov/404status.html",
  "http://4sq.com/15C3nS9\",": "https://www.swarmapp.com/sewsueme/checkin/5175d467e4b00b60396d301f?s=ZUQsGNttwY1X92Fo9zyyqpIiosc&ref=tw",
  "http://4sq.com/19iEU6u\",": "https://www.swarmapp.com/user/2832802/checkin/51f03a77498e999f9b58b44c?s=nXhx2Qx9kBxeWJuU2ox9cSptX6M&ref=tw",
  "http://4sq.com/1hlKD0H\",": "https://www.swarmapp.com/alli_burnie/checkin/53921a8f498eb18ae1c920ea?s=u804xRsf3TFZVBKpusk8erpBnMk&ref=tw",
  "http://4sq.com/LZ6065\",": "https://www.swarmapp.com/markus64/checkin/4fde3986e4b0d087148b9fa1?s=iGxZKLJkCW3q31WNgDyewEV9dRI&ref=tw",
  "http://4sq.com/RfgMa\",": "http://4sq.com/RfgMa%22,",
  and so on...
  
  "http://sfomuseum.org": "https://www.flysfo.com/museum",
  "http://sfomuseum.tumblr.com": "-",
  "http://sfomuseum.tumblr.com/": "-",
  "http://sfomuseum.tumblr.com/post/129239878186/today-is-askacurator-day-on-twitter-and-we": "http://sfomuseum.tumblr.com/post/129239878186/today-was-askacurator-day-on-twitter-and-we#_=_",
  "http://sfomuseum.tumblr.com/post/129248260051/ask": "-",
  "http://sfomuseum.tumblr.com/post/130087062136/the-san-francisco-49ers-first-mascot-was-a-mule": "-",
  "http://sfomuseum.tumblr.com/post/132123520141/bonnie-jones-moon-worked-for-pan-american-world": "-",
  "http://sfomuseum.tumblr.com/post/132218459498/friday-november-6-600-pm-in-the-aviation": "-",
  "http://sfomuseum.tumblr.com/post/132416845556/sfo-museum-based-at-san-francisco-international": "-",
  "http://sfomuseum.tumblr.com/post/133504040926/ida-staggers-was-first-hired-by-twa": "-",
  and so on...
  
  "http://www.youtube.com/watch?v=vPaqRmByXo4": "https://www.youtube.com/watch?v=vPaqRmByXo4",
  "http://yfrog.com/es4vbedj": "http://imageshack.com/lost",
  "http://yfrog.com/es554ccj": "http://imageshack.com/lost",
  "http://yfrog.com/esg8bfwj": "http://imageshack.com/lost",
  "http://yfrog.com/esj53hj": "http://imageshack.com/lost",
  "http://yfrog.com/esm4gvvj": "http://imageshack.com/lost",
  "http://yfrog.com/esq63gkj": "http://imageshack.com/lost",
  "http://yfrog.com/g09u1okj": "http://imageshack.com/lost",
  "http://yfrog.com/gy2hiosj": "http://imageshack.com/lost",
  "http://yfrog.com/gy486zkj": "?",
  "http://yfrog.com/gy8x5drhj": "?",
  "http://yfrog.com/gyc8yqbwj": "?",
  and so on...  
}
```

Unshortened URLs that are the same as their input are encoded as `"-"`. URLs that were unable to be unshortened, for whatever reason, are encoded as `"?"`.

You can also use the output of `unshorten` to pre-seed lookups for subsequent invocations. For example, to start you might do this:

```
$> grep expanded_url /path/to/tweet.js \
	| awk '{ print $3 }' | sort | uniq | sed 's/^"//' | sed 's/"\,$//' \
	| ./bin/unshorten -stdin \
	> report-1.txt
```

Time will pass and unshortened URLs will be stored as a JSON-encoded dictionary in a file called `report-1.json`.

At some later date you can run the same command but using `report-1.json` as a lookup table, using the `-seed report-1.txt` flag, writing the results to a second `report-2.json` file:

```
$> grep expanded_url /path/to/tweet.js \
	| awk '{ print $3 }' | sort | uniq | sed 's/^"//' | sed 's/"\,$//' \
	| ./bin/unshorten -stdin -seed report-1.txt \
	> report-2.txt
```

Considerably less time will pass. This example assumes that `tweet.js` has not changed so both the `report-(n).json` files are the same:

```
$> ls -la report*.json
-rw-r--r--. 1 user domain users 576949 Mar  1 14:24 report-1.json
-rw-r--r--. 1 user domain users 576949 Mar  1 14:40 report-2.json
```
