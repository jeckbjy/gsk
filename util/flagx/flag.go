package flagx

import (
	"bufio"
	"flag"
	"io"
	"os"
	"strings"
	"time"
)

var ConfigFile = ""
var Prefix = ""

func Bool(name string, value bool, usage string) *bool {
	return flag.Bool(name, value, usage)
}

func Int(name string, value int, usage string) *int {
	return flag.Int(name, value, usage)
}

func Int64(name string, value int64, usage string) *int64 {
	return flag.Int64(name, value, usage)
}

func Uint(name string, value uint, usage string) *uint {
	return flag.Uint(name, value, usage)
}

func Uint64(name string, value uint64, usage string) *uint64 {
	return flag.Uint64(name, value, usage)
}

func String(name string, value string, usage string) *string {
	return flag.String(name, value, usage)
}

func Float64(name string, value float64, usage string) *float64 {
	return flag.Float64(name, value, usage)
}

func Duration(name string, value time.Duration, usage string) *time.Duration {
	return flag.Duration(name, value, usage)
}

func BoolVar(p *bool, name string, value bool, usage string) {
	flag.BoolVar(p, name, value, usage)
}

func IntVar(p *int, name string, value int, usage string) {
	flag.IntVar(p, name, value, usage)
}

func Int64Var(p *int64, name string, value int64, usage string) {
	flag.Int64Var(p, name, value, usage)
}

func UintVar(p *uint, name string, value uint, usage string) {
	flag.UintVar(p, name, value, usage)
}

func Uint64Var(p *uint64, name string, value uint64, usage string) {
	flag.Uint64Var(p, name, value, usage)
}

func StringVar(p *string, name string, value string, usage string) {
	flag.StringVar(p, name, value, usage)
}

func Float64Var(p *float64, name string, value float64, usage string) {
	flag.Float64Var(p, name, value, usage)
}

func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	flag.DurationVar(p, name, value, usage)
}

// 扩展且兼容官方flag库,支持从文件中加载,支持从Env中加载
// Env key会做一些变换,仅使用大写字母且使用_代替-
// 默认解析顺序:file->env->flag
// 出处:https://github.com/namsral/flag
func Parse() {
	if err := ParseFile(flag.CommandLine, ConfigFile); err != nil {
		os.Exit(2)
	}

	if err := ParseEnv(flag.CommandLine, Prefix); err != nil {
		os.Exit(2)
	}

	flag.Parse()
}

// 从配置文件中解析,格式`key=value`
func ParseFile(set *flag.FlagSet, path string) error {
	if path == "" {
		return nil
	}

	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	return parseReader(set, fp)
}

func parseReader(set *flag.FlagSet, reader io.Reader) error {
	kv := make(map[string]string)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// match `key=value`
		key, val := splitKV(line)
		if key != "" {
			kv[key] = val
		}
	}

	// visit set
	var result error
	set.VisitAll(func(f *flag.Flag) {
		value, ok := kv[f.Name]
		if !ok {
			return
		}

		if err := set.Set(f.Name, value); err != nil {
			result = err
			return
		}
	})
	return result
}

// ParseEnv 解析环境变量,会将Key变为大写并使用'_'代替'-'
func ParseEnv(set *flag.FlagSet, prefix string) error {
	return parseEnv(set, prefix, os.Environ())
}

// parseEnv for test
func parseEnv(set *flag.FlagSet, prefix string, environ []string) error {
	env := make(map[string]string)
	for _, s := range environ {
		key, val := splitKV(s)
		if key != "" {
			env[key] = val
		}
	}

	var result error

	set.VisitAll(func(f *flag.Flag) {
		key := strings.ToUpper(f.Name)
		if prefix != "" {
			key = prefix + "_" + key
		}
		key = strings.Replace(key, "-", "_", -1)

		value, ok := env[key]
		if !ok {
			return
		}

		if err := set.Set(f.Name, value); err != nil {
			result = err
			return
		}
	})

	return result
}

// parse key=value
func splitKV(str string) (string, string) {
	idx := strings.Index(str, "=")
	if idx == -1 {
		return str, "true"
	}

	return str[:idx], str[idx+1:]
}
