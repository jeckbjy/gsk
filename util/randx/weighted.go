package randx

import (
	"math/rand"
	"sort"
)

// 根据权重随机,返回索引和随机值
func Weighted(n int, fn func(int) int) (int, int) {
	if n == 0 {
		return -1, 0
	}
	if n == 1 {
		return 0, 0
	}

	weights := make([]int, n)
	total := 0
	for i := 0; i < n; i++ {
		weights[i] = total + fn(i)
		total = weights[i]
	}

	// 每个区间段不包括最大值,[0,max)
	value := rand.Intn(total)
	index := sort.Search(n, func(i int) bool {
		return value < weights[i]
	})

	return index, value
}
