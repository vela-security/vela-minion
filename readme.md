# minion
内置通信模块 方便数据与中心端通信

## vela.push
- err = vela.push(mime , value)
- 向服务器端提交数据
- mime 数据编码
- value 数据值

### 扩展接口
- vela.push.sysinfo(value)
- vela.push.cpu (value)
- vela.push.disk(value)
- vela.push.listen(value)
- vela.push.memory(value)
- vela.push.socket(value)
- vela.push.network(value)
- vela.push.process(value)
- vela.push.service(value)
- vela.push.account(value)
- vela.push.filesystem(value)
- vela.push.task(value)
- vela.push.json(opcode ,value)

```lua
    local opcode = vela.require("opcode")
    local err = vela.push(mime.xxx , aaa)
```

## vela.stream.kfk
- kfk = vela.stream.kfk{name , addr , topic}
- 利用tunnel代理链接远程地址, 满足lua.writer
- name 名称
- addr 远程地址
- topic 默认topic
### 内置接口
- [kfk.sdk(topic)]()  满足lua.writer 只是重定向topic
- [kfk.start()]()
```lua
    local kfk = vela.stream.kfk{
        name  = "abcd",
        addr  = {"192.168.1.1:9092" , "192.168.1.10:9092"},
        topic = "aa"
    }
    kfk.start()
    kfk.push("x")
    kfk.push("2")

    local sdk = kfk.sdk("bb")
    sdk.push("b")
    sdk.push("c")
```

## vela.stream.tcp
- tcp = vela.stream.tcp{name , addr}
- 利用tunnel代理tcp链接 , 满足lua.writer
```lua
    local tcp = vela.stream.tcp{
        name = "tcp",
        addr = {"192.168.1.1:9092"}
    }
```


## vela.stream.sub
- sub = vela.stream.sub(name , addr)
- 暂无
