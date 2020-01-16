package middleware

import "github.com/jeckbjy/gsk/arpc"

// Proxy 代理转发,可用于Gateway或者Proxy转发消息
func Proxy() arpc.HandlerFunc {
	return nil
}
