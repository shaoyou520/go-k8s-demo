# netlink 内核通信

## 什么是netlink？
netlink 是 Linux 系统里用户态程序、内核模块之间的一种 IPC 方式，  
特别是用户态程序和内核模块之间的 IPC 通信。比如在 Linux 终端里常用的 ip 命令，   
就是使用 netlink 去跟内核进行通信的。例如想在golang代码中实现ip link add xx的效果，  
一种办法是使用exec包执行对应的ip命令，另一种是采用netlink的方式， 
但是自己操作netlink还是有点繁琐。

Netlink 是 linux用户态程序用来与内核通信的接口。  
它可用于添加和删除接口、设置 ip 地址和路由以及配置 ipsec。  
Netlink 通信需要提升权限，因此在大多数情况下，此代码需要以 root 身份运行。

go get github.com/vishvananda/netlink

### func LinkSetNsPid(link Link, nspid int) error 
LinkSetNsPid 将设备放入新的网络命名空间。pid 必须是正在运行的进程的 pid。  
相当于：`ip link set $link netns $pid`

### func LinkSetNsFd(link Link, fd int) error
LinkSetNsFd 将设备放入新的网络命名空间。fd 必须是网络命名空间的打开文件描述符。  
类似于：`ip link set $link netns $ns`

### func LinkByName(name string ) ( Link，error )
LinkByName 通过名称查找链接并返回指向该对象的指针。
