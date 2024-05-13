## BACnet设备

### 设备type值
`GENERIC_BACNET_IP`
### 配置示例
#### 广播配置模式
```json
{
  "type": "BOARDCAST",
  "localIp": "192.168.1.50",
  "subnetCidr": 24,
  "localPort": 47808,
  "frequency": 10000
}
```

#### bacnetIp设备
```json
{
  "type": "SINGLE",
  "ip": "192.168.1.36",
  "port": 47808,
  "localPort": 47808,
  "frequency": 10000
}
```

#### bacnetMstp over ip设备
```json
{
  "type": "SINGLE",
  "ip": "192.168.1.36",
  "port": 47808,
  "isMstp": 1,
  "subnet": 20,
  "localPort": 47808,
  "frequency": 10000
}
```

> 字段含义

- type: bacnet采集驱动运行模式（可选SINGLE、BOARDCAST）
- ip:  bacnet设备ip（仅type=SINGLE时生效）
- port:  bacnet端口，通常是47808（仅type=SINGLE时生效）
- isMstp: 是否为mstp over ip设备，若是则子网号必须填写（仅type=SINGLE时生效）
- subnet: 虚拟子网号，一般是mstp转Ip网关用于给mstp组分配的一个网络号
- localIp: 本地地址（仅type=BOARDCAST时生效）
- subnetCidr: 子网掩码长度（localIp和subnetCidr共同计算出广播地址）
- localPort:  本地监听端口，填0表示默认47808
- frequency:  采集间隔，单位毫秒

### 点位配置
- tag: 点位名称
- alias: 别名
- bacnetDeviceId: 设备id，整数类型（若isMstp=1，则deviceId应该必填；若是bacnetip设备，则填1即可）
- objectType: 点位类型，必填（下拉框，枚举值）
- objectId: 对象id，必填（范围0-4194303)

`objectType`可选类型
* AI
* AO
* AV
* BI
* BO
* BV
* MI
* MO
* MV

### 采集后输出的数据格式
```json
{
  "tag1": "1",
  "tag2": "2",
  "tag3": "3"
}
```

### 注意事项
* 由于bacnet需要本地监听UDP端口用于收发信令，因此`localPort`不能配置重复
* 若使用广播模式`BOARDCAST`，则建议`localPort`设置为47808
