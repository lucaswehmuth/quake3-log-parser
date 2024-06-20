package utils

import "sort"

// Pair represents a key-value pair from a map
type Pair struct {
	Key   string
	Value int
}

// SortMapByValueDescending sorts a map by its values in descending order and returns a sorted slice of Pairs
func SortMapByValueDescending(m map[string]int) []Pair {
	pl := make([]Pair, 0, len(m))
	for k, v := range m {
		pl = append(pl, Pair{Key: k, Value: v})
	}

	// Sort using sort.Slice with a custom comparison function
	sort.Slice(pl, func(i, j int) bool {
		return pl[i].Value > pl[j].Value
	})

	return pl
}
