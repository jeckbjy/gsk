package dsn

import "testing"

func TestDSN(t *testing.T) {
	var opt struct {
		Password string `dsn:"pass"`
		DB       int    `dsn:"db"`
		PoolSize int    `dsn:"poolSize"`
	}
	u := "redis://test_addr:1234?pass=test&db=0&poolSize=100"
	dsn, err := Parse(u)
	if err != nil {
		t.Error(err)
	} else {
		if err := dsn.Bind(&opt); err != nil {
			t.Fatal(err)
		}

		t.Log(dsn.Driver)
		t.Log(dsn.Host)
		t.Log(dsn.Password)
		t.Log(opt)
	}
}
