package picker

import (
	"slices"
	"strings"
	"unicode"
)

type match struct {
	index     int
	positions []int
}

type term struct {
	runes         []rune
	caseSensitive bool
}

func filter(items []Item, query string) []match {
	terms := parseTerms(query)
	matches := make([]match, 0, len(items))
	for i, item := range items {
		if positions, ok := matchTerms([]rune(item.Label), terms); ok {
			matches = append(matches, match{index: i, positions: positions})
		}
	}
	return matches
}

func parseTerms(query string) []term {
	fields := strings.Fields(query)
	terms := make([]term, len(fields))
	for i, field := range fields {
		terms[i] = term{
			runes:         []rune(field),
			caseSensitive: field != strings.ToLower(field),
		}
	}
	return terms
}

func matchTerms(text []rune, terms []term) ([]int, bool) {
	var positions []int
	for _, t := range terms {
		termPositions, ok := t.match(text)
		if !ok {
			return nil, false
		}
		positions = append(positions, termPositions...)
	}
	slices.Sort(positions)
	return slices.Compact(positions), true
}

func (t term) match(text []rune) ([]int, bool) {
	positions := make([]int, len(t.runes))
	next := len(t.runes) - 1
	for i := len(text) - 1; i >= 0 && next >= 0; i-- {
		if t.equal(text[i], t.runes[next]) {
			positions[next] = i
			next--
		}
	}
	return positions, next < 0
}

func (t term) equal(textRune, termRune rune) bool {
	if t.caseSensitive {
		return textRune == termRune
	}
	return unicode.ToLower(textRune) == termRune
}
