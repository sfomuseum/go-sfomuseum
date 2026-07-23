// package secret provides methods for generating secrets associated with SFO Museum Instagram photos.
package secret

import (
	"crypto/md5"
	"fmt"
)

// DeriveSecret will produce a secret key for Instagram photos derived from 'id'
// See also: flysfo/www-sfomuseum-millsfield/www/include/lib_sfomuseum_instagram.php
func DeriveSecret(id string) string {
	rev_id := reverse(id)
	hash_id := hash(rev_id)
	secret := trim(hash_id)
	return secret
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func hash(s string) string {
	data := []byte(s)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func trim(s string) string {
	return s[0:10]
}
