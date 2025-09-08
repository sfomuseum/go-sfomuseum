package campus

import (
	"context"
	"fmt"
	"io"
	_ "log/slog"
	"strings"

	"github.com/whosonfirst/go-reader/v2"
)

type Element interface {
	Id() int64
	AltId() string
	Placetype() string
	Walk(context.Context, ElementCallbackFunc) error
	AsTree(context.Context, reader.Reader, io.Writer, int) error
}

type ElementCallbackFunc func(context.Context, Element) error

func elementTree(ctx context.Context, el Element, r reader.Reader, wr io.Writer, indent int) error {

	fmt.Fprintf(wr, "%s%s\n", strings.Repeat("\t", indent), treeLabel(ctx, r, el))

	cb := func(ctx context.Context, other_el Element) error {
		return other_el.AsTree(ctx, r, wr, indent+1)
	}

	return el.Walk(ctx, cb)
}
