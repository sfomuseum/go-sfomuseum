package document

import (
	"context"
	"errors"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"	
)

func AppendCreatedAtTimestamp(ctx context.Context, body []byte) ([]byte, error) {

	created_rsp := gjson.GetBytes(body, "created_at")

	if !created_rsp.Exists() {
		return nil, errors.New("Missing created_at property")
	}

	str_created := created_rsp.String()

	// twitter: Mon Feb 13 18:49:22 +0000 2017
	// go: Mon Jan 2 15:04:05 -0700 MST 2006

	t_fmt := "Mon Jan 2 15:04:05 -0700 2006"
	t, err := time.Parse(t_fmt, str_created)

	if err != nil {
		return nil, err
	}

	body, err = sjson.SetBytes(body, "created", t.Unix())

	if err != nil {
		return nil, err
	}

	return body, nil
}
