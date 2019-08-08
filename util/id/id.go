package id

import (
	"github.com/jeckbjy/micro/util/id/sid"
	"github.com/jeckbjy/micro/util/id/xid"
)

// ShortID(SID)和XID的区别是:
// SID不能直接应用在集群间,有冲突的可能,需要制定唯一workerid,而且不能超过32,限制比较多,
// 但生成的id比较短,大部分情况下也不会冲突
// XID:可以应用于集群间保证唯一,算法使用的是mongo的方式,string使用base32编码,
// 若改成base64可以缩短字符长度,但会产生特殊字符

// NewSID 新建ShortID
func NewSID() string {
	return sid.MustGenerate()
}

// NewXID 新建一个string类型的ID
func NewXID() string {
	return xid.New().String()
}
