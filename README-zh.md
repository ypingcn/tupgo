# TUPGo

[English](./README.md)

## 用法

1. 使用 tars2go 转换 TEST.tars 文件

```
module TEST
{
    struct S
    {
        0 optional int a;
    };
};
```

2. 使用 TUPGo

```golang
    var buffer []byte 
    var err error

    s1 := TEST.S{A : 100}
    req := TUPGo.TarsUniPacket{}
    req.Init()
    req.SetServantName("test")
    req.SetFuncName("test")
    err = req.Put("tReq",&s1)
    if err != nil {
        log.Fatalln(err)
    }
    buffer, err = req.Encode()
    if err != nil {
        log.Fatalln(err)
    }

    s2 := TEST.S{}
    rsp := TUPGo.TarsUniPacket{}
    rsp.Init()
    err = rsp.Decode(buffer)
    if err != nil {
        log.Fatalln(err)
    }
    err = rsp.Get("tReq",&s2)
    if err != nil {
        log.Fatalln(err)
    }
    fmt.Println(s2.A)
```

## 已知问题

1. 当前不支持非int类型的接口返回值