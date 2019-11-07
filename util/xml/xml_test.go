package xml

import (
	"encoding/xml"
	"testing"
)

const xmlData1 = `
<xml>
   <appid>wx2421b1c4370ec43b</appid>
   <attach>支付测试</attach>
   <body>APP支付测试</body>
   <mch_id>10000100</mch_id>
   <nonce_str>1add1a30ac87aa2db72f57a2375d8fec</nonce_str>
   <notify_url>http://wxpay.wxutil.com/pub_v2/pay/notify.v2.php</notify_url>
   <out_trade_no>1415659990</out_trade_no>
   <spbill_create_ip>14.23.150.211</spbill_create_ip>
   <total_fee>1</total_fee>
   <trade_type>APP</trade_type>
   <sign>0CB01533B8C1EF103065174F50BCA001</sign>
</xml>
`

const xmlData2 = `
<xml>
   <return_code><![CDATA[SUCCESS]]></return_code>
   <return_msg><![CDATA[OK]]></return_msg>
   <appid><![CDATA[wx2421b1c4370ec43b]]></appid>
   <mch_id><![CDATA[10000100]]></mch_id>
   <nonce_str><![CDATA[IITRi8Iabbblz1Jc]]></nonce_str>
   <sign><![CDATA[7921E432F65EB8ED0CE9755F0E86D72F]]></sign>
   <result_code><![CDATA[SUCCESS]]></result_code>
   <prepay_id><![CDATA[wx201411101639507cbf6ffd8b0779950874]]></prepay_id>
   <trade_type><![CDATA[APP]]></trade_type>
</xml>
`

func TestStrMap(t *testing.T) {
	result1 := StringMap{}
	if err := xml.Unmarshal([]byte(xmlData1), &result1); err != nil {
		t.Fatal(err)
	} else {
		t.Log(result1)
		aa, _ := xml.MarshalIndent(result1, "", "\t")
		t.Log(string(aa))
	}

	result2 := StringMap{}
	if err := xml.Unmarshal([]byte(xmlData2), &result2); err != nil {
		t.Fatal(err)
	} else {
		t.Log(result2)
		bb, _ := xml.MarshalIndent(result2, "", "\t")
		t.Log(string(bb))
	}
}
