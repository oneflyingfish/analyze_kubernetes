<h2 align="center">Kubelet源码阅读笔记</h2>

> 主体程序路径：kubernetes/cmd/kubelet

#### 1. 目录结构

```shell
kubelet/
├── app
│   ├── auth.go
│   ├── init_others.go
│   ├── init_windows.go
│   ├── init_windows_test.go
│   ├── options
│   │   ├── container_runtime.go
│   │   ├── globalflags.go
│   │   ├── globalflags_linux.go
│   │   ├── globalflags_other.go
│   │   ├── globalflags_providerless.go
│   │   ├── globalflags_providers.go
│   │   ├── options.go
│   │   ├── options_test.go
│   │   ├── osflags_others.go
│   │   └── osflags_windows.go
│   ├── OWNERS
│   ├── plugins.go
│   ├── plugins_providerless.go
│   ├── plugins_providers.go
│   ├── server_bootstrap_test.go
│   ├── server.go
│   ├── server_linux.go
│   ├── server_others.go
│   ├── server_test.go
│   ├── server_unsupported.go
│   └── server_windows.go
├── kubelet.go
└── OWNERS
```

#### 2. 程序入口：

```go
func main() {
    // 通过 command.Execute() 执行
    command := app.NewKubeletCommand()	// type: *cobra.Command，创建kubelet命令行应用
    
	code := run(command)				// 运行kubelet并记录日志，进程正常情况不会退出
	os.Exit(code)						// 出错时返回1，结束进程；否则返回0
}
```

#### 3. kubelet命令行创建

> 即实例化 var command *cobra.command

```go
func NewKubeletCommand() *cobra.Command {
	// ...
	return cmd
}
```

>* cleanFlagSet := pflag.NewFlagSet(componentKubelet, pflag.ContinueOnError)
>  * type: pflag.FlagSet
>  * 创建pflag.FlagSet结构，默认值填充。
>  * 作用：用于存储kubelet启动时附加的命令行参数











