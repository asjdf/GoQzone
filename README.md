# GoQzone

<img src="https://i.loli.net/2021/07/26/u6o9gKdCzckMWlG.png" alt="logo" style="zoom:50%;" />

发送QQ空间说说的Go程序 （扩充功能中）



## 使用方法

引用

```go
import "GoQzone/pkg/goQzone"
```



新建会话，登录

```go
client := goQzone.Init()
if err := client.QrLogin();err != nil{
	panic(err)
}
```



发布

```go
err = client.NewPost().Content("说说内容").Pic(图片 []byte类型).Send()
if err != nil {
	fmt.Println(err)
}
```



其他发布类型

```
client.NewPost().Pic(图片 []byte类型).Send() // 只发一张图
client.NewPost().Content("说说内容").Send()  // 只有文字
client.NewPost().Pic(图片 []byte类型).Pic(图片 []byte类型).Send() // 多张图片
```

