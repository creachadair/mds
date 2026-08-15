package mstr_test

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/creachadair/mds/mstr"
)

const names = `
file002.txt
video-25.3.mpg
file1.txt
file12.txt
video-3.mpg
file2.txt
video-25.19.mpg
40-pic-100.jpg
5-pic-99.png
5-pic-910.heic
`

func ExampleCompareNatural() {
	// The order of the inputs as-written.
	input := strings.Fields(names)

	// Lexicographic order.
	lex := strings.Fields(names)
	slices.SortFunc(lex, cmp.Compare)

	// Natural order.
	nat := strings.Fields(names)
	slices.SortFunc(nat, mstr.CompareNatural)

	// Strict natural order.
	snat := strings.Fields(names)
	slices.SortFunc(snat, mstr.CompareNaturalStrict)

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(tw, "#\tINPUT\tLEXICOGRAPHIC\tNATURAL\tSTRICT")
	for i := range lex {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", i+1, input[i], lex[i], nat[i], snat[i])
	}
	tw.Flush()
	// Output:
	//
	// #   INPUT            LEXICOGRAPHIC    NATURAL          STRICT
	// 1   file002.txt      40-pic-100.jpg   5-pic-99.png     5-pic-99.png
	// 2   video-25.3.mpg   5-pic-910.heic   5-pic-910.heic   5-pic-910.heic
	// 3   file1.txt        5-pic-99.png     40-pic-100.jpg   40-pic-100.jpg
	// 4   file12.txt       file002.txt      file1.txt        file1.txt
	// 5   video-3.mpg      file1.txt        file002.txt      file2.txt
	// 6   file2.txt        file12.txt       file2.txt        file002.txt
	// 7   video-25.19.mpg  file2.txt        file12.txt       file12.txt
	// 8   40-pic-100.jpg   video-25.19.mpg  video-3.mpg      video-3.mpg
	// 9   5-pic-99.png     video-25.3.mpg   video-25.3.mpg   video-25.3.mpg
	// 10  5-pic-910.heic   video-3.mpg      video-25.19.mpg  video-25.19.mpg
}
