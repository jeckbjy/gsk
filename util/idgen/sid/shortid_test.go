package sid

import "testing"

func TestGenerate(t *testing.T) {
	for i := 0; i < 10; i++ {
		id := MustGenerate()
		t.Log(id)
	}
}
