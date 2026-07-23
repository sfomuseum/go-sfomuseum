package twitter

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"path/filepath"
	"log"

	"gocloud.dev/blob"	
)

const JAVASCRIPT_PREFIX string = "window.YTD.tweet.part0 = "

type OpenTweetsOptions struct {
	TrimPrefix bool
}

func OpenTweets(ctx context.Context, tweets_uri string, opts *OpenTweetsOptions) (io.ReadCloser, error) {

	tweets_fname := filepath.Base(tweets_uri)
	tweets_root := filepath.Dir(tweets_uri)

	log.Println(tweets_root)
	log.Println(tweets_fname)
	
	tweets_bucket, err := blob.OpenBucket(ctx, tweets_root)

	if err != nil {
		return nil, err
	}

	var tweets_fh io.ReadCloser

	fh, err := tweets_bucket.NewReader(ctx, tweets_fname, nil)

	if err != nil {
		return nil, err
	}

	tweets_fh = fh

	if opts != nil && opts.TrimPrefix {

		trimmed_fh, err := trimJavaScriptPrefix(ctx, tweets_fh)

		tweets_fh.Close()

		if err != nil {
			return nil, err
		}

		tweets_fh = trimmed_fh
	}

	return tweets_fh, nil
}

func trimJavaScriptPrefix(ctx context.Context, fh io.Reader) (io.ReadCloser, error) {
	return trimPrefix(ctx, fh, JAVASCRIPT_PREFIX)
}

func trimPrefix(ctx context.Context, fh io.Reader, prefix string) (io.ReadCloser, error) {

	offset := int64(len(prefix))
	whence := 0

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return nil, err
	}

	br := bytes.NewReader(body)

	_, err = br.Seek(offset, whence)

	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(br), nil

	// var buf bytes.Buffer
	// tee := io.TeeReader(fh, &buf)
	// return tee, nil
}
