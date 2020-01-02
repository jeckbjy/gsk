package flagx

import (
	"flag"
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	file := `
bool
bool2=true
int=22
int64=0x23
uint=24
uint64=25
string=hello
float64=2718e28
duration=2m
`
	environ := []string{
		"BOOL",
		"BOOL2=true",
		"INT=22",
		"INT64=0x23",
		"UINT=24",
		"UINT64=25",
		"STRING=hello",
		"FLOAT64=2718e28",
		"DURATION=2m",
	}
	for i := 0; i < 2; i++ {
		f := flag.NewFlagSet("test", flag.ExitOnError)
		boolFlag := f.Bool("bool", false, "bool value")
		bool2Flag := f.Bool("bool2", false, "bool2 value")
		intFlag := f.Int("int", 0, "int value")
		int64Flag := f.Int64("int64", 0, "int64 value")
		uintFlag := f.Uint("uint", 0, "uint value")
		uint64Flag := f.Uint64("uint64", 0, "uint64 value")
		stringFlag := f.String("string", "0", "string value")
		float64Flag := f.Float64("float64", 0, "float64 value")
		durationFlag := f.Duration("duration", 5*time.Second, "time.Duration value")

		if i == 0 {
			if err := parseReader(f, strings.NewReader(file)); err != nil {
				t.Fatal(err)
			}
		} else if i == 1 {
			if err := parseEnv(f, "", environ); err != nil {
				t.Fatal(err)
			}
		}

		if *boolFlag != true {
			t.Error("bool flag should be true, is ", *boolFlag)
		}
		if *bool2Flag != true {
			t.Error("bool2 flag should be true, is ", *bool2Flag)
		}
		if *intFlag != 22 {
			t.Error("int flag should be 22, is ", *intFlag)
		}
		if *int64Flag != 0x23 {
			t.Error("int64 flag should be 0x23, is ", *int64Flag)
		}
		if *uintFlag != 24 {
			t.Error("uint flag should be 24, is ", *uintFlag)
		}
		if *uint64Flag != 25 {
			t.Error("uint64 flag should be 25, is ", *uint64Flag)
		}
		if *stringFlag != "hello" {
			t.Error("string flag should be `hello`, is ", *stringFlag)
		}
		if *float64Flag != 2718e28 {
			t.Error("float64 flag should be 2718e28, is ", *float64Flag)
		}
		if *durationFlag != 2*time.Minute {
			t.Error("duration flag should be 2m, is ", *durationFlag)
		}
	}
}

func TestOverride(t *testing.T) {
	file := `
age=1
female=false
length=2.2
`
	environ := []string{
		"AGE=2",
		"FEMALE=true",
	}
	args := []string{
		"-age=3",
	}

	f := flag.NewFlagSet("test", flag.ExitOnError)
	age := f.Int("age", 0, "age of gopher")
	female := f.Bool("female", true, "")
	length := f.Float64("length", 0, "")
	if err := parseReader(f, strings.NewReader(file)); err != nil {
		t.Fatal(err)
	}
	if err := parseEnv(f, "", environ); err != nil {
		t.Fatal(err)
	}
	if err := f.Parse(args); err != nil {
		t.Fatal(err)
	}
	if *age != 3 {
		t.Errorf("bad age,%+v", *age)
	}
	if *female != true {
		t.Errorf("bad female,%+v", *female)
	}
	if *length != 2.2 {
		t.Fatalf("bad length,%+v", *length)
	}
	t.Log(*age, *female, *length)
}
