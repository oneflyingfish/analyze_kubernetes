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
>
> 说明：为了阅读方便，快速抓住核心，本文中的示例代码在源代码的基础上可能存在略微调整、伪代码化等等操作。如果阅读时存在排版疑问，请以文档原始编辑器(`typora v0.11.17 beta for windows`)为准

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
  * **集群各node节点之间不共享的配置集**

* 默认初始化kubeletConfig结构

  ```go
  kubeletConfig, err := options.NewKubeletConfiguration()
  ```

  * type: config.kubeletConfiguration

  * 作用：从配置文件`--kubeconfig=$PATH/kubelet.kubeconfig`获取的参数初始化结构

    * Node节点在CSR证书申请被批准后，自动在本地生成`kubelet.kubeconfig`，下次启动将根据此配置文件直接注册，而不用再次发起请求

  * **集群各node节点之间共享的配置集**

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
  
    * 从文件加载`KubeletConfig`参数
  
      > 先从配置文件读取，如果命令行同时指定其中某字段，优先级：命令行>配置文件
      >
      > 特例：针对`KubeletConfig.FeatureGates`，未冲突字段采用合并的方式，冲突字段优先级：命令行>配置文件
      >
      >
      > 说明：巧妙的用了一个临时`pflag.FlagSet`而非最终保留的`FlagSet`结构来对命令行参数进行再次解析，不会因此影响到`KubeletFlags`，避免了整个应用的重复解析问题。
      
      * 从磁盘文件`kubelet.kubeconfig`读取配置文件: `kubeletConfig,err = loadConfigFile(configFile)`
      
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

          * `loader := configfiles.NewFsLoader(...)`
      
            ```go
            func NewFsLoader(fs utilfs.Filesystem, kubeletFile string) (Loader, error) {
      
                // 此处初始化一个kubelet的默认文件编解码器：kubeletCodecs
      
                return &fsLoader{
                    fs:            fs,				// 即为DefaultFs类型
                    kubeletCodecs: kubeletCodecs,
                    kubeletFile:   kubeletFile,
                }, nil
            }
            ```
      
          * 数据读取以及格式化
      
            ```go
            func (loader *fsLoader) Load() (*kubeletconfig.KubeletConfiguration, error) {
                data, err := loader.fs.ReadFile(loader.kubeletFile)
        
                // ...
        
                kc, err := utilcodec.DecodeKubeletConfiguration(loader.kubeletCodecs, data)
        
                // ...
        
                // 读取kubeletconfig结构中所有路径字段的指针，形成 []*string
                paths := kubeletconfig.KubeletConfigurationPathRefs(kc)
        
                // 读取kubelet.kubeconfig文件所在目录，作为root目录
                root_dir := filepath.Dir(loader.kubeletFile)
        
                // 将kubeconfig结构中所有字段的目录，修改为：
                // *path = filepath.Join(root_dir, *path)
                resolveRelativePaths(paths, root_dir)
                return kc, nil
            }
            ```
      
      * `kubeletConfigFlagPrecedence(kc *kubeletconfiginternal.KubeletConfiguration, args []string)`
      
          > * 首先构造了一个假的全局`pflag.FlagSet`(实际上并不会使用，仅仅是局部变量)变量`fs`
          >* 以配置文件的`KubeletConfig`的值作为默认值向`fs`注册`flag`，并均标记为`Deprecated`
          > * `fs`解析命令行传入的所有参数，如果参数指定了`KubeletConfig`的值，将会覆盖上述默认值
          >* 针对`KubeletConfig.FeatureGates`取并集，冲突时优先级：命令行参数 > 配置文件
          > * 写回`KubeletConfig.FeatureGates`的原值
          >
          > 
          >说明：即针对`Kubeletconfig`参数配置优先级：命令行参数 > 配置文件，同时在命令行未特殊指定的情况下，保留原始的`KubeletConfig.FeatureGates`值（特性开启/关闭状态尽可能不变）。**该过程不会影响到`KubeletFlags`**的值
          >   
          >存在原因：为了解决issue#56171: https://github.com/kubernetes/kubernetes/issues/56171 （保证二进制版本向后兼容）
          
      * `newFlagSetWithGlobals()`
      
        > 实例化一个`*pflag.FlagSet`结构，拥有全局的`flag.FlagSet`（即`flag.CommandLine`)所拥有的所有`flags`(除了技术限制外，都被标记为`Deprecated`)
      
      * `newFakeFlagSet(...)`
      
        > 在`newFlagSetWithGlobals()`的基础上创建一个增强版`*pflag.FlagSet`结构，实质上仅仅把所有的Value绑定到了一个空结构体
      
        延伸参考：
      
        ```go
        // 对f中的所有flag以字母顺序或字典顺序执行：fn(flag)
        func (f *FlagSet) VisitAll(fn func(*Flag)){
            // ...
        }
        ```
      
      * `options.NewKubeletFlags().AddFlags(fs)`
  
        > * 实例化**一次性**的`options.KubeletFlags`结构，此处假定为`kf`
        > * 向`fs`注册了`kubelet`所有的`flags`, 值与`kf`的字段绑定，实质上放弃了对传入`KubeletFlags`参数值的读取
      
      * `options.AddKubeletConfigFlags(fs, kc)`
  
        > * 以配置文件的`KubeletConfig`的值作为默认值向`fs`注册`flag`，因此如果命令行参数重新制定，将会覆盖默认值，否则保持不变
        > * 因为在后期版本中，这些字段都被迁移到通过`--config=$file`中的`$file`指定，因此标记为`Deprecated`
      
      * 对于配置文件有值，但是命令行参数没有指定的`KubeletConfig.FeatureGates`参数，使用配置文件的值进行补充写回，达到合并目的。冲突字段则优先级：命令行参数 > 配置文件
      
        ```go
        for k, v := range original {
            if _, ok := kc.FeatureGates[k]; !ok {
                kc.FeatureGates[k] = v 					// 值不存在时原值写回，存在（即从命令行参数读取到对应的flag）则使用新值
            }
        }
        ```
      
      * `utilfeature.DefaultMutableFeatureGate.SetFromMap(kubeletConfig.FeatureGates)`
      
        > 由于上面更新了`Kubeconfig`的值，因此同步更新k8s alpha/experimental版本特性开闭状态
    
    * 验证`kubeletConfig`的内容是否合法：
    
      * `kubeletconfigvalidation.ValidateKubeletConfiguration(kubeletConfig)`保证内容格式合法
    
      * 确保`kubeletConfig.KubeReservedCgroup`为`kubeletConfig.KubeletCgroups`首部开始的子字符串，即`KubeReservedCgroup`路径为`KubeletCgroups`的子路径
    
        > 延伸阅读：
        >
        > `cgroup`（control group）是Linux内核的一项功能，提供了一系列资源管理控制器，由`systemd`自动挂载，用来控制进程对资源的分配，包括CPU、内存、网络带宽等
        >
        > * `kubeletConfig.KubeletCgroups`可由`--kubelet-cgroups`指定：创建和运行Kubelet的cgroups的绝对名称。
        > * `kubeletConfig.KubeReservedCgroup`可由`--kube-reserved-cgroup`指定：顶级cgroup的绝对名称，用于管理通过`--system-reserved`标志预留计算资源的非`kubernetes`组件，例如`"/system-reserverd"`默认为`""`
    
    * 动态`KubeletConfig`配置: `--dynamic-config-dir`指定，需将`KubeletConfig.FeatureGates`的`DynamicKubeletConfig`功能开启
    
      > 
