package writer

import (
	"context"

	"github.com/whosonfirst/go-whosonfirst/v4/database/opensearch/document"
)

// DocumentWriter is an interface to provide methods not offered by the `whosonfirst/go-writer.Writer` interface.
type DocumentWriter interface {
	// AppendPrepareFunc will append 'fn' to the list of `go-whosonfirst-elasticsearch/document.PrepareDocumentFunc` functions
	// to be applied to each document written.
	AppendPrepareFunc(context.Context, document.PrepareDocumentFunc) error
}
