package mdiff_test

import (
	"os"

	"github.com/creachadair/mds/mdiff"
)

func ExampleFormat() {
	diff := mdiff.New(
		[]string{"I", "saw", "three", "mice", "running", "away"},
		[]string{"three", "blind", "mice", "ran", "home"},
	)

	mdiff.Format(os.Stdout, diff, nil)

	// Output:
	//
	// 1,2d0
	// < I
	// < saw
	// 3a2
	// > blind
	// 5,6c4,5
	// < running
	// < away
	// ---
	// > ran
	// > home
}

func ExampleFormatUnified() {
	diff := mdiff.New(
		[]string{"I", "saw", "three", "mice", "running", "away"},
		[]string{"three", "blind", "mice", "ran", "home"},
	).AddContext(3).Unify()

	mdiff.FormatUnified(os.Stdout, diff, mdiff.NoHeader)

	// Output:
	//
	// @@ -1,6 +1,5 @@
	// -I
	// -saw
	//  three
	// +blind
	//  mice
	// -running
	// -away
	// +ran
	// +home
}
