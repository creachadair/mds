package mdiff_test

import (
	"os"
	"time"

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

func ExampleFormatContext() {
	diff := mdiff.New(
		[]string{"I", "saw", "three", "mice", "running", "away"},
		[]string{"three", "blind", "mice", "ran", "home"},
	).AddContext(3).Unify()

	ts := time.Date(2024, 3, 18, 22, 30, 35, 0, time.UTC)
	mdiff.FormatContext(os.Stdout, diff, &mdiff.FileInfo{
		Left: "old", LeftTime: ts,
		Right: "new", RightTime: ts.Add(3 * time.Second),
		TimeFormat: time.ANSIC,
	})

	// Output:
	//
	// *** old	Mon Mar 18 22:30:35 2024
	// --- new	Mon Mar 18 22:30:38 2024
	// ***************
	// *** 1,6 ****
	// - I
	// - saw
	//   three
	//   mice
	// ! running
	// ! away
	// --- 1,5 ----
	//   three
	// + blind
	//   mice
	// ! ran
	// ! home

}

func ExampleFormatUnified() {
	diff := mdiff.New(
		[]string{"I", "saw", "three", "mice", "running", "away"},
		[]string{"three", "blind", "mice", "ran", "home"},
	).AddContext(3).Unify()

	mdiff.FormatUnified(os.Stdout, diff, nil) // nil means "no header"

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
