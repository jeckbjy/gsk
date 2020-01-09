package randx

import (
	"testing"
)

func TestWeighted(t *testing.T) {
	items := []int{10, 20, 30, 40}
	index, value := Weighted(len(items), func(i int) int {
		return items[i]
	})

	t.Log(index, value)
}
