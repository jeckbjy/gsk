package breaker

import "fmt"

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
	// 区别于标准的状态,这里增加了health状态,假设大部分情况系统的良好的,不用做额外的容错
	// 当第一个错误发生时,再开启容错统计,可以减少统计带来的开销
	stateHealth
)

//服务的健康状况 = 请求失败数 / 请求总数.
//熔断器开关由关闭到打开的状态转换是通过当前服务健康状况和设定阈值比较决定的.
//
//	当熔断器开关关闭时, 请求被允许通过熔断器. 如果当前健康状况高于设定阈值, 开关继续保持关闭. 如果当前健康状况低于设定阈值, 开关则切换为打开状态.
//	当熔断器开关打开时, 请求被禁止通过.
//	当熔断器开关处于打开状态, 经过一段时间后, 熔断器会自动进入半开状态, 这时熔断器只允许一个请求通过. 当该请求调用成功时, 熔断器恢复到关闭状态. 若该请求失败, 熔断器继续保持打开状态, 接下来的请求被禁止通过.
type State int

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return fmt.Sprintf("unknown state: %d", s)
	}
}
