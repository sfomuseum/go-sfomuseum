package document

import (
	"context"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	es_document "github.com/whosonfirst/go-whosonfirst-elasticsearch/document"
)

// SFOMuseumPrepareEmbeddingsDocumentFunc returns a `es_document.PrepareDocumentFunc` function
// that will yield a document for indexing using OpenSeach-style "semantic search" (embeddings).
// Note: As of this writing there is a hardcoded list of fields to read and a single hardcoded field
// in to which those data are stored. This is not ideal and it would be better to have something with
// sensible defaults that could be overridden. That does not exist today.
func SFOMuseumPrepareEmbeddingsDocumentFunc() es_document.PrepareDocumentFunc {

	sfom_f := func(ctx context.Context, body []byte) ([]byte, error) {

		new_body := []byte(`{}`)

		to_consider := []string{
			"wof:name",
			"sfomuseum:description",
		}

		to_index := make([]string, 0)

		for _, path := range to_consider {

			rsp := gjson.GetBytes(body, path)

			if !rsp.Exists() {
				continue
			}

			to_index = append(to_index, rsp.String())
		}

		str_body := strings.Join(to_index, " ")

		new_body, err := sjson.SetBytes(new_body, "search", []byte(str_body))

		if err != nil {
			return nil, fmt.Errorf("Failed to assign sfomuseum:description, %w", err)
		}

		return new_body, nil
	}

	return sfom_f
}
