package luhn

import "testing"

// verify from: https://www.dcode.fr/luhn-algorithm
func TestGenerate(t *testing.T) {
	t.Log(GenerateInt(0))
	t.Log(GenerateInt(1))
	t.Log(GenerateInt(11))
	t.Log(GenerateInt(12))
	t.Log(GenerateInt(21))
	t.Log(GenerateStr("11"))
	t.Log(GenerateStr("12"))
	t.Log(GenerateStr("21"))

	for i := uint64(0); i < 100; i++ {
		x := GenerateInt(i)
		if !Check(x) {
			t.Errorf("check fail,%+v", i)
		} else {
			t.Logf("check ok,%+v", i)
		}
	}
}
