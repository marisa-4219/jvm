# JVM

### 用于管理/动态切换本地JDK版本的工具, 这是最简单也很可靠的实现。

### 使用前强烈建议建议保存环境变量的Path

###### 对于JDK动态下载的功能是没有实现的（暂时?）

编译为可执行文件, 在终端运行（管理员权限） `./jvm.exe install` 即可将工具注册到本地环境

- 安装jvm到环境变量<br>
 `jvm install - install jvm to env `<br>
- 切换JDK<br>
 `jvm use     - <jdk version> change jdk `<br>
- 在命令行内修改store.json中的sdk目录/虚拟的`JAVA_HOME`目录（映射文件）<br>
 `jvm set     - <property> <value> set store.json property [sdk_path|virtual_java_home] `<br>
- 查看store.json<br>
 `jvm store   - show store `
- 搜索目录下的所有jdk<br>
 `jvm search  - list sdk_path all JDK `
- 搜索目录下的所有jdk并保存到store.json<br>
 `jvm update  - save sdk_path all JDK to store.json`
- 查看帮助<br>
 `jvm help    - show help `

### 安装流程

- 编译为可执行文件, 在终端运行（管理员权限） `./jvm.exe install` 并重启终端（让环境变量生效）
- 设置SDK目录 `jvm set sdk_path D:/sdk/...` 只需要到jdk的上级目录即可
- 尝试搜索JDK `jvm search` 查看是否能够正常输出JDK列表
  ```
     F:\Java\jvm> jvm search 
     1.8.0_171 -> F:\Java\sdk\8
     11.0.11   -> F:\Java\sdk\11
     17.0.1    -> F:\Java\sdk\17
     19        -> F:\Java\sdk\19
     21        -> F:\Java\sdk\21
  ```
- 如果上一步可以正常输出则保存到 `store.json` 执行 `jvm update`
- 使用 `jvm store` 查看当前 `store.json` 内容
    ```
    F:\Java\jvm> jvm store
    Store:
    {
    "powershell": "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
    "sdk_path": "F:/Java/sdk",
    "sdk": {
    "1.8.0_171": "F:\\Java\\sdk\\8",
    "11.0.11": "F:\\Java\\sdk\\11",
    "17.0.1": "F:\\Java\\sdk\\17",
    "19": "F:\\Java\\sdk\\19",
    "21": "F:\\Java\\sdk\\21"
    },
    "virtual_java_home": "F:\\Java\\jvm\\VirtualJavaSDK"
    }
    ```
- 切换jdk `jvm use 1.8.0_171` 如果你是第一次切换请重启你的终端（让环境变量生效, 后续是变更映射文件，将不再需要重启终端）
