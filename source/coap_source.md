# CoAP Server
CoAP（Constrained Application Protocol）是一种为互联网上的受限节点（如传感器、开关等）设计的简单但强大的网络协议。它是基于REST原则的，类似于HTTP，但专为物联网（IoT）设备设计，这些设备通常具有有限的处理能力、内存和电源。
### CoAP的主要特点包括：
1. **简洁性**：为了适应受限设备，CoAP设计得尽可能简洁。
2. **低功耗**：通过减少传输的数据量和优化通信机制，CoAP有助于降低设备的功耗。
3. **支持多种网络类型**：包括传统的IP网络和专为IoT设计的6LoWPAN、RoLL等。
4. **可靠性**：支持确认消息和重传机制，确保数据可靠传输。
5. **发现和组播**：CoAP支持资源发现和多播，便于设备之间的相互识别和通信。
### CoAP的工作模式：
- **CON模式**：确保消息的可靠传输，适用于对可靠性要求较高的场景。
- **NON模式**：不保证消息的可靠传输，适用于对实时性要求较高但可以容忍一定程度丢包的场景。
### CoAP的消息类型：
- **请求**：包括GET、POST、PUT和DELETE，类似于HTTP的方法。
- **响应**：包括2.01 Created、2.02 Deleted、2.03 Valid、4.04 Not Found等，类似于HTTP的状态码。
### CoAP的应用场景：
- **物联网**：传感器网络、智能家居、工业控制等。
- **M2M通信**：机器之间的直接通信。
CoAP是物联网通信协议的重要组成部分，有助于推动物联网的发展和普及。随着技术的不断进步，CoAP在未来的应用场景可能会更加广泛。

## 数据

```json
{
    "ts": 1713691696641,
    "type": "POST",
    "payload": "012"
}
```

## 测试
示例客户端：
```go
package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"time"

	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/udp"
)

type ReadSeeker struct {
	io.ReadSeeker
}

func main() {
	co, err := udp.Dial("localhost:2582")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := co.Post(ctx, "/", message.AppOctets, bytes.NewReader([]byte{48, 49, 50}), message.Option{})
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	log.Printf("Response payload: %v", resp.Body())
}

```