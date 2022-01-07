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
  * 作用：从`kubelet`命令启动时追加的参数（Flags）初始化结构

* 默认初始化kubeletConfig结构

  ```go
  kubeletConfig, err := options.NewKubeletConfiguration()
  ```

  * type: config.kubeletConfiguration

  * 作用：从配置文件`--kubeconfig=$PATH/kubelet.kubeconfig`获取的参数初始化结构

    * Node节点在CSR证书申请被批准后，自动在本地生成`kubelet.kubeconfig`，下次启动将根据此配置文件直接注册，而不用再次发起请求

  * `kubelet.kubeconfig`文件内容示例：

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

  * `DisableFlagParsing: true`

    >  即禁用cobra包的flags自动解析。flags会被直接解析为参数`args`的一部分（注意`子命令`不是`flags`），其中包括`--help`等。
    
    以下通过简单示例说明此参数影响：
    
    * `DisableFlagParsing: false`

      ```shell
    apt-get install package -f
      
      # 解析结果
      command: `apt-get install`
      flags: ["-f"]
      args: ["package"]
      ```
    
    * `DisableFlagParsing: true`

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

    * 对程序的输入命令进行初步解析，判断输入参数合法性

      * `cleanFlagSet.Parse(args)`

        > * 解析程序flags
        >
        > * 如果出现有未定义的flags，将返回error
  
      * `cleanFlagSet.Args()`

        > * 解析程序子命令
        > * kubelet并不支持子命令，将直接报错并结束程序
  
    * `cleanFlagSet.GetBool("help")`

      > 判断是否包含`--help`，有则直接跳转到`kubelet help`，结束程序

    * `verflag.PrintAndExitIfRequested()`

      >* 解析`--version`字段
      >
      >* 内部基于`type versionValue int`手动约定枚举类型`versionValue`
      >
      >* `var versionFlag = Version(versionFlagName, VersionFalse, "Print version information and quit")` 
      >
      >  > 实质上就是基于`pflag`将一个`*versionValue`类型的值绑定到名为`versionFlagName`的`flag`,其默认值为`VersionFalse`，真实值存储在`versionFlag`
      >  >
      >
      >* 使用`AddFlags(fs *flag.FlagSet)`将此处声明的`flag`注册到`fs`中，即可接受到程序真实的参数输入
      >
      >* 传入值：
      >
      >  > * `"raw"` : 打印`版本号`，结束进程
      >  >
      >  >   > 例如：
      >  >   >
      >  >   > ```shell
      >  >   > version.Info{Major:"1", Minor:"23", GitVersion:"v1.23.1", GitCommit:"86ec240af8cbd1b60bcc4c03c20da9b98005b92e", GitTreeState:"clean", BuildDate:"2021-12-16T11:39:51Z", GoVersion:"go1.17.5", Compiler:"gc", Platform:"linux/amd64"}
      >  >   > ```
      >  >
      >  > * `"true"`: 打印`程序名（此处默认为字符串"Kubernetes"）+版本号`，结束进程
      >  >
      >  >   > 例如：
      >  >   >
      >  >   > ```shell
      >  >   > Kubernetes v1.23.1
      >  >   > ```
      >  >
      >  > * `"false"`: 直接结束函数，无操作
      >
      >额外补充：
      >
      >```go
      >// 此接口为pflag包内容
      >type Value interface {
      >	String() string
      >	Set(string) error
      >	Type() string
      >}
      >
      >// 注意：可见到在verflag.go中对约定的枚举类型versionValue实现了pflag.Value接口
      >
      >// verflag.go: line 77~78
      >*p = value					// 这里的p可以是自定义类型，只要实现了pflag.Value接口就行。此处实质上是*Int（手动约定枚举类型）
      >flag.Var(p, name, usage)	// 即定义一个flag标志，与p绑定，默认值为value，名称为name，用法为usage
      >
      >// *p = VersionFalse		// VersionFalse = 0			
      >// flag.Var(p, "version", "Print version information and quit")
      >```
    
    * `utilfeature.DefaultMutableFeatureGate.SetFromMap(kubeletConfig.FeatureGates)`
    
      > `kubeletConfig.FeatureGates`为`map[string]bool`类型，存储了k8s alpha/experimental版本特性是否启用
      >
      > 此函数作用：
      >
      > * 用`kubeletConfig.FeatureGates`覆盖默认的`FeatureGates`参数项
      >   * 内部核心实现涉及`know：Map`（存储了所有已知的特性及其描述）和`enable: Map`（存储了实际的特性及其开关状态）
      > * 拦截对未知特性的设置
      > * 拦截对禁止修改（通过`know[Feature_name].LockToDefault : bool`进行判定）的特性的设置
    
    * `options.ValidateKubeletFlags(kubeletFlags)`
    
      > * 验证是否有非`kubelet`支持的flag（该结构决定存储的flag一定是`kubernetes`支持的，但不一定是`kubelet`支持）
      > * 验证flag的设置是否与其它选项的设置冲突，例如`FeatureGate`（可能将某特性设置为禁用）
    
    * 对容器运行时为`remote`时可能出现的错误给出提示
    
    * 从磁盘文件`kubelet.kubeconfig`读取配置文件
    
      * `DefaultFs`
    
        > `utilfs.DefaultFs`实质上是对默认OS文件操作的一种封装，目的是可以自动的在所有传入的文件path前面自动加上一个`root`路径
    
        代码：
    
        ```go
        type DefaultFs struct {
        	root string
        }
        
        // 实现举例
        func (fs *DefaultFs) Remove(name string) error {
            real_name := filepath.Join(fs.root, name)		// 自动加上前缀
            
            return os.Remove(real_name)
        }
        // ...
        ```
    
        



