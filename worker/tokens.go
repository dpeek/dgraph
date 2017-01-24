package worker

import (
	"bytes"

	"github.com/boltdb/bolt"
	"github.com/dgraph-io/dgraph/posting"
	"github.com/dgraph-io/dgraph/x"
)

func getTokens(funcArgs []string) ([]string, error) {
	if len(funcArgs) != 2 {
		return nil, x.Errorf("Function requires 2 arguments, but got %d",
			len(funcArgs))
	}
	return posting.DefaultIndexKeys(funcArgs[1])
}

// getInequalityTokens gets tokens geq / leq compared to given token.
func getInequalityTokens(attr, ineqValueToken string, f string) ([]string, error) {
	var out []string
	pstore.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("data")).Cursor()
		k, v := c.Seek(x.IndexKey(attr, ineqValueToken))

		hit := v != nil && len(v) > 0
		if f == "eq" {
			if hit {
				out = []string{ineqValueToken}
				return nil
			}
			return nil
		}

		if hit {
			out = []string{ineqValueToken}
		}

		indexPrefix := x.ParsedKey{Attr: attr}.IndexPrefix()
		isGeqOrGt := f == "geq" || f == "gt"

		for {
			if isGeqOrGt {
				c.Next()
			} else {
				c.Prev()
			}
			if k == nil || !bytes.HasPrefix(k, indexPrefix) {
				break
			}

			pk := x.Parse(k)
			x.AssertTrue(pk != nil)
			out = append(out, pk.Term)
		}
		return nil
	})
	return out, nil
}
