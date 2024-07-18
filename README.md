# ZBProxy plugin examples

ZBProxy 的模块化设计允许你注册自定义规则、嗅探协议。

本实例演示了在不修改 ZBProxy 本体代码、仅在项目中调用 ZBProxy 的情况下，实现一个公会专属加速IP。

使用时，修改 Hypixel API Key 和内嵌的配置中的公会名称，更新 ZBProxy 依赖版本，编译运行即可。

本示例仅演示最小功能实践，仅为演示模块化功能而编写，更多命令行功能（如加载文件配置、从数据库拉取）可额外添加。

## 如何新建一个基于 ZBProxy 的项目

命令行执行，引入 ZBProxy 依赖。

```shell
go get github.com/layou233/zbproxy/v3@dev-next
```

一切就绪。
