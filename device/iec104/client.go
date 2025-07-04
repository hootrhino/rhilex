// Copyright (C) 2024 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package iec104

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	contextTimeout    = 30 * time.Second
	dialTimeout       = 5 * time.Second
	testInterval      = 20 * time.Second
	totalCallInterval = 20 * time.Second
	retryTimes        = 3 // 存在备用服务器时，单个服务器重试次数
)

// Client 104客户端
type Client struct {
	address    string
	subAddress string
	curAddress string
	DataHandlerInterface
	conn        net.Conn
	IsConnected bool
	cancel      context.CancelFunc
	Logger      *logrus.Logger
	rsn         int16
	ssn         int16
	dataChan    chan *APDU
	sendChan    chan []byte
	iFrameNum   int
	task        func(c *APDU)
	wg          *sync.WaitGroup
	tlsConfig   *tls.Config
	needCert    bool
}

type DataHandlerInterface interface {
	Datahandler(*APDU) error
}

// NewClient 初始化客户端,连接失败，每隔10秒重试
func NewClient(handler DataHandlerInterface, address string, logger *logrus.Logger, tlsConfig *tls.Config, Cert bool) *Client {
	cli := &Client{
		address:    address,
		curAddress: address,
		dataChan:   make(chan *APDU, 1),
		sendChan:   make(chan []byte, 1),
		Logger:     logger,
		wg:         new(sync.WaitGroup),
		tlsConfig:  new(tls.Config),
		needCert:   Cert,
	}
	if tlsConfig != nil {
		cli.tlsConfig = tlsConfig
	}
	cli.DataHandlerInterface = handler
	return cli
}

// Run 运行
func (c *Client) Run() {
	if c.needCert == false {
		c.conn = c.dail()
	} else {
		conn, err := tls.Dial("tcp", c.curAddress, c.tlsConfig)
		if err != nil {
			i := 0
			for {
				conn, err = tls.Dial("tcp", c.curAddress, c.tlsConfig)
				if err != nil {
					time.Sleep(dialTimeout)
					i++
					c.Logger.Infof("连接服务器失败，开始第%d次重试", i+1)
				} else {
					c.Logger.Infoln("连接服务器成功")
					break
				}
			}
		}
		c.conn = conn
	}
	c.sendUFrame(startDtAct)
	c.IsConnected = true
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.wg.Add(3)
	go c.read(ctx)
	go c.write(ctx)
	go c.handler(ctx)
	c.rsn = 0
	c.ssn = 0
	c.iFrameNum = 0
}

// 建立tcp连接，支持重试和主备切换
func (c *Client) dail() net.Conn {
	var conn net.Conn
	var err error
	c.Logger.Infof("开始连接服务器:%v", c.curAddress)
	i := -1
	for {
		conn, err = net.DialTimeout("tcp", c.curAddress, dialTimeout)
		if err != nil {
			time.Sleep(dialTimeout)
			i++
			if i == retryTimes && c.subAddress != "" {
				i = 0
				if c.curAddress == c.address {
					c.curAddress = c.subAddress
				} else {
					c.curAddress = c.address
				}
				c.Logger.Infof("尝试超过3次，切换服务器为:%s,开始第%d次重试", c.curAddress, i+1)
			} else {
				c.Logger.Infof("连接服务器失败，开始第%d次重试", i+1)
			}
		} else {
			c.Logger.Infoln("连接服务器成功")
			break
		}
	}
	return conn
}

// Read 读数据
func (c *Client) read(ctx context.Context) {
	c.Logger.Info("socket读协程启动")
	defer func() {
		// c.cancel()
		c.wg.Done()
		c.Logger.Info("socket读协程停止")
	}()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := c.parseData(ctx)
			if err != nil {
				return
			}
		}
	}
}

// Write 写数据
func (c *Client) write(ctx context.Context) {
	c.Logger.Info("socket写协程启动")
	defer func() {
		// c.cancel()
		c.wg.Done()
		c.Logger.Info("socket写协程停止")
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-c.sendChan:
			_, err := c.conn.Write(data)
			if err != nil {
				return
			}
		}
	}
}

// handler 处理接收到的已解析数据
func (c *Client) handler(ctx context.Context) {
	c.Logger.Info("数据处理协程启动")
	defer func() {
		// c.cancel()
		c.wg.Done()
		c.Logger.Info("数据接收协程停止")
	}()
	for {
		select {
		case resp := <-c.dataChan:
			c.Logger.Debugf("接收到数据类型:%d,原因:%d,长度:%d", resp.ASDU.TypeID, resp.ASDU.Cause, len(resp.Signals))
			go c.Datahandler(resp)
		case <-ctx.Done():
			return
		}
	}
}

// ParseData 解析接收到的数据
func (c *Client) parseData(ctx context.Context) error {
	handleErr := func(tag string, err error) {
		c.Logger.Errorf("%s read socket读操作异常: %v", tag, err)
	}

	buf := make([]byte, 2)
	// 读取启动符和长度
	n, err := c.conn.Read(buf)
	if err != nil {
		handleErr("读取启动符和长度", err)
		return err
	}
	c.conn.SetDeadline(time.Now().Add(contextTimeout))
	if n == 0 {
		c.Logger.Info("读取到空数据,10s后继续读取数据")
		time.Sleep(10 * time.Second)
		return nil
	}
	length := int(buf[1])
	// 读取正文
	contentBuf := make([]byte, length)
	n, err = c.conn.Read(contentBuf)
	if err != nil {
		handleErr("读取正文", err)
		return err
	}
	// 长度不够继续读取，直至达到期望长度
	i := 1
	for n < length {
		i++
		nextLength := length - n
		nextBuf := make([]byte, nextLength)
		m, err := c.conn.Read(nextBuf)
		if err != nil {
			handleErr("循环读取正文", err)
			return err
		}
		contentBuf = append(contentBuf[:n], nextBuf[:m]...)
		n = len(contentBuf)
		c.Logger.Debugf("循环读取数据，当前为第%d次读取，期望长度:%d,本次长度:%d,当前总长度:%d", i, length, m, n)
	}
	c.Logger.Debugf("收到原始数据: [% X],rsn:%d,ssn:%d,长度:%d", append(buf, contentBuf[:n]...), c.rsn, c.ssn, 2+len(contentBuf[:n]))
	apdu := new(APDU)
	err = apdu.parseAPDU(contentBuf[:n])
	if err != nil {
		c.Logger.Warnf("解析APDU异常: %v", err)
		c.Logger.Panicln("退出程序")
		return err
	}
	switch apdu.CtrFrame.(type) {
	case IFrame:
		c.incrRsn()
		switch apdu.ASDU.TypeID {
		case MEiNA1:
			c.Logger.Info("接收到初始化结束，开始发送总召唤")
			c.sendSFrame()
			// c.SendTotalCall()
		case CIcNa1:
			if apdu.ASDU.Cause == 7 {
				c.Logger.Info("接收总召唤确认帧")
				c.sendSFrame()
			} else if apdu.ASDU.Cause == 10 {
				c.Logger.Info("接收总召唤结束帧")
				c.sendSFrame()
				c.Logger.Info("发送电度总召唤")
				c.SendElectricityTotalCall()
			}
		case CCiNa1:
			if apdu.ASDU.Cause == 7 {
				c.Logger.Info("接收电度总召唤确认帧")
			} else if apdu.ASDU.Cause == 10 {
				c.Logger.Info("接收电度总召唤结束帧")
			}
			c.sendSFrame()
		default:
			c.iFrameNum++
			c.Logger.Debugf("接收到第%d个I帧", c.iFrameNum)
			c.dataChan <- apdu
			c.sendSFrame()
		}
	case SFrame:
		c.Logger.Debugln("接收到S帧")
	case UFrame:
		c.Logger.Debugln("接收到U帧")
		uFrame := apdu.CtrFrame.(UFrame)
		switch uFrame.cmd {
		case startDtCon:
			c.Logger.Info("U帧为启动确认帧，发送总召唤")
			// c.SendTotalCall()
		case testFrAct:
			c.Logger.Info("U帧为测试激活帧,发送测试确认帧")
			c.sendUFrame(testFrCon)
		}
	default:
		c.Logger.Debugln("接收到未知帧")
	}
	return nil
}

// sendUFrame 发送U帧
func (c *Client) sendUFrame(cmd [4]byte) {
	data := convertBytes(convert4BytesToSlice(cmd))
	c.Logger.Debugf("发送U帧: [% X]", data)
	c.sendChan <- data
}

// sendSFrame 发送S帧
func (c *Client) sendSFrame() {
	rsnBytes := parseLittleEndianUInt16(uint16(c.rsn << 1))
	sendBytes := make([]byte, 0, 0)
	sendBytes = append(sendBytes, 0x01, 0x00)
	sendBytes = append(sendBytes, rsnBytes...)
	data := convertBytes(sendBytes)
	c.Logger.Debugf("发送S帧: [% X]", data)
	c.sendChan <- data
}

// SendTotalCall 发送总召唤
func (c *Client) SendTotalCall() {
	ssnBytes := parseLittleEndianUInt16(uint16(c.ssn << 1))
	rsnBytes := parseLittleEndianUInt16(uint16(c.rsn << 1))
	totalCallData := make([]byte, 0, 0)
	totalCallData = append(totalCallData, ssnBytes...)
	totalCallData = append(totalCallData, rsnBytes...)
	totalCallData = append(totalCallData, 0x64, 0x01, 0x06, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x14)
	data := convertBytes(totalCallData)
	c.Logger.Debugf("发送总召唤: [% X]", data)
	c.sendChan <- data
}

// SendSingleCmd 发送单点遥控
func (c *Client) SendSingleCmd(address uint, val uint) (err error) {
	ssnBytes := parseLittleEndianUInt16(uint16(c.ssn << 1))
	rsnBytes := parseLittleEndianUInt16(uint16(c.rsn << 1))
	singleData := make([]byte, 0, 0)
	singleData = append(singleData, ssnBytes...)
	singleData = append(singleData, rsnBytes...)
	singleData = append(singleData, 0x2D, 0x01, 0x06, 0x00, 0x01, 0x00)
	singleData = append(singleData, byte(address), byte(address>>8), byte(address>>16))
	if val == 0 {
		singleData = append(singleData, 0x01)
	} else {
		singleData = append(singleData, 0x02)
	}
	data := convertBytes(singleData)
	c.Logger.Debugf("发送发送单点遥控: [% X]", data)
	c.sendChan <- data
	return nil
}

// SendDoubleCmd 发送双点遥控
func (c *Client) SendDoubleCmd(address uint, val uint) (err error) {
	ssnBytes := parseLittleEndianUInt16(uint16(c.ssn << 1))
	rsnBytes := parseLittleEndianUInt16(uint16(c.rsn << 1))
	doubleData := make([]byte, 0, 0)
	doubleData = append(doubleData, ssnBytes...)
	doubleData = append(doubleData, rsnBytes...)
	doubleData = append(doubleData, 0x2E, 0x01, 0x06, 0x00, 0x01, 0x00)
	doubleData = append(doubleData, byte(address), byte(address>>8), byte(address>>16))
	if val == 0 {
		doubleData = append(doubleData, 0x01)
	} else {
		doubleData = append(doubleData, 0x02)
	}
	data := convertBytes(doubleData)
	c.Logger.Debugf("发送发送双点遥控: [% X]", data)
	c.sendChan <- data
	return nil
}

// SendTotalCall 发送电度总召唤
func (c *Client) SendElectricityTotalCall() {
	ssnBytes := parseLittleEndianUInt16(uint16(c.ssn << 1))
	rsnBytes := parseLittleEndianUInt16(uint16(c.rsn << 1))
	totalCallData := make([]byte, 0, 0)
	totalCallData = append(totalCallData, ssnBytes...)
	totalCallData = append(totalCallData, rsnBytes...)
	totalCallData = append(totalCallData, 0x65, 0x01, 0x06, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x05)
	data := convertBytes(totalCallData)
	c.Logger.Debugf("发送电度总召唤: [% X]", data)
	c.sendChan <- data
}

// incrRsn 增加rsn
func (c *Client) incrRsn() {
	c.rsn++
	if c.rsn < 0 {
		c.rsn = 0
	}
}

//func (c *Client) handleSignal() {
//	signals := make(chan os.Signal, 1)
//	signal.Notify(signals, os.Kill, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
//	<-signals
//	c.cancel()
//	//c.wg.Wait()
//	if c.conn != nil {
//		c.conn.Close()
//	}
//	c.Logger.Println("断开服务器连接，程序关闭")
//	os.Exit(0)
//}

// Close 结束程序
func (c *Client) Close() error {
	c.IsConnected = false
	c.cancel()
	c.conn.Close()
	return nil
}
