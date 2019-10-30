package basex

// Shuffle a given string using Fisher–Yates shuffle Algorithm
// same seed will get same result
func Shuffle(alphabet string, seed uint64) string {
	source := []rune(alphabet)
	length := len(source)
	for i := length - 1; i >= 0; i-- {
		seed = (seed*9301 + 49297) % 233280
		j := int(seed * uint64(length) / 233280)
		source[i], source[j] = source[j], source[i]
	}

	return string(source)
}

// CheckUnique check string is unique
func CheckUnique(alphabet string) bool {
	runes := []rune(alphabet)
	found := make(map[rune]struct{})
	for _, r := range runes {
		if _, seen := found[r]; !seen {
			found[r] = struct{}{}
		}
	}

	return len(found) == len(runes)
}
