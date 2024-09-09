package mdiff_test

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/creachadair/mds/mdiff"
)

func ExampleNormal() {
	diff := mdiff.New(
		[]string{"I", "saw", "three", "mice", "running", "away"},
		[]string{"three", "blind", "mice", "ran", "home"},
	)

	diff.Format(os.Stdout, mdiff.Normal, nil)

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

func ExampleContext() {
	diff := mdiff.New(
		[]string{"I", "saw", "three", "mice", "running", "away"},
		[]string{"three", "blind", "mice", "ran", "home"},
	).AddContext(3).Unify()

	ts := time.Date(2024, 3, 18, 22, 30, 35, 0, time.UTC)
	diff.Format(os.Stdout, mdiff.Context, &mdiff.FileInfo{
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

func ExampleUnified() {
	diff := mdiff.New(
		[]string{"I", "saw", "three", "mice", "running", "away"},
		[]string{"three", "blind", "mice", "ran", "home"},
	).AddContext(3).Unify()

	diff.Format(os.Stdout, mdiff.Unified, nil) // nil means "no header"

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

func ExampleRead() {
	const textDiff = `1,2d0
< I
< saw
3a2
> blind
5,6c4,5
< running
< away
---
> ran
> home`

	p, err := mdiff.Read(strings.NewReader(textDiff))
	if err != nil {
		log.Fatalf("Read: %v", err)
	}
	printChunks(p.Chunks)

	// Output:
	//
	// Chunk 1: left 1:3, right 1:1
	//  edit 1.1: -[I saw]
	// Chunk 2: left 4:4, right 2:3
	//  edit 2.1: +[blind]
	// Chunk 3: left 5:7, right 4:6
	//  edit 3.1: ![running away:ran home]
}

func ExampleReadUnified() {
	const textDiff = `@@ -1,3 +1 @@
-I
-saw
 three
@@ -3,2 +1,3 @@
 three
+blind
 mice
@@ -4,3 +3,3 @@
 mice
-running
-away
+ran
+home`

	p, err := mdiff.ReadUnified(strings.NewReader(textDiff))
	if err != nil {
		log.Fatalf("ReadUnified: %v", err)
	}
	printChunks(p.Chunks)

	// Output:
	//
	// Chunk 1: left 1:4, right 1:1
	//  edit 1.1: -[I saw]
	//  edit 1.2: =[three]
	// Chunk 2: left 3:5, right 1:4
	//  edit 2.1: =[three]
	//  edit 2.2: +[blind]
	//  edit 2.3: =[mice]
	// Chunk 3: left 4:7, right 3:6
	//  edit 3.1: =[mice]
	//  edit 3.2: -[running away]
	//  edit 3.3: +[ran home]
}

func printChunks(cs []*mdiff.Chunk) {
	for i, c := range cs {
		fmt.Printf("Chunk %d: left %d:%d, right %d:%d\n",
			i+1, c.LStart, c.LEnd, c.RStart, c.REnd)
		for j, e := range c.Edits {
			fmt.Printf(" edit %d.%d: %v\n", i+1, j+1, e)
		}
	}
}
