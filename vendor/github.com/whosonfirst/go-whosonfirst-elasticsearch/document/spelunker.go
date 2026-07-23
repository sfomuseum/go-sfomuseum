package document

import (
	"context"
	"log/slog"

	sp_document "github.com/whosonfirst/go-whosonfirst-spelunker/document"
)

// PrepareSpelunkerV1Document prepares a Who's On First document for indexing with the
// "v1" Elasticsearch (v2.x) schema. For details please consult:
// https://github.com/whosonfirst/es-whosonfirst-schema/tree/master/schema/2.4
func PrepareSpelunkerV1Document(ctx context.Context, body []byte) ([]byte, error) {

	slog.Warn("The whosonfirst/go-whosonfirst-elasticsearch.PrepareSpelunkerV1Document method is deprecated. Please use whosonfirst/go-whosonfirst-spelunker/document.PrepareSpelunkerV1Document instead")
	return sp_document.PrepareSpelunkerV1Document(ctx, body)
}

// AppendSpelunkerV1Properties appends properties specific to the v1" Elasticsearch (v2.x) schema
// to a Who's On First document for. For details please consult:
// https://github.com/whosonfirst/es-whosonfirst-schema/tree/master/schema/2.4
func AppendSpelunkerV1Properties(ctx context.Context, body []byte) ([]byte, error) {

	slog.Warn("The whosonfirst/go-whosonfirst-elasticsearch.AppendSpelunkerV1Properties method is deprecated. Please use whosonfirst/go-whosonfirst-spelunker/document.AppendSpelunkerV1Properties instead")
	return sp_document.AppendSpelunkerV1Properties(ctx, body)
}
