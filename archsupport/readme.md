# 跨平台

这里放置一些对特定硬件的支持库，一般指的是定制化网关产品。如果有不同操作系统上的实现库，建议统一放置此处。可参考下面的hello文件里面的程序。

## 当前兼容

### RHILEXG1 网关

RHILEXG1 是 RHILEX 团队的默认硬件，操作系统为 `64位OpenWrt、Armbian`, CPU 架构为 `32位全志H3`。

### 树莓派4B+

除此之外，还对 `树莓派4B`的 GPIO 做了支持。树莓派的lua标准库命名空间为 `raspberry`。

## 环境变量

如果要启用硬件特性，需要在启动的时候加入 `ARCHSUPPORT` 环境变量来指定运行的版本, 例如要在 RHILEXG1-H3网关上运行：

```sh
ARCHSUPPORT=RHILEXG1 rhilex run
```

## 支持硬件列表

| 硬件名                    | 环境参数     | 示例                                  |
| ------------------------- | ------------ | ------------------------------------- |
| RHILEXG1 RHILEXG1版本网关 | RHILEXG1     | `ARCHSUPPORT=RHILEXG1 rhilex run`     |
| RHILEXG1 T507版本网关     | RHILEXG1T507 | `ARCHSUPPORT=RHILEXG1T507 rhilex run` |
| RHILEXG1 T113版本网关     | RHILEXG1T113 | `ARCHSUPPORT=RHILEXG1T113 rhilex run` |
| 树莓派4B、4B+             | RPI4         | `ARCHSUPPORT=RPI4B rhilex run`        |
| 玩客云S805                | WKYS805      | `ARCHSUPPORT=WKYS805 rhilex run`      |

> 警告: 这些属于板级高级功能，和硬件架构以及外设有关，默认关闭。 如果你自己需要定制，最好针对自己的硬件进行跨平台适配, 如果没有指定平台，可能会导致预料之外的结果。

## 常见函数

### RHILEXG1版本网关

1. GPIO 设置

   ```lua
   rhilexg1:GPIOSet(Pin, Value)
   ```
   参数表

   | 参数名 | 类型 | 说明           |
   | ------ | ---- | -------------- |
   | Pin    | int  | GPIO引脚       |
   | Value  | int  | 高低电平, 0、1 |
2. GPIO 获取

   ```lua
   rhilexg1:GPIOGet(Pin)
   ```
   | 参数名 | 类型 | 说明     |
   | ------ | ---- | -------- |
   | Pin    | int  | GPIO引脚 |

## 示例脚本
1. 玩客云WS1608
```lua
function Main(arg)
    while true do
        ws1608:GPIOSet("red", 1)
        time:Sleep(2000)
        ws1608:GPIOSet("red", 0)
        time:Sleep(2000)
    end
end

```
>必须使用这个系统：Linux aml-s812 5.9.0-rc7-aml-s812 #20.12 SMP Sun Dec 13 22:50:05 CST 2020 armv7l GNU/Linux, Armbian 20.12 Buster \l

1. RHILEXG1 网关
```lua
function Main(arg)
    while true do
        rhilexg1:GPIOSet(6, 1)
        time:Sleep(2000)
        rhilexg1:GPIOSet(7, 0)
        time:Sleep(2000)
    end
end

```