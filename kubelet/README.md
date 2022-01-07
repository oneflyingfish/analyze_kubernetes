<h2 align="center">Kubelet源码阅读笔记</h2>

> 主体程序路径：`kubernetes/cmd/kubelet`
>
> 实验环境：
>
> * 服务器：
>   * 系统：`Centos 7.9`
>   * go：`v1.17.5`
> * 本地：
>   * 系统：`Windows 10`
>   * 开发环境：`VS Code` + `SSH-Remote`
> * kubernetes：`v1.23.1`

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

* 默认初始化cleanFlagSet结构

  ```go
  cleanFlagSet := pflag.NewFlagSet(componentKubelet, pflag.ContinueOnError)
  ```

  * type: pflag.FlagSet

  * 作用：类似pflag结构的定义，用于生成命令行应用

  * 拓展：

    * pflag的三种使用方式

      > import "github.com/spf13/pflag"

      * 方式一：
    
        ```go
        flagSet := pflag.NewFlagSet("main", pflag.ExitOnError)
        
        version := flagSet.BoolP("version", "v", false, "print version string")		// 参数定义
        flagSet.Parse(os.Args[1:])													// 解析：命令应用启动参数
        
        // 读取：使用参数
        // 方法一：
        fmt.Println(*version)		
        
        // 方法二：
        value, _ := flagSet.GetBool("version")
        fmt.Println(value)
        ```
      
      * 方式二：
      
        ```go
        flagSet := pflag.NewFlagSet("main", pflag.ExitOnError)
        flagSet.Parse(os.Args[1:])													// 解析：命令应用启动参数
        value, _ := flagSet.GetBool("version")										// 读取：使用参数
        fmt.Println(value)
        ```
      
        
      
      * 方式三：
      
        ```go
        version := pflag.BoolP("version","v", false, "print version string")		// 参数定义
        pflag.Parse()																// 解析
        
        fmt.println(version)														// 读取
        ```
      
      * 附加：
      
        若使用package为golang内置的`flag`而不是`pflag`（增强版），可使用如下：
      
        ```go
        flagSet := flag.NewFlagSet("main", flag.ExitOnError)
        
        version := flagSet.Bool("version", false, "print version string")			// 参数定义
        flagSet.Parse(os.Args[1:])													// 解析：命令应用启动参数
        
        // 读取：使用参数
        // 方法一：
        value := flagSet.Lookup("version").Value.(flag.Getter).Get().(bool)			
        fmt.Println(value)
        
        // 方法二
        fmt.Println(*version)														
        ```
      
      > 启动参照： ./go_exec --version=true
    

* 默认初始化kubeletFlags结构

  ```go
  // 默认值初始化KubeletFlags结构,包括docker,证书路径，插件目录，CIDR等等
  kubeletFlags := options.NewKubeletFlags()
  ```

  * type: option.KubeletFlags
  * 作用：存储从`kubelet`命令启动时追加的参数（Flags）

* 默认初始化kubeletConfig结构

  ```go
  kubeletConfig, err := options.NewKubeletConfiguration()
  ```

  * type: config.kubeletConfiguration

  * 作用：存储从配置文件`--kubeconfig=$PATH/kubelet.kubeconfig`获取的参数

    * Node节点在CSR证书申请被批准后，自动在本地生成`kubelet.kubeconfig`，下次启动将根据此配置文件直接注册，而不用再次发起请求

  * 内容示例：

    ```yaml
    apiVersion: v1
    clusters:
    - cluster:
        certificate-authority-data: LS0tLS1CR......LS0tLS0K
        server: https://$Master_IP:6443
      name: default-cluster
    contexts:
    - context:
        cluster: default-cluster
        namespace: default
        user: default-auth
      name: default-context
    current-context: default-context
    kind: Config
    preferences: {}
    users:
    - name: default-auth
      user:
        client-certificate: $PATH/ssl/kubelet-client-current.pem
        client-key: $PATH/ssl/kubelet-client-current.pem
    ```

    

* cmd初始化

  * DisableFlagParsing: true

    >  即禁用cobra包的flags自动解析。flags会被直接解析为参数`args`的一部分（注意`子命令`不是`flags`），其中包括`--help`等。
    >
    > 
    >
    > 以下通过简单示例说明此参数影响：

    * DisableFlagParsing: false

      ```shell
      apt-get install package -f
      
      # 解析结果
      command: `apt-get install`
      flags: ["-f"]
      args: ["package"]
      ```

    * DisableFlagParsing: true

      ```shell
      apt-get install package -f
      
      # 解析结果
      command: `apt-get install`
      flags: []
      args: ["-f","package"]
      ```

    * 将此参数值设置为`true`，使得`kubelet`可以在`Run`函数里，完全自定义可控的方式处理程序

  * Run函数编写

    > 即执行kubelet命令时的调用入口函数

    * 对程序的输入命令进行解析，判断输入参数合法性
      * `cleanFlagSet.Parse(args)`
        * 解析程序flags
        * 如果出现有未定义的flags，将返回error
      * `cleanFlagSet.Args()`
        * 解析程序子命令
        * kubelet并不支持子命令，将直接报错并结束程序

    * 

    











