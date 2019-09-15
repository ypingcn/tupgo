# TUPGo

[中文](./README-zh.md)

## Usage

1. use tars2go for TEST.tars

```
module TEST
{
    struct S
    {
        0 optional int a;
    };
};
```

2. use TUPGo

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

## Known Issue

1. interface return value with not int type is not supported currently