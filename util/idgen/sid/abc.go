package sid

import (
	"errors"
	"fmt"
)

// DefaultABC is the default URL-friendly alphabet.
const DefaultABC = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-"

// NewAbc constructs a new instance of shuffled alphabet to be used for Id representation.
func NewAbc(alphabet string, seed uint64) (Abc, error) {
	runes := []rune(alphabet)
	l := len(runes)
	if l != 32 && l != 64 {
		return Abc{}, fmt.Errorf("alphabet length must be 32 or 64")
	}

	if nonUnique(runes) {
		return Abc{}, errors.New("alphabet must contain unique characters only")
	}
	abc := Abc{alphabet: nil}
	abc.shuffle(alphabet, seed)
	return abc, nil
}

// MustNewAbc acts just like NewAbc, but panics instead of returning errors.
func MustNewAbc(alphabet string, seed uint64) Abc {
	res, err := NewAbc(alphabet, seed)
	if err == nil {
		return res
	}
	panic(err)
}

func nonUnique(runes []rune) bool {
	found := make(map[rune]struct{})
	for _, r := range runes {
		if _, seen := found[r]; !seen {
			found[r] = struct{}{}
		}
	}
	return len(found) < len(runes)
}

// Abc represents a shuffled alphabet used to generate the Ids and provides methods to
// encode data.
type Abc struct {
	alphabet []rune
}

func (abc *Abc) Len() int {
	return len(abc.alphabet)
}

// String returns a string representation of the Abc instance.
func (abc *Abc) String() string {
	return fmt.Sprintf("Abc{alphabet='%v')", abc.Alphabet())
}

// Alphabet returns the alphabet used as an immutable string.
func (abc *Abc) Alphabet() string {
	return string(abc.alphabet)
}

func (abc *Abc) shuffle(alphabet string, seed uint64) {
	source := []rune(alphabet)
	for len(source) > 1 {
		seed = (seed*9301 + 49297) % 233280
		i := int(seed * uint64(len(source)) / 233280)

		abc.alphabet = append(abc.alphabet, source[i])
		source = append(source[:i], source[i+1:]...)
	}
	abc.alphabet = append(abc.alphabet, source[0])
}
