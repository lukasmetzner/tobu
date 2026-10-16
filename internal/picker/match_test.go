package picker

import (
	"slices"
	"testing"
)

func TestFilter(t *testing.T) {
	items := []Item{
		{Label: "/home/user/projects/tobu"},
		{Label: "/home/user/projects/Hetzner/hcloud-go"},
		{Label: "/home/user/.config/nvim"},
	}

	tests := []struct {
		name    string
		query   string
		indices []int
	}{
		{name: "empty query keeps input order", query: "", indices: []int{0, 1, 2}},
		{name: "fuzzy subsequence", query: "prtb", indices: []int{0}},
		{name: "smart case insensitive", query: "hetzner", indices: []int{1}},
		{name: "smart case sensitive", query: "hetzner/H", indices: []int{}},
		{name: "uppercase requires exact case", query: "Nvim", indices: []int{}},
		{name: "all terms must match", query: "proj go", indices: []int{1}},
		{name: "order follows input not match quality", query: "o", indices: []int{0, 1, 2}},
		{name: "no match", query: "xyz", indices: []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indices := []int{}
			for _, m := range filter(items, tt.query) {
				indices = append(indices, m.index)
			}
			if !slices.Equal(indices, tt.indices) {
				t.Errorf("filter(%q) = %v, want %v", tt.query, indices, tt.indices)
			}
		})
	}
}

func TestMatchPrefersRightmostPositions(t *testing.T) {
	matches := filter([]Item{{Label: "/tobu/src/tobu"}}, "tobu")
	if len(matches) != 1 {
		t.Fatalf("expected one match, got %d", len(matches))
	}
	if want := []int{10, 11, 12, 13}; !slices.Equal(matches[0].positions, want) {
		t.Errorf("positions = %v, want %v", matches[0].positions, want)
	}
}

func TestMatchMergesOverlappingTerms(t *testing.T) {
	matches := filter([]Item{{Label: "/abc"}}, "bc c")
	if len(matches) != 1 {
		t.Fatalf("expected one match, got %d", len(matches))
	}
	if want := []int{2, 3}; !slices.Equal(matches[0].positions, want) {
		t.Errorf("positions = %v, want %v", matches[0].positions, want)
	}
}
