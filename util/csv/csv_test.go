package csv

import (
	"fmt"
	"testing"
)

type Client struct { // Our example struct, you can use "-" to ignore a field
	Id      string `csv:"client_id"`
	Name    string `csv:"client_name"`
	Age     string `csv:"client_age"`
	NotUsed string `csv:"-"`
}

func (t Client) String() string {
	return fmt.Sprintf("{%v %v %v %v}", t.Id, t.Name, t.Age, t.NotUsed)
}

func TestUnmarshal(t *testing.T) {
	file := `
client_id,client_name,client_age
int,string,int
id,name,age
1,Jose,42
2,Daniel,26
3,Vincent,32
`

	var clients []*Client

	if err := Unmarshal([]byte(file), &clients); err != nil {
		t.Error(err)
	} else {
		t.Log(clients)
	}
}
