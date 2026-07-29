package campus

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"runtime"
	
	_ "github.com/mattn/go-sqlite3"

	sfom_sql "github.com/sfomuseum/go-database/sql"
	"github.com/sfomuseum/go-edtf"
	"github.com/tidwall/gjson"
	wof_indexer "github.com/whosonfirst/go-whosonfirst/v4/database/sql/indexer"
	wof_tables "github.com/whosonfirst/go-whosonfirst/v4/database/sql/tables"
)

var WARN_IS_CURRENT = true

func NewDatabaseWithIterator(ctx context.Context, dsn string, iterator_uri string, paths ...string) (*sql.DB, error) {

	db_query := url.Values{}
	db_query.Set("dsn", dsn)

	db_uri := url.URL{}
	db_uri.Scheme = "sql"
	db_uri.Host = "sqlite3"

	db_uri.RawQuery = db_query.Encode()

	db, err := sfom_sql.OpenWithURI(ctx, db_uri.String())

	to_index := make([]sfom_sql.Table, 0)

	geojson_opts, err := wof_tables.DefaultGeoJSONTableOptions()

	if err != nil {
		return nil, err
	}

	geojson_opts.IndexAltFiles = false

	geojson_table, err := wof_tables.NewGeoJSONTableWithDatabaseAndOptions(ctx, db, geojson_opts)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, geojson_table)

	supersedes_table, err := wof_tables.NewSupersedesTableWithDatabase(ctx, db)

	if err != nil {
		return nil, fmt.Errorf("Failed to create supersedes table, %w", err)
	}

	to_index = append(to_index, supersedes_table)

	spr_opts, err := wof_tables.DefaultSPRTableOptions()

	if err != nil {
		return nil, fmt.Errorf("Failed to create spr table options, %w", err)
	}

	spr_table, err := wof_tables.NewSPRTableWithDatabaseAndOptions(ctx, db, spr_opts)

	if err != nil {
		return nil, fmt.Errorf("Failed to create spr table, %w", err)
	}

	to_index = append(to_index, spr_table)

	record_opts := &wof_indexer.LoadRecordFuncOptions{
		StrictAltFiles: false,
	}

	record_func := wof_indexer.LoadRecordFunc(record_opts)

	idx_opts := &wof_indexer.IndexerOptions{
		DB:             db,
		Tables:         to_index,
		LoadRecordFunc: record_func,
		Workers: runtime.NumCPU(),
	}

	idx, err := wof_indexer.NewIndexer(idx_opts)

	if err != nil {
		return nil, fmt.Errorf("Failed to create database indexer, %w", err)
	}

	err = idx.IndexURIs(ctx, iterator_uri, paths...)

	if err != nil {
		return nil, fmt.Errorf("Failed to index paths, %w", err)
	}

	return db, nil
}

func findChildIDs(ctx context.Context, db *sql.DB, parent_id int64, placetype string) ([]int64, error) {

	q := `SELECT s.id FROM spr s, geojson g WHERE s.id=g.id AND s.parent_id=? AND JSON_EXTRACT(g.body, '$.properties."sfomuseum:placetype"')=?`

	slog.Debug(q, "parent_id", parent_id, "placetype", placetype)

	rows, err := db.QueryContext(ctx, q, parent_id, placetype)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	children := make([]int64, 0)

	for rows.Next() {

		var superseded_by int64
		err := rows.Scan(&superseded_by)

		if err != nil {
			return nil, err
		}

		children = append(children, superseded_by)
	}

	err = rows.Close()

	if err != nil {
		return nil, err
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return children, nil
}

func loadFeatureWithDBAndChecks(ctx context.Context, db *sql.DB, id int64) ([]byte, error) {

	body, err := loadFeatureWithDB(ctx, db, id)

	if err != nil {
		return nil, fmt.Errorf("Failed to load feature for record %d, %w", id, err)
	}

	name_rsp := gjson.GetBytes(body, "properties.wof:name")
	inception_rsp := gjson.GetBytes(body, "properties.edtf:inception")
	cessation_rsp := gjson.GetBytes(body, "properties.edtf:cessation")

	deprecated_rsp := gjson.GetBytes(body, "properties.edtf:deprecated")

	if deprecated_rsp.Exists() && deprecated_rsp.String() != "" {
		return nil, nil
	}

	current_rsp := gjson.GetBytes(body, "properties.mz:is_current")

	if !current_rsp.Exists() {
		return nil, fmt.Errorf("Missing properties.mz:is_current property for record %d", id)
	}

	if current_rsp.Int() != 1 && WARN_IS_CURRENT {

		cessation_str := cessation_rsp.String()

		if cessation_str == "" || cessation_str == edtf.OPEN {
			slog.Warn("Unexpected mz:is_current property", "id", id, "mz:is_current", current_rsp.Int(), "name", name_rsp.String(), "inception", inception_rsp.String(), "cessation", cessation_rsp.String())
		}

		// return nil, nil
	}

	return body, nil
}

func loadFeatureWithDB(ctx context.Context, db *sql.DB, id int64) ([]byte, error) {

	q := "SELECT body FROM geojson WHERE id = ?"

	row := db.QueryRowContext(ctx, q, id)

	var body string
	err := row.Scan(&body)

	if err != nil {
		return nil, err
	}

	return []byte(body), nil
}
