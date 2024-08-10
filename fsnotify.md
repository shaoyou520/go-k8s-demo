go get github.com/fsnotify/fsnotify

fsnotify的使用比较简单：

* 先调用NewWatcher创建一个监听器；
* 然后调用监听器的Add增加监听的文件或目录；
* 如果目录或文件有事件产生，监听器中的通道Events可以取出事件。如果出现错误，监听器中的通道Errors可以取出错误信息。

```

package main

import (
  "log"

  "github.com/fsnotify/fsnotify"
)

func main() {
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    log.Fatal("NewWatcher failed: ", err)
  }
  defer watcher.Close()

  done := make(chan bool)
  go func() {
    defer close(done)

    for {
      select {
      case event, ok := <-watcher.Events:
        if !ok {
          return
        }
        log.Printf("%s %s\n", event.Name, event.Op)
      case err, ok := <-watcher.Errors:
        if !ok {
          return
        }
        log.Println("error:", err)
      }
    }
  }()

  err = watcher.Add("./")
  if err != nil {
    log.Fatal("Add failed:", err)
  }
  <-done
}
```

fsnotify.Event:
```
type Event struct {
  Name string
  Op   Op
}
```

```
type Op uint32

const (
  Create Op = 1 << iota
  Write
  Remove
  Rename
  Chmod
)
```
Chmod事件在文件或目录的属性发生变化时触发，在 Linux 系统中可以通过chmod命令改变文件或目录属性。

事件中的Op是按照位来存储的，可以存储多个，可以通过&操作判断对应事件是不是发生了。

```
if event.Op & fsnotify.Write != 0 {
  fmt.Println("Op has Write")
}
```
我们在代码中不需要以上判断，因为Op的String()方法已经帮我们处理了这种情况了：
```
func (op Op) String() string {
  // Use a buffer for efficient string concatenation
  var buffer bytes.Buffer

  if op&Create == Create {
    buffer.WriteString("|CREATE")
  }
  if op&Remove == Remove {
    buffer.WriteString("|REMOVE")
  }
  if op&Write == Write {
    buffer.WriteString("|WRITE")
  }
  if op&Rename == Rename {
    buffer.WriteString("|RENAME")
  }
  if op&Chmod == Chmod {
    buffer.WriteString("|CHMOD")
  }
  if buffer.Len() == 0 {
    return ""
  }
  return buffer.String()[1:] // Strip leading pipe
}

```