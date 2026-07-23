package publish

import (
	"fmt"

	"github.com/sfomuseum/go-sfomuseum/instagram/media"
	"github.com/tidwall/gjson"
)

// DeriveMediaId will derive a (hopefully) persistent SFO Museum specific media ID
// from JSON properties in 'body'. This might be a `go-sfomuseum-instagram/media.Photo`
// instance or a WOF-style SFO Museum record.
func DeriveMediaId(body []byte, prefix string) (string, error) {

	path_taken := "taken_at"
	path_phash := "perceptual_hash"
	path_fhash := "file_hash"

	if prefix != "" {
		path_taken = fmt.Sprintf("%s.%s", prefix, path_taken)
		path_phash = fmt.Sprintf("%s.%s", prefix, path_phash)
		path_fhash = fmt.Sprintf("%s.%s", prefix, path_fhash)
	}

	taken_rsp := gjson.GetBytes(body, path_taken)

	if !taken_rsp.Exists() {
		return "", fmt.Errorf("Missing '%s' property", path_taken)
	}

	// START OF ok, see this?
	// This is in case IG decides to change how datetime strings are formatted
	// again. If they do we will update the layout passed to time.Parse but
	// continue to use that (now old) layout to format strings. Good times...

	t, err := media.ParseTime(taken_rsp.String())

	if err != nil {
		return "", fmt.Errorf("Failed to parse %s, %w", taken_rsp.String(), err)
	}

	taken_at := t.Format(media.TIME_FORMAT)

	// END OF ok, see this?

	hash_rsp := gjson.GetBytes(body, path_phash)

	if !hash_rsp.Exists() {
		hash_rsp = gjson.GetBytes(body, path_fhash)
	}

	if !hash_rsp.Exists() {
		return "", fmt.Errorf("Missing both '%s' and '%s' property", path_phash, path_fhash)

	}

	hash := hash_rsp.String()

	media_id := fmt.Sprintf("%s %s", taken_at, hash)

	// log.Println("Derive", media_id)
	return media.DeriveMediaIdFromString(media_id), nil
}
