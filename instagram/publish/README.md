# go-sfomuseum-instagram-publish

Tools for working with Instagram exports in a SFO Museum context.

## Documentation

Documentation is incomplete at this time.

## Tools

### publish

```
$> ./bin/publish \
	-media-bucket-uri file:///usr/local/data/instagram/instagram-sfomuseum-2024-11-27-p55zxMWB \
	file:///usr/local/data/instagram/instagram-sfomuseum-2024-11-27-p55zxMWB/media.json
```

Note: The `media.json` file in the example above is _NOT_ included with Instagram exports by default. You will need to create it manually using the `derive-media-json` tool in the [sfomuseum/go-sfomuseum-instagram](https://github.com/sfomuseum/go-sfomuseum-instagram) package. For example:

```
$> cd /usr/local/go-sfomuseum-instagram
$> make cli

$> ./bin/derive-media-json \
	/usr/local/data/instagram/instagram-sfomuseum-2024-11-27-p55zxMWB/your_instagram_activity/content/posts_1.html \
	> /usr/local/data/instagram/instagram-sfomuseum-2024-11-27-p55zxMWB/media.json

```

## See also

* https://github.com/sfomuseum/go-sfomuseum-instagram