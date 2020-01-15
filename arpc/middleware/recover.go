package middleware

import (
	"fmt"

	"github.com/jeckbjy/gsk/arpc"
)

// Recover 用于拦截panic,打印报错
func Recover() arpc.HandlerFunc {
	return func(ctx arpc.Context) error {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", err)
				}
				ctx.Abort(err)
			}
		}()

		return ctx.Next()
	}
}
