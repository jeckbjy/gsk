package cli

import (
	"log"
	"testing"
)

func TestCmd(t *testing.T) {
	args, err := ParseCommandLine("test 1 -a=test")
	if err != nil {
		t.Fatal(err)
	}

	app := New()
	_ = app.Add(&testCmd{})
	result, _ := app.Exec(args, map[string]string{"project": "Apollo"})
	t.Log(result)
}

type testCmd struct {
	Project string `cli:"meta"`
	Arg0    int    `cli:"desc=参数0"`
	Arg1    string `cli:"flag=a|arg1,default=aa,desc=参数1"`
}

func (cmd *testCmd) Run(ctx Context) error {
	log.Print(cmd.Project, "\t", cmd.Arg0, "\t", cmd.Arg1)
	return ctx.Text("test result ok")
}
