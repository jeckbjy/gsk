package cli

import "testing"

func TestEngine(t *testing.T) {
	app := New()

	if err := app.Exec(nil, "rank top id"); err != nil {
		t.Error(err)
	}
}
