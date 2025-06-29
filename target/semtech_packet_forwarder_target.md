## Semtech UDP 协议

	  ______                              _
	 / _____)             _              | |
	( (____  _____ ____ _| |_ _____  ____| |__
	 \____ \| ___ |    (_   _) ___ |/ ___)  _ \
	 _____) ) ____| | | || |_| ____( (___| | | |
	(______/|_____)_|_|_| \__)_____)\____)_| |_|
	  (C)2013 Semtech-Cycleo

Basic communication protocol between Lora gateway and server
=============================================================


1. Introduction
----------------

The protocol between the gateway and the server is purposefully very basic and
for demonstration purpose only, or for use on private and reliable networks.

There is no authentication of the gateway or the server, and the acknowledges
are only used for network quality assessment, not to correct UDP datagrams
losses (no retries).


2. System schematic and definitions
------------------------------------

	 ((( Y )))
	     |
	     |
	+ - -|- - - - - - - - - - - - - +        xxxxxxxxxxxx          +--------+
	| +--+-----------+     +------+ |       xx x  x     xxx        |        |
	| |              |     |      | |      xx  Internet  xx        |        |
	| | Concentrator |<--->| Host |<-------xx     or    xx-------->|        |
	| |              | SPI |      | |      xx  Intranet  xx        | Server |
	| +--------------+     +------+ |       xxxx   x   xxxx        |        |
	|    ^                     ^    |           xxxxxxxx           |        |
	|    | PPS +-------+ NMEA  |    |                              |        |
	|    +-----|  GPS  |-------+    |                              +--------+
	|          | (opt) |            |
	|          +-------+            |
	|                               |
	|             Gateway           |
	+- - - - - - - - - - - - - - - -+

__Concentrator__: radio RX/TX board, based on Semtech multichannel modems
(SX130x), transceivers (SX135x) and/or low-power stand-alone modems (SX127x).

__Host__: embedded computer on which the packet forwarder is run. Drives the
concentrator through a SPI link.

__GPS__: GNSS (GPS, Galileo, GLONASS, etc) receiver with a "1 Pulse Per Second"
output and a serial link to the host to send NMEA frames containing time and
geographical coordinates data. Optional.

__Gateway__: a device composed of at least one radio concentrator, a host, some
network connection to the internet or a private network (Ethernet, 3G, Wifi,
microwave link), and optionally a GPS receiver for synchronization.

__Server__: an abstract computer that will process the RF packets received and
forwarded by the gateway, and issue RF packets in response that the gateway
will have to emit.

It is assumed that the gateway can be behind a NAT or a firewall stopping any
incoming connection.
It is assumed that the server has an static IP address (or an address solvable
through a DNS service) and is able to receive incoming connections on a
specific port.


3. Upstream protocol
---------------------

### 3.1. Sequence diagram ###

	+---------+                                                    +---------+
	| Gateway |                                                    | Server  |
	+---------+                                                    +---------+
      | -----------------------------------\ |
      | ------------------------------------ || When 1-N RF packets are received                               |                             |
| -------------------------------------------------------------- |
|                                                                |
| PUSH_DATA (token X, GW MAC, JSON payload)                      |
| -------------------------------------------------------------> |
|                                                                |
| PUSH_ACK (token X)                                             |
| <------------------------------------------------------------- |
| ------------------------------\                                |
|                                                                | process packets *after* ack | - |
| -------------------------------                                |
|                                                                |

### 3.2. PUSH_DATA packet ###

That packet type is used by the gateway mainly to forward the RF packets
received, and associated metadata, to the server.

 | Bytes  | Function                                                   |
 | :----: | ---------------------------------------------------------- |
 |   0    | protocol version = 2                                       |
 |  1-2   | random token                                               |
 |   3    | PUSH_DATA identifier 0x00                                  |
 |  4-11  | Gateway unique identifier (MAC address)                    |
 | 12-end | JSON object, starting with {, ending with }, see section 4 |

### 3.3. PUSH_ACK packet ###

That packet type is used by the server to acknowledge immediately all the
PUSH_DATA packets received.

 | Bytes | Function                                          |
 | :---: | ------------------------------------------------- |
 |   0   | protocol version = 2                              |
 |  1-2  | same token as the PUSH_DATA packet to acknowledge |
 |   3   | PUSH_ACK identifier 0x01                          |


4. Upstream JSON data structure
--------------------------------

The root object can contain an array named "rxpk":

``` json
{
	"rxpk":[ {...}, ...]
}
```

That array contains at least one JSON object, each object contain a RF packet
and associated metadata with the following fields:

 | Name  |  Type  | Function                                                     |
 | :---: | :----: | ------------------------------------------------------------ |
 | time  | string | UTC time of pkt RX, us precision, ISO 8601 'compact' format  |
 | tmms  | number | GPS time of pkt RX, number of milliseconds since 06.Jan.1980 |
 | tmst  | number | Internal timestamp of "RX finished" event (32b unsigned)     |
 | freq  | number | RX central frequency in MHz (unsigned float, Hz precision)   |
 | chan  | number | Concentrator "IF" channel used for RX (unsigned integer)     |
 | rfch  | number | Concentrator "RF chain" used for RX (unsigned integer)       |
 | stat  | number | CRC status: 1 = OK, -1 = fail, 0 = no CRC                    |
 | modu  | string | Modulation identifier "LORA" or "FSK"                        |
 | datr  | string | LoRa datarate identifier (eg. SF12BW500)                     |
 | datr  | number | FSK datarate (unsigned, in bits per second)                  |
 | codr  | string | LoRa ECC coding rate identifier                              |
 | rssi  | number | RSSI in dBm (signed integer, 1 dB precision)                 |
 | lsnr  | number | Lora SNR ratio in dB (signed float, 0.1 dB precision)        |
 | size  | number | RF packet payload size in bytes (unsigned integer)           |
 | data  | string | Base64 encoded RF packet payload, padded                     |

Example (white-spaces, indentation and newlines added for readability):

``` json
{"rxpk":[
	{
		"time":"2013-03-31T16:21:17.528002Z",
		"tmst":0,
		"chan":2,
		"rfch":0,
		"freq":866.349812,
		"stat":1,
		"modu":"LORA",
		"datr":"SF7BW125",
		"codr":"4/6",
		"rssi":-35,
		"lsnr":5.1,
		"size":32,
		"data":"-DS4CGaDCdG+48eJNM3Vai-zDpsR71Pn9CPA9uCON84"
	},{
		"time":"2013-03-31T16:21:17.530974Z",
		"tmst":3512348514,
		"chan":9,
		"rfch":1,
		"freq":869.1,
		"stat":1,
		"modu":"FSK",
		"datr":50000,
		"rssi":-75,
		"size":16,
		"data":"VEVTVF9QQUNLRVRfMTIzNA=="
	},{
		"time":"2013-03-31T16:21:17.532038Z",
		"tmst":3316387610,
		"chan":0,
		"rfch":0,
		"freq":863.00981,
		"stat":1,
		"modu":"LORA",
		"datr":"SF10BW125",
		"codr":"4/7",
		"rssi":-38,
		"lsnr":5.5,
		"size":32,
		"data":"ysgRl452xNLep9S1NTIg2lomKDxUgn3DJ7DE+b00Ass"
	}
]}
```

The root object can also contain an object named "stat" :

``` json
{
	"rxpk":[ {...}, ...],
	"stat":{...}
}
```

It is possible for a packet to contain no "rxpk" array but a "stat" object.

``` json
{
	"stat":{...}
}
```

That object contains the status of the gateway, with the following fields:

 | Name  |  Type  | Function                                                     |
 | :---: | :----: | ------------------------------------------------------------ |
 | time  | string | UTC 'system' time of the gateway, ISO 8601 'expanded' format |
 | lati  | number | GPS latitude of the gateway in degree (float, N is +)        |
 | long  | number | GPS latitude of the gateway in degree (float, E is +)        |
 | alti  | number | GPS altitude of the gateway in meter RX (integer)            |
 | rxnb  | number | Number of radio packets received (unsigned integer)          |
 | rxok  | number | Number of radio packets received with a valid PHY CRC        |
 | rxfw  | number | Number of radio packets forwarded (unsigned integer)         |
 | ackr  | number | Percentage of upstream datagrams that were acknowledged      |
 | dwnb  | number | Number of downlink datagrams received (unsigned integer)     |
 | txnb  | number | Number of packets emitted (unsigned integer)                 |

Example (white-spaces, indentation and newlines added for readability):

``` json
{"stat":{
	"time":"2014-01-12 08:59:28 GMT",
	"lati":46.24000,
	"long":3.25230,
	"alti":145,
	"rxnb":2,
	"rxok":2,
	"rxfw":2,
	"ackr":100.0,
	"dwnb":2,
	"txnb":2
}}
```


5. Downstream protocol
-----------------------

### 5.1. Sequence diagram ###

	+---------+                                                    +---------+
	| Gateway |                                                    | Server  |
	+---------+                                                    +---------+
      | -----------------------------------\ |
      | ------------------------------------ || Every N seconds (keepalive time)                               |  |
| -------------------------------------------------------------- |
|                                                                |
| PULL_DATA (token Y, MAC@)                                      |
| -------------------------------------------------------------> |
|                                                                |
| PULL_ACK (token Y)                                             |
| <------------------------------------------------------------- |
|                                                                |

	+---------+                                                    +---------+
	| Gateway |                                                    | Server  |
	+---------+                                                    +---------+
	     |      ------------------------------------------------------\ |
      |                                                                | Anytime after first PULL_DATA for each packet to TX | - |
      | -------------------------------------------------------------- |
      |                                                                |
      | PULL_RESP (token Z, JSON payload)                              |
      | <------------------------------------------------------------- |
      |                                                                |
      | TX_ACK (token Z, JSON payload)                                 |
      | -------------------------------------------------------------> |

### 5.2. PULL_DATA packet ###

That packet type is used by the gateway to poll data from the server.

This data exchange is initialized by the gateway because it might be
impossible for the server to send packets to the gateway if the gateway is
behind a NAT.

When the gateway initialize the exchange, the network route towards the
server will open and will allow for packets to flow both directions.
The gateway must periodically send PULL_DATA packets to be sure the network
route stays open for the server to be used at any time.

 | Bytes | Function                                |
 | :---: | --------------------------------------- |
 |   0   | protocol version = 2                    |
 |  1-2  | random token                            |
 |   3   | PULL_DATA identifier 0x02               |
 | 4-11  | Gateway unique identifier (MAC address) |

### 5.3. PULL_ACK packet ###

That packet type is used by the server to confirm that the network route is
open and that the server can send PULL_RESP packets at any time.

 | Bytes | Function                                          |
 | :---: | ------------------------------------------------- |
 |   0   | protocol version = 2                              |
 |  1-2  | same token as the PULL_DATA packet to acknowledge |
 |   3   | PULL_ACK identifier 0x04                          |

### 5.4. PULL_RESP packet ###

That packet type is used by the server to send RF packets and associated
metadata that will have to be emitted by the gateway.

 | Bytes | Function                                                   |
 | :---: | ---------------------------------------------------------- |
 |   0   | protocol version = 2                                       |
 |  1-2  | random token                                               |
 |   3   | PULL_RESP identifier 0x03                                  |
 | 4-end | JSON object, starting with {, ending with }, see section 6 |

### 5.5. TX_ACK packet ###

That packet type is used by the gateway to send a feedback to the server
to inform if a downlink request has been accepted or rejected by the gateway.
The datagram may optionnaly contain a JSON string to give more details on
acknoledge. If no JSON is present (empty string), this means than no error
occured.

 | Bytes  | Function                                                              |
 | :----: | --------------------------------------------------------------------- |
 |   0    | protocol version = 2                                                  |
 |  1-2   | same token as the PULL_RESP packet to acknowledge                     |
 |   3    | TX_ACK identifier 0x05                                                |
 |  4-11  | Gateway unique identifier (MAC address)                               |
 | 12-end | [optional] JSON object, starting with {, ending with }, see section 6 |

6. Downstream JSON data structure
----------------------------------

The root object of PULL_RESP packet must contain an object named "txpk":

``` json
{
	"txpk": {...}
}
```

That object contain a RF packet to be emitted and associated metadata with the following fields:

 | Name  |  Type  | Function                                                         |
 | :---: | :----: | ---------------------------------------------------------------- |
 | imme  |  bool  | Send packet immediately (will ignore tmst & time)                |
 | tmst  | number | Send packet on a certain timestamp value (will ignore time)      |
 | tmms  | number | Send packet at a certain GPS time (GPS synchronization required) |
 | freq  | number | TX central frequency in MHz (unsigned float, Hz precision)       |
 | rfch  | number | Concentrator "RF chain" used for TX (unsigned integer)           |
 | powe  | number | TX output power in dBm (unsigned integer, dBm precision)         |
 | modu  | string | Modulation identifier "LORA" or "FSK"                            |
 | datr  | string | LoRa datarate identifier (eg. SF12BW500)                         |
 | datr  | number | FSK datarate (unsigned, in bits per second)                      |
 | codr  | string | LoRa ECC coding rate identifier                                  |
 | fdev  | number | FSK frequency deviation (unsigned integer, in Hz)                |
 | ipol  |  bool  | Lora modulation polarization inversion                           |
 | prea  | number | RF preamble size (unsigned integer)                              |
 | size  | number | RF packet payload size in bytes (unsigned integer)               |
 | data  | string | Base64 encoded RF packet payload, padding optional               |
 | ncrc  |  bool  | If true, disable the CRC of the physical layer (optional)        |

Most fields are optional.
If a field is omitted, default parameters will be used.

Examples (white-spaces, indentation and newlines added for readability):

``` json
{"txpk":{
	"imme":true,
	"freq":864.123456,
	"rfch":0,
	"powe":14,
	"modu":"LORA",
	"datr":"SF11BW125",
	"codr":"4/6",
	"ipol":false,
	"size":32,
	"data":"H3P3N2i9qc4yt7rK7ldqoeCVJGBybzPY5h1Dd7P7p8v"
}}
```

``` json
{"txpk":{
	"imme":true,
	"freq":861.3,
	"rfch":0,
	"powe":12,
	"modu":"FSK",
	"datr":50000,
	"fdev":3000,
	"size":32,
	"data":"H3P3N2i9qc4yt7rK7ldqoeCVJGBybzPY5h1Dd7P7p8v"
}}
```

The root object of TX_ACK packet must contain an object named "txpk_ack":

``` json
{
	"txpk_ack": {...}
}
```

That object contain status information concerning the associated PULL_RESP packet.

 | Name  |  Type  | Function                                                                       |
 | :---: | :----: | ------------------------------------------------------------------------------ |
 | error | string | Indication about success or type of failure that occured for downlink request. |

The possible values of "error" field are:

 |      Value       | Definition                                                                    |
 | :--------------: | ----------------------------------------------------------------------------- |
 |       NONE       | Packet has been programmed for downlink                                       |
 |     TOO_LATE     | Rejected because it was already too late to program this packet for downlink  |
 |    TOO_EARLY     | Rejected because downlink packet timestamp is too much in advance             |
 | COLLISION_PACKET | Rejected because there was already a packet programmed in requested timeframe |
 | COLLISION_BEACON | Rejected because there was already a beacon planned in requested timeframe    |
 |     TX_FREQ      | Rejected because requested frequency is not supported by TX RF chain          |
 |     TX_POWER     | Rejected because requested power is not supported by gateway                  |
 |   GPS_UNLOCKED   | Rejected because GPS is unlocked, so GPS timestamp cannot be used             |

Examples (white-spaces, indentation and newlines added for readability):

``` json
{"txpk_ack":{
	"error":"COLLISION_PACKET"
}}
```
## Demo Data
```json
{
  "time": "2023-04-01T12:34:56Z",
  "tmst": 1234567890,
  "chan": 3,
  "rfch": 1,
  "freq": 868.1,
  "stat": 1,
  "modu": "LORA",
  "datr": "SF7BW125",
  "codr": "4/5",
  "rssi": -50,
  "lsnr": 7.5,
  "size": 32,
  "data": "0102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F20"
}
```

这个数据表示：
- `Time` 是一个UTC时间戳，表示数据包被接收的时间。
- `Tmst` 是一个整数，表示数据包被接收的时间，单位是毫秒。
- `Chan` 是一个整数，表示接收数据包的通道号。
- `Rfch` 是一个整数，表示使用的RF链路号。
- `Freq` 是一个浮点数，表示接收数据包的频率，单位是MHz。
- `Stat` 是一个整数，表示接收状态。
- `Modu` 是一个字符串，表示调制方式，这里为"LoRa"。
- `Datr` 是一个字符串，表示数据速率，这里为"SF7BW125"。
- `Codr` 是一个字符串，表示编码率，这里为"4/5"。
- `Rssi` 是一个整数，表示接收信号强度指示，单位是dBm。
- `Lsnr` 是一个浮点数，表示信噪比，单位是dB。
- `Size` 是一个整数，表示数据包的大小，单位是字节。
- `Data` 是一个字符串，表示接收到的数据，这里为32字节的十六进制编码数据。

## 配置
```json
{
    "type": "SEMTECH_UDP_FORWARDER",
    "name": "SEMTECH_UDP_FORWARDER Server",
    "description": "SEMTECH_UDP_FORWARDER Server",
    "config": {
        "host": "192.168.10.103",
        "port": 1700,
        "mac": "0001020304050607"
    }
}
```