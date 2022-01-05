<h2 align="center">Kubelet源码阅读笔记</h2>

### 一、Kubelet组件

> 主体程序路径：kubernetes/cmd/kubelet

#### 1. 整体结构：

```go
func main() {
	
    // 通过 command.Execute() 执行
    command := app.NewKubeletCommand()	// type: *cobra.Command，创建kubelet命令行应用
    
	code := run(command)				// 运行kubelet并记录日志，进程正常情况不会退出
	os.Exit(code)						// 出错时返回1，结束进程；否则返回0
}
```

#### 2. kubelet命令行创建

> 即实例化 var command *cobra.command

