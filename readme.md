# minion
内置通信模块 方便数据与中心端通信

## rock.push
- err = rock.push(mime , value)
- 向服务器端提交数据
- mime 数据编码
- value 数据值
```lua
    local mime = rock.require("mime")
    local err = rock.push(mime.xxx , aaa)
```

## rock.stream.kfk
- kfk = rock.stream.kfk{name , addr , topic}
- 利用tunnel代理链接远程地址, 满足lua.writer
- name 名称
- addr 远程地址
- topic 默认topic
### 内置接口
- [kfk.sdk(topic)]()  满足lua.writer 只是重定向topic
- [kfk.start()]()
```lua
    local kfk = rock.stream.kfk{
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

## rock.stream.tcp
- tcp = rock.stream.tcp{name , addr}
- 利用tunnel代理tcp链接 , 满足lua.writer
```lua
    local tcp = rock.stream.tcp{
        name = "tcp",
        addr = {"192.168.1.1:9092"}
    }
```


## rock.stream.sub
- sub = rock.stream.sub(name , addr)
- 暂无
