package httpx

import (
	"testing"
)

func TestPost(t *testing.T) {
	url := "https://openapi.alipay.com/gateway.do"
	var result string
	rsp, err := Post(url, nil, &result, ContentType(TypeForm))
	if err != nil {
		t.Error(err)
	} else {
		//t.Log(result)
		t.Log(rsp)
	}
}

func TestGet(t *testing.T) {
	var result string
	rsp, err := Get("https://www.baidu.com/", &result)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(rsp.StatusCode, rsp.Header)
		t.Log(result)
	}

	// get with query
	xmres := make(map[string]interface{})

	query := map[string]string{
		"appId":     "testappid",
		"cpOrderId": "testOrderId",
		"uid":       "uid",
		"signature": "testSign",
	}
	if _, err := Get("http://mis.migc.xiaomi.com/api/biz/service/queryOrder.do", &xmres, QueryMap(query)); err != nil {
		t.Fatal(err)
	} else {
		t.Log(xmres)
	}
}
