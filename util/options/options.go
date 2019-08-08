package options

import (
	"context"
	"time"
)

func SetString(x *string, d string) {
	if *x == "" {
		*x = d
	}
}

func SetDuration(x *time.Duration, d time.Duration) {
	if *x == 0 {
		*x = d
	}
}

func SetContext(x *context.Context) {
	if *x == nil {
		*x = context.Background()
	}
}
