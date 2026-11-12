package uv

import (
	"fmt"
	"image/color"
	"strings"
	"testing"
)

func TestScrollRows(t *testing.T) {
	for _, tc := range []struct {
		name   string
		insert bool
		area   Rectangle
		n      int
		want   []string
	}{
		{"delete/full", false, Rect(0, 1, 6, 4), 2, []string{"AAAAAA", "BBBBBB", "EEEEEE", "xxxxxx", "xxxxxx", "FFFFFF"}},
		{"insert/full", true, Rect(0, 1, 6, 4), 2, []string{"AAAAAA", "BBBBBB", "xxxxxx", "xxxxxx", "CCCCCC", "FFFFFF"}},
		{"delete/margins", false, Rect(1, 1, 4, 4), 2, []string{"AAAAAA", "BBBBBB", "CEEEEC", "DxxxxD", "ExxxxE", "FFFFFF"}},
		{"insert/margins", true, Rect(1, 1, 4, 4), 2, []string{"AAAAAA", "BBBBBB", "CxxxxC", "DxxxxD", "ECCCCE", "FFFFFF"}},
		{"delete/clamped", false, Rect(0, 1, 6, 4), 20, []string{"AAAAAA", "BBBBBB", "xxxxxx", "xxxxxx", "xxxxxx", "FFFFFF"}},
		{"insert/clamped", true, Rect(0, 1, 6, 4), 20, []string{"AAAAAA", "BBBBBB", "xxxxxx", "xxxxxx", "xxxxxx", "FFFFFF"}},
		{"delete/one", false, Rect(0, 1, 6, 4), 1, []string{"AAAAAA", "BBBBBB", "DDDDDD", "EEEEEE", "xxxxxx", "FFFFFF"}},
		{"insert/one", true, Rect(0, 1, 6, 4), 1, []string{"AAAAAA", "BBBBBB", "xxxxxx", "CCCCCC", "DDDDDD", "FFFFFF"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBuffer(6, 6)
			for y := range 6 {
				for x := range 6 {
					b.SetCell(x, y, &Cell{Content: string(rune('A' + y)), Width: 1, Style: Style{Fg: color.RGBA{R: uint8(y + 1), A: 255}}, Link: Link{URL: fmt.Sprint(y)}})
				}
			}
			fill := &Cell{Content: "x", Width: 1, Style: Style{Bg: color.RGBA{B: 255, A: 255}}}
			if tc.insert {
				b.InsertLineArea(2, tc.n, fill, tc.area)
			} else {
				b.DeleteLineArea(2, tc.n, fill, tc.area)
			}
			if got := b.String(); got != strings.Join(tc.want, "\n") {
				t.Fatalf("got %q, want %q", got, strings.Join(tc.want, "\n"))
			}
			for y, row := range tc.want {
				for x, ch := range row {
					want := *fill
					if ch != 'x' {
						i := int(ch - 'A')
						want = Cell{Content: string(ch), Width: 1, Style: Style{Fg: color.RGBA{R: uint8(i + 1), A: 255}}, Link: Link{URL: fmt.Sprint(i)}}
					}
					if got := b.CellAt(x, y); !got.Equal(&want) {
						t.Fatalf("cell %d,%d: got %#v, want %#v", x, y, got, want)
					}
				}
			}
			// Recycled rows must not share storage with any surviving row.
			b.SetCell(0, 3, &Cell{Content: "!", Width: 1})
			for y, row := range tc.want {
				if y != 3 && b.CellAt(0, y).Content != string(row[0]) {
					t.Fatalf("changing recycled row changed row %d", y)
				}
			}
		})
	}
}

func TestScrollWideRows(t *testing.T) {
	for _, insert := range []bool{false, true} {
		b := NewBuffer(6, 5)
		for y := range 5 {
			b.SetCell(y%3, y, &Cell{Content: "界", Width: 2, Style: Style{Attrs: AttrBold}})
		}
		before := make([]Line, 5)
		for y := range 5 {
			before[y] = append(Line(nil), b.Line(y)...)
		}
		if insert {
			b.InsertLineArea(1, 1, nil, b.Bounds())
		} else {
			b.DeleteLineArea(1, 1, nil, b.Bounds())
		}
		for y := range 5 {
			src := y
			if y >= 1 {
				if insert {
					src = y - 1
				} else {
					src = y + 1
				}
			}
			for x := range 6 {
				want := EmptyCell
				if y == 0 || (insert && y > 1) || (!insert && y < 4) {
					want = before[src][x]
				}
				if !b.CellAt(x, y).Equal(&want) {
					t.Fatalf("insert=%t cell %d,%d got %#v, want %#v", insert, x, y, b.CellAt(x, y), want)
				}
			}
		}
	}
}

func TestScrollOutsideVerticalBounds(t *testing.T) {
	for _, insert := range []bool{false, true} {
		for _, tc := range []struct {
			name string
			y, n int
			area Rectangle
			want string
		}{
			{"past bottom", 0, 3, Rect(0, 0, 3, 3), "\n"},
			{"before top", -1, 2, Rect(0, -1, 3, 2), "\nBBB"},
		} {
			t.Run(fmt.Sprintf("insert=%t/%s", insert, tc.name), func(t *testing.T) {
				b := NewBuffer(3, 2)
				for y := range 2 {
					for x := range 3 {
						b.SetCell(x, y, &Cell{Content: string(rune('A' + y)), Width: 1})
					}
				}
				if insert {
					b.InsertLineArea(tc.y, tc.n, nil, tc.area)
				} else {
					b.DeleteLineArea(tc.y, tc.n, nil, tc.area)
				}
				if got := b.String(); got != tc.want {
					t.Fatalf("got %q, want %q", got, tc.want)
				}
			})
		}
	}
}

func TestStyleEqualWithDefaultColors(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b Style
		want bool
	}{
		{"default", Style{}, Style{}, true},
		{"attributes", Style{Attrs: AttrBold}, Style{Attrs: AttrBold}, true},
		{"different attributes", Style{}, Style{Attrs: AttrBold}, false},
		{"underline", Style{Underline: UnderlineStyleSingle}, Style{Underline: UnderlineStyleSingle}, true},
		{"different underline", Style{}, Style{Underline: UnderlineStyleSingle}, false},
		{"foreground", Style{}, Style{Fg: color.RGBA{}}, false},
		{"background", Style{}, Style{Bg: color.RGBA{}}, false},
		{"underline color", Style{}, Style{UnderlineColor: color.RGBA{}}, false},
		{"color representations", Style{Fg: color.RGBA{R: 255, A: 255}}, Style{Fg: color.NRGBA{R: 255, A: 255}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.a.Equal(&tc.b) != tc.want || tc.b.Equal(&tc.a) != tc.want {
				t.Fatalf("equal=%t, reverse=%t, want %t", tc.a.Equal(&tc.b), tc.b.Equal(&tc.a), tc.want)
			}
		})
	}
}

func BenchmarkScrollRows(b *testing.B) {
	for _, height := range []int{24, 60} {
		b.Run(fmt.Sprint(height), func(b *testing.B) {
			buf := NewBuffer(160, height)
			area := buf.Bounds()
			b.ReportAllocs()
			for b.Loop() {
				buf.DeleteLineArea(0, 1, nil, area)
			}
		})
	}
}
