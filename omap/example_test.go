package omap_test

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/creachadair/mds/omap"
)

const input = `
the thousand injuries of fortunato i had borne as i best could but when he
ventured upon insult i vowed revenge you who so well know the nature of my soul
will not suppose however that i gave utterance to a threat at length i would be
avenged this was a point definitely settled but the very definitiveness with
which it was resolved precluded the idea of risk i must not only punish but
punish with impunity a wrong is unredressed when retribution overtakes its
redresser it is equally unredressed when the avenger fails to make himself felt
as such to him who has done the wrong it must be understood that neither by
word nor deed had i given fortunato cause to doubt my good will i continued as
was my wont to smile in his face and he did not perceive that my smile now was
at the thought of his immolation
`

func ExampleMap() {
	// Construct a map on a naturally ordered key (string).
	m := omap.New[string, int]()
	for _, w := range strings.Fields(input) {
		m.Set(w, m.Get(w)+1)
	}

	// Construct a map with an explicit ordering function.
	c := omap.NewFunc[int, string](cmp.Compare)

	// Traverse a map in key order.
	for it := m.First(); it.IsValid(); it.Next() {
		c.Set(it.Value(), it.Key())
	}

	// Traverse a map in reverse order.
	for it := c.Last(); it.IsValid(); it.Prev() {
		fmt.Println(it.Value(), it.Key())
		if it.Key() <= 3 {
			break
		}
	}

	// Output:
	//
	// i 8
	// the 7
	// to 5
	// was 4
	// when 3
}
