package unshortener

import (
	"context"
	_ "log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Unshortener interface {
	Unshorten(context.Context, *url.URL) (*url.URL, error)
}

type ThrottledUnshortener struct {
	Unshortener
	throttle <-chan time.Time
	timeout  time.Duration
	client   *http.Client
}

type CachedUnshortener struct {
	Unshortener
	worker Unshortener
	cache  *sync.Map
}

func UnshortenString(ctx context.Context, sh Unshortener, str_u string) (*url.URL, error) {

	select {
	case <-ctx.Done():
		return nil, nil
	default:
		// pass
	}

	u, err := url.Parse(str_u)

	if err != nil {
		return nil, err
	}

	return sh.Unshorten(ctx, u)
}

func NewCachedUnshortener(worker Unshortener) (Unshortener, error) {

	seed := make(map[string]string)

	return NewCachedUnshortenerWithSeed(worker, seed)
}

func NewCachedUnshortenerWithSeed(worker Unshortener, seed map[string]string) (Unshortener, error) {

	cache := new(sync.Map)

	for k, v := range seed {
		cache.Store(k, v)
	}

	sh := CachedUnshortener{
		worker: worker,
		cache:  cache,
	}

	return &sh, nil
}

func (sh *CachedUnshortener) Unshorten(ctx context.Context, u *url.URL) (*url.URL, error) {

	select {
	case <-ctx.Done():
		return nil, nil
	default:
		// pass
	}

	str_url := u.String()

	v, ok := sh.cache.Load(str_url)

	if ok {
		str_url = v.(string)
		return url.Parse(str_url)
	}

	u2, err := sh.worker.Unshorten(ctx, u)

	if err != nil {
		return nil, err
	}

	sh.cache.Store(u.String(), u2.String())
	return u2, nil
}

func NewThrottledUnshortener(rate time.Duration, timeout time.Duration) (Unshortener, error) {

	throttle := time.Tick(rate)

	client := &http.Client{
		// something something something client.CheckRedirect - configure for more than (default number of) hops?
		// https://stackoverflow.com/questions/23297520/how-can-i-make-the-go-http-client-not-follow-redirects-automatically
		// https://jonathanmh.com/tracing-preventing-http-redirects-golang/
	}

	sh := ThrottledUnshortener{
		throttle: throttle,
		timeout:  timeout,
		client:   client,
	}

	return &sh, nil
}

func (sh *ThrottledUnshortener) Unshorten(ctx context.Context, u *url.URL) (*url.URL, error) {

	/*
		t1 := time.Now()
		var t2 time.Time

		defer func() {
			log.Printf("TIME TO FETCH %s %v (%v)\n", u.String(), time.Since(t2), time.Since(t1))
		}()
	*/

	<-sh.throttle

	/*
		t2 = time.Now()
	*/

	select {
	case <-ctx.Done():
		return nil, nil
	default:
		// pass
	}

	req_ctx, cancel := context.WithTimeout(ctx, sh.timeout)
	defer cancel()

	req, err := http.NewRequest(http.MethodHead, u.String(), nil)

	if err != nil {
		return nil, err
	}

	rsp, err := sh.client.Do(req.WithContext(req_ctx))

	if err != nil {
		return nil, err
	}

	return rsp.Request.URL, nil
}
