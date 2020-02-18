# cli Command-line interface

command解析参数,格式化输出命令,可用于控制台参数解析,也可以用于服务器端admin模块,
这里要解决的问题是如何自动化的绑定数据并作简单的数据验证以及信息输出,以便于高效开发

在分布式集群环境下,admin的开发更加复杂一些,需要将所有不同节点上的Command汇总并根据不同的策略将Command发送到特定的节点上,
常见的策略有:广播或随机选择节点

cli文件下定义了所有用到的接口

## 特性

- 基于反射自动解析参数
- 支持简单的参数验证,比如参数个数验证,类型转换验证,参数范围验证
- 支持默认参数
- TODO:格式化输出help信息

## 使用示例

```go
package cli

import (
	"log"
	"testing"
)

func TestCmd(t *testing.T) {
	args, err := ParseCommandLine("test 1 -a=test")
	if err != nil {
		t.Fatal(err)
	}

	app := New()
	_ = app.Add(&testCmd{})
	result, _ := app.Exec(args, map[string]string{"project": "Apollo"})
	t.Log(result)
}

type testCmd struct {
	Project string `cli:"meta"`
	Arg0    int    `cli:"desc=参数0"`
	Arg1    string `cli:"flag=a|arg1,default=aa,desc=参数1"`
}

func (cmd *testCmd) Run(ctx Context) error {
	log.Print(cmd.Project, "\t", cmd.Arg0, "\t", cmd.Arg1)
	return ctx.Text("test result ok")
}
```

## 常用的cli库

- [cobra](https://github.com/spf13/cobra)
- [urfave/cli](https://github.com/urfave/cli)
- [go-flags](https://github.com/jessevdk/go-flags)