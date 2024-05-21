# 通用 AIS 数据接收处理器

> [!WARNING]
> 这个组件是针对一款特殊AIS设备开发的专有支持，不具备通用性，请不要随意在生产中使用！
## 简介
AIS是自动识别系统（Automatic Identification System）的缩写。它是一种用于船舶和船舶交通管理的技术，旨在提高航海安全、保护环境和增强航运效率。

AIS系统通过使用卫星和地面设备，以无线电通信的方式，自动地交换船舶的位置、速度、航向和其他相关信息。这些信息可以被附近的其他船只、陆地基站和航管部门接收和解读。AIS系统广泛应用于海洋航行和船舶交通管制中，以提供实时的船舶动态监测和管理。

通过AIS系统，船舶可以相互检测并识别彼此的位置、航向和速度，从而帮助船舶避免碰撞和保持安全距离。此外，船舶管理部门和港口管理机构也可以利用AIS数据来监测和调度船舶交通，提高港口运作的效率和安全性。

总的来说，AIS是一种通过无线电通信传输船舶相关信息的系统，用于提高航海安全、船舶交通管理和港口运作效率。

## AIS 规范
以下是一个示例AIS数据报文，包含标签信息：

```
!AIVDM,1,1,,A,13P1wPIP000JcMDJ`R5mriW000S?,0*27
```

在这个示例中，报文以"!AIVDM"作为标识符开头，表明这是一个AIS数据报文。其余部分是报文的内容，由逗号分隔的字段组成。

在这个报文中，标签信息位于第6个字段，表示消息类型。根据示例中的报文，消息类型为"A"，对应的十进制值为10。根据AIS协议的定义，消息类型10代表船舶静态和船舶详细信息。

请注意，示例报文中的其他字段表示报文的序列号、发射模式、二进制数据和校验和等。这些字段的含义和结构需要根据AIS协议规范进行解析和处理。

AIS报文的前导部分是位于报文开头的一部分，它通常包含了一些自定义的信息或附加数据，而不是标准的AIS消息内容。前导部分的具体含义可以因不同的系统、设备或应用而异。

在提供的示例报文 `\1G1:370208949,t:2320,c:1660780800*55` 中，前导部分为 `\1G1:370208949,t:2320,c:1660780800`。请注意，这个前导部分并不是标准的AIS报文格式，它可能是特定应用或系统中定义的自定义信息。

根据前导部分的内容推测其含义：

- `\1G1:370208949`：这可能是一个自定义的标识符或设备编号，用于标识报文的来源或发送方。
- `t:2320`：这可能表示一个时间戳，其中 2320 可能代表某种时间或计数值。
- `c:1660780800`：这可能是一个校验和或校验码，用于验证报文的完整性和准确性。

需要注意的是，前导部分的含义和格式可以因特定的应用、系统或设备而有所不同。为了准确理解前导部分的含义，建议参考相关的应用或系统文档，或联系相关的设备制造商或开发者以获取详细信息。

## 消息类型
AIS（Automatic Identification System）是一种用于船舶自动识别和通信的系统，它可以通过无线电信号在船舶之间传递信息。以下是一些常见的AIS消息类型及其示例：

1. **位置报告消息 (Position Report)**:
   - 消息类型: 1, 2, 3
   - 示例: `!AIVDM,1,1,,A,13aG;2P001G?U<jDUBEP1wUoP06,0*54`

2. **静态和船位信息消息 (Static and Voyage-Related Data)**:
   - 消息类型: 5, 24
   - 示例: `!AIVDM,2,1,2,B,55MsP00kLC7L7R?v8d4d<fn` 或 `!AIVDM,1,1,,B,H3OUnT@0T3Uto9w`

3. **航行状态消息 (Voyage-Related Data)**:
   - 消息类型: 5, 24
   - 示例: `!AIVDM,2,1,2,B,55MsP00kLC7L7R?v8d4d<fn` 或 `!AIVDM,1,1,,B,H3OUnT@0T3Uto9w`

4. **标准地区通知消息 (Standard SAR Aircraft Position Report)**:
   - 消息类型: 9
   - 示例: `!AIVDM,1,1,,A,963OwjP000G?P0bEmoV00000000,0*7A`

5. **AIS Base Station Broadcast消息**:
   - 消息类型: 4
   - 示例: `!AIVDM,1,1,,B,400jlu?P00G@0pQKWs6QGwvL0H0;,0*7F`

6. **AIS Binary Message消息**:
   - 消息类型: 6
   - 示例: `!AIVDM,1,1,,A,63Rjpm0<1` 或 `!AIVDM,1,1,,A,8D5Mm9h0,2*5F`

7. **AIS Broadcast消息**:
   - 消息类型: 14
   - 示例: `!AIVDM,1,1,,A,H4aOvj0026aIpj`:

8. **目标状态消息 (Safety-Related Message)**:
   - 消息类型: 18
   - 示例: `!AIVDM,1,1,,A,853OvPP02F91ACPFJ5Dr:0<4h@E`
## 区别
AIVDM (Automatic Identification System Data Message) 和 AIVDO (Automatic Identification System Data Object) 是两种在自动识别系统 (AIS) 中使用的数据格式。AIS 是一个用于船舶间通信的系统，它可以帮助船舶彼此识别和交换信息。 AIVDM 和 AIVDO 的主要区别在于它们所传输的信息内容和格式。
### 信息内容：
- AIVDM：AIVDM 消息用于传输更复杂的数据，例如船舶的详细信息、位置、速度、航行状态等。
- AIVDO：AIVDO 对象用于传输简单的数据，例如文本消息、图像或音频文件等。
### 数据格式：
- AIVDM：AIVDM 消息使用一种称为“分块”(segmenting) 的方法来传输数据。数据被分成多个块，每个块都包含有关消息类型、消息长度和其他元数据的信息。
- AIVDO：AIVDO 对象使用一种更加简单和直接的数据格式。它们不使用分块方法，而是直接传输数据。

### 应用场景：
- AIVDM：由于其传输的复杂数据，AIVDM 消息通常在需要交换大量船舶信息的情况下使用，例如在港口或近海区域。
- AIVDO：AIVDO 对象更适合用于传输少量的数据，例如在简单的文本消息或文件传输情况下。
## 设备支持：
- AIVDM：许多 AIS 设备都支持 AIVDM 消息，因为它是在 AIS 系统中最常用的数据格式。
- AIVDO：虽然一些 AIS 设备可能支持 AIVDO，但它不如 AIVDM 普及，并且通常在需要传输特定类型数据的情况下使用。

以上只是一些常见的AIS消息类型及其示例。实际上，AIS有多种不同类型的消息，每种消息类型都包含特定的信息，例如船舶位置、速度、航向、船名等。消息的格式和字段根据AIS规范进行编码。

## 配置
本插件是基于TCP来传输AIS报文，公共配置如下：
```go
type _AISDeviceMasterConfig struct {
	Mode     string `json:"mode"` // TCP UDP UART
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required"`
	ParseAis bool   `json:"parseAis"`
}
```

## 规则脚本
```lua
Actions = { function(args)
    local error1, JsonT = json:J2T(args)
    local t =
    {
        id = string:MakeUid(),
        method = "thing.event.property.post",
        params = {
            ais_data = JsonT['ais_data']
        }
    }
    local jsons = json:T2J(t)
    local error = data:ToMqtt('OUTQAQXBVCU', jsons)
    if error ~= nil then
        stdlib:Throw(error)
    end
    return true, args
end }
```
## 数据样例
假设有一个设备发送了GNS报文,经过RHILEX解析以后，规则脚本里面的 data 格式如下:
- RMC数据
   ```json
   {
      "type":"RMC",
      "gwid":"HR0001",
      "validity":"A",
      "latitude":48.11729999999999,
      "longitude":11.516666666666667,
      "speed":22.4,
      "course":84.4,
      "date":"23-03-94 12:35:19.0000",
      "variation":-3.1,
      "ffa_mode":"",
      "nav_status":""
   }
   ```
- VDM数据
   ```json
   {
      "type":"VDM",
      "gwid":"HR0001",
      "message_id":19,
      "user_id":413825345,
      "name":"YUXINHUO16626",
      "sog":3.2,
      "longitude":114.347,
      "latitude":30.62909,
      "cog":226.3,
      "true_heading":511,
      "timestamp":35
   }
   ```

## 注意
本功能实际上是专门给某一款串口射频天线写的，并不具备通用性。