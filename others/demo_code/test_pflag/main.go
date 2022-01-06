// package main

// import (
// 	"flag"
// 	"fmt"
// )

// // 定义命令行参数对应的变量，这三个变量都是指针类型
// var cliName = flag.String("name", "nick", "Input Your Name")
// var cliAge = flag.Int("age", 28, "Input Your Age")
// var cliGender = flag.String("gender", "male", "Input Your Gender")

// // 定义一个值类型的命令行参数变量，在 Init() 函数中对其初始化
// // 因此，命令行参数对应变量的定义和初始化是可以分开的
// var cliFlag int

// func Init() {
// 	flag.IntVar(&cliFlag, "flagname", 1234, "Just for demo")
// }

// func main() {
// 	// 初始化变量 cliFlag
// 	Init()
// 	// 把用户传递的命令行参数解析为对应变量的值
// 	flag.Parse()

// 	// flag.Args() 函数返回没有被解析的命令行参数
// 	// func NArg() 函数返回没有被解析的命令行参数的个数
// 	fmt.Printf("args=%s, num=%d\n", flag.Args(), flag.NArg())
// 	for i := 0; i != flag.NArg(); i++ {
// 		fmt.Printf("arg[%d]=%s\n", i, flag.Arg(i))
// 	}

// 	// 输出命令行参数
// 	fmt.Println("name=", *cliName)
// 	fmt.Println("age=", *cliAge)
// 	fmt.Println("gender=", *cliGender)
// 	fmt.Println("flagname=", cliFlag)
// }

package main

import (
	"fmt"
	"strings"

	flag "github.com/spf13/pflag"
)

// 定义命令行参数对应的变量
var cliName = flag.StringP("name", "n", "nick", "Input Your Name")
var cliAge = flag.IntP("age", "a", 22, "Input Your Age")
var cliGender = flag.StringP("gender", "g", "male", "Input Your Gender")
var cliOK = flag.BoolP("ok", "o", false, "Input Are You OK")
var cliDes = flag.StringP("des-detail", "d", "skip", "Input Description")
var cliOldFlag = flag.StringP("badflag", "b", "just for test", "Input badflag")

func wordSepNormalizeFunc(f *flag.FlagSet, name string) flag.NormalizedName {
	from := []string{"-", "_"}
	to := "."
	for _, sep := range from {
		name = strings.Replace(name, sep, to, -1)
	}
	return flag.NormalizedName(name)
}

func main() {
	// 设置标准化参数名称的函数 flag.CommandLine实质为FlagSet结构
	flag.CommandLine.SetNormalizeFunc(wordSepNormalizeFunc)

	// 为 age 参数设置 NoOptDefVal
	flag.Lookup("age").NoOptDefVal = "25"

	// 把 badflag 参数标记为即将废弃的，请用户使用 des-detail 参数
	flag.CommandLine.MarkDeprecated("badflag", "please use --des-detail instead")
	// 把 badflag 参数的 shorthand 标记为即将废弃的，请用户使用 des-detail 的 shorthand 参数
	flag.CommandLine.MarkShorthandDeprecated("badflag", "please use -d instead")

	// 在帮助文档中隐藏参数 gender
	flag.CommandLine.MarkHidden("gender")

	// 把用户传递的命令行参数解析为对应变量的值
	flag.Parse()

	fmt.Println("name=", *cliName)
	fmt.Println("age=", *cliAge)
	fmt.Println("gender=", *cliGender)
	fmt.Println("ok=", *cliOK)
	fmt.Println("des=", *cliDes)
}
