package trie

import (
	"sort"
	"testing"
)

func TestDat(t *testing.T) {
	words := []string{
		"hello",
		"hello world",
		"我",
		"我爱你",
		"我爱中国",
		"你好",
		"大家好",
	}

	sort.Strings(words)
	// for _, str := range words {
	// 	fmt.Println(str)
	// }

	dat := NewDATrie()
	dat.Build(words)
	// dat.dump()
	for _, word := range words {
		length := dat.Match(word, true)
		if length != len(word) {
			t.Errorf("cannot match:%+v, %+v", word, length)
		} else {
			t.Logf("ok---match:%+v,%+v", length, word)
		}
	}

	t.Logf("------------------------")
	wordsPrefix := []string{
		"hello boy",
		"你好呀",
	}
	for _, word := range wordsPrefix {
		length := dat.Match(word, true)
		if length == 0 {
			t.Errorf("cannot match:%+v, %+v", word, length)
		} else {
			t.Logf("ok---match prefix:%+v,%+v", length, word)
		}
	}

	wordsBad := []string{
		"hell",
		"good",
		"good job",
		"江泽民",
		"胡锦涛",
		"习近平",
	}

	t.Logf("------------------------")
	for _, word := range wordsBad {
		length := dat.Match(word, true)
		if length > 0 {
			t.Errorf("match bad word:%+v, %+v", word, length)
		} else {
			t.Logf("ok---not match:%+v,%+v", length, word)
		}
	}
}
