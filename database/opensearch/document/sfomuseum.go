package document

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sfomuseum/go-sfomuseum/instagram/media"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	es_document "github.com/whosonfirst/go-whosonfirst-elasticsearch/document"
)

func SFOMuseumPrepareDocumentFunc() es_document.PrepareDocumentFunc {

	sfom_f := func(ctx context.Context, body []byte) ([]byte, error) {

		var err error

		id_rsp := gjson.GetBytes(body, "wof:id")
		str_id := id_rsp.String()

		im_rsp := gjson.GetBytes(body, "millsfield:images")

		if im_rsp.Exists() {

			count := len(im_rsp.Array())

			body, err = sjson.SetBytes(body, "millsfield:count_images", count)

			if err != nil {
				return nil, fmt.Errorf("Failed to assign millsfield:count_images, %w", err)
			}
		}

		// Instagram stuff
		// tl;dr is "convert IG's goofy datetime strings in RFC3339 so that Elasticsearch isn't sad"
		// See also: sfomuseum/go-sfomuseum-instagram and sfomuseum/go-sfomuseum-instagram-publish

		ig_rsp := gjson.GetBytes(body, "instagram:post")

		if ig_rsp.Exists() {

			taken_rsp := gjson.GetBytes(body, "instagram:post.taken_at")

			t, err := media.ParseTime(taken_rsp.String())

			if err != nil {
				return nil, fmt.Errorf("Failed to parse '%s', %w", taken_rsp.String(), err)
			}

			body, err = sjson.SetBytes(body, "instagram:post.taken_at", t.Format(time.RFC3339))

			if err != nil {
				return nil, err
			}

			tags_rsp := gjson.GetBytes(body, "instagram:post.caption.hashtags")

			if tags_rsp.Exists() {

				hashtags := make([]string, 0)

				for _, t := range tags_rsp.Array() {
					hashtags = append(hashtags, strings.ToLower(t.String()))
				}

				body, err = sjson.SetBytes(body, "instagram:post.caption.hashtags", hashtags)

				if err != nil {
					return nil, fmt.Errorf("Failed to update IG hash tags, %w", err)
				}

			}

		}

		sfom_dates := []string{
			"sfomuseum:date_start",
			"sfomuseum:date_end",
		}

		for _, path := range sfom_dates {

			dt_rsp := gjson.GetBytes(body, path)

			if !dt_rsp.Exists() {
				continue
			}

			if dt_rsp.String() == "" {

				body, err = sjson.DeleteBytes(body, path)

				if err != nil {
					return nil, fmt.Errorf("Failed to delete '%s' for %s", path, str_id)
				}
			}
		}

		return body, nil
	}

	return sfom_f
}
