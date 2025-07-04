package device

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/adrianmo/go-nmea"
	aislib "github.com/hootrhino/go-ais"

	serial "github.com/hootrhino/goserial"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/resconfig"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

var __AisCodec = aislib.CodecNew(false, false, false)

// --------------------------------------------------------------------------------------------------
// 把AIS包里面的几个结构体拿出来了，主要是适配JSON格式, 下面这些结构体和AIS包里面的完全一样
// --------------------------------------------------------------------------------------------------
type _AISCommonConfig struct {
	Mode     string `json:"mode" validate:"required"`
	ParseAis *bool  `json:"parseAis" validate:"required"`
	GwSN     string `json:"gwsn" validate:"required"`
}
type _AISDeviceMasterConfig struct {
	CommonConfig _AISCommonConfig     `json:"commonConfig" validate:"required"`
	HostConfig   resconfig.HostConfig `json:"hostConfig"`
	UartConfig   resconfig.UartConfig `json:"uartConfig"`
}
type AISDeviceMaster struct {
	typex.XStatus
	status      typex.SourceState
	mainConfig  _AISDeviceMasterConfig
	RuleEngine  typex.Rhilex
	tcpListener net.Listener // TCP 接收端
	serialPort  serial.Port
	// session
	DevicesSessionMap map[string]*__AISDeviceSession
}

/*
*
* AIS 数据解析服务器
*
 */
func NewAISDeviceMaster(e typex.Rhilex) typex.XDevice {
	aism := new(AISDeviceMaster)
	aism.RuleEngine = e
	aism.mainConfig = _AISDeviceMasterConfig{
		HostConfig: resconfig.HostConfig{
			Host:    "127.0.0.1",
			Port:    6005,
			Timeout: 3000,
		},
		CommonConfig: _AISCommonConfig{
			Mode:     "TCP",
			ParseAis: new(bool),
			GwSN:     "HR0001",
		},
	}
	aism.DevicesSessionMap = map[string]*__AISDeviceSession{}
	return aism
}

//  初始化
func (aism *AISDeviceMaster) Init(devId string, configMap map[string]any) error {
	aism.PointId = devId
	if err := utils.BindSourceConfig(configMap, &aism.mainConfig); err != nil {
		return err
	}
	if !utils.SContains([]string{"UART", "TCP"}, aism.mainConfig.CommonConfig.Mode) {
		return errors.New("unsupported mode, only can be one of 'TCP' or 'RTU'")
	}
	if err := aism.mainConfig.UartConfig.Validate(); err != nil {
		return nil
	}
	return nil
}

// 启动
func (aism *AISDeviceMaster) Start(cctx typex.CCTX) error {
	aism.Ctx = cctx.Ctx
	aism.CancelCTX = cctx.CancelCTX
	if aism.mainConfig.CommonConfig.Mode == "TCP" {
		//
		listener, err := net.Listen("tcp",
			fmt.Sprintf("%s:%v", aism.mainConfig.HostConfig.Host, aism.mainConfig.HostConfig.Port))
		if err != nil {
			return err
		}
		aism.tcpListener = listener
		glogger.GLogger.Infof("AIS TCP server started on TCP://%s:%v",
			aism.mainConfig.HostConfig.Host, aism.mainConfig.HostConfig.Port)
		go aism.handleTcpConnect(listener)
		aism.status = typex.SOURCE_UP
		return nil
	}
	// 串口收发卡
	if aism.mainConfig.CommonConfig.Mode == "UART" {
		config := serial.Config{
			Address:  aism.mainConfig.UartConfig.Uart,
			BaudRate: aism.mainConfig.UartConfig.BaudRate,
			DataBits: aism.mainConfig.UartConfig.DataBits,
			Parity:   aism.mainConfig.UartConfig.Parity,
			StopBits: aism.mainConfig.UartConfig.StopBits,
			Timeout:  time.Duration(aism.mainConfig.UartConfig.Timeout) * time.Millisecond,
		}
		var err error
		aism.serialPort, err = serial.Open(&config)
		if err != nil {
			glogger.GLogger.Error("serial port start failed err:", err, ", config:", config)
			return err
		}
		go func() {
			buffer := [4096]byte{}
			defer aism.serialPort.Close()
			for {
				select {
				case <-aism.Ctx.Done():
					{
						return
					}
				default:
					{
					}
				}
				offset := 0
				endl1 := false
				endl2 := false
				ok := false
				oneByte := [1]byte{}
				readyStatus := false // 超时也是就绪状态
				ctx1, cancel1 := context.WithTimeout(aism.Ctx,
					time.Duration(aism.mainConfig.UartConfig.Timeout)*time.Millisecond)
				defer cancel1()
				for {
					select {
					// 控制时间防止死机
					case <-ctx1.Done():
						{
							if !readyStatus {
								glogger.GLogger.Warnf("serialPort %s Read timeout", aism.mainConfig.UartConfig.Uart)
							}
							break
						}
					default:
						{
						}
					}
					_, err := aism.serialPort.Read(oneByte[:])
					if err != nil {
						if strings.Contains(err.Error(), "timeout") {
							readyStatus = true
							continue
						}
						readyStatus = false
						aism.status = typex.SOURCE_DOWN
						glogger.GLogger.Errorf("serialPort %s Read error", aism.mainConfig.UartConfig.Uart)
						return
					}
					if oneByte[0] == '\r' {
						endl1 = true
						continue
					}
					if oneByte[0] == '\n' {
						endl2 = true
						ok = true
					}
					if endl1 && endl2 {
						break
					} else {
						buffer[offset] = oneByte[0]
						offset++
					}
				}
				// 可能AIS报文传输失败了
				if !ok {
					glogger.GLogger.Info("serialPort Read may occurred error:", err)
					continue
				}
				rawAiSString := string(buffer[:offset])
				if err != nil {
					glogger.GLogger.Error(err)
					aism.status = typex.SOURCE_DOWN
					return
				}
				// 可能会收到心跳包: !HRT710,Q,003,0*06
				if strings.HasPrefix(rawAiSString, "!HRT") {
					continue
				}
				if strings.HasPrefix(rawAiSString, "!DYA") {
					continue
				}
				// 这段是个兼容代码，现阶段适配了一款AIS USB 串口接收器，以后会自己做
				{
					if strings.HasPrefix("NONE", rawAiSString) {
						glogger.GLogger.Info("AIS33VRx Receiver Heart Beat Packet")
						continue
					}
					if strings.HasPrefix("AIS33VRx", rawAiSString) {
						glogger.GLogger.Info(rawAiSString)
						continue
					}
					if strings.HasPrefix("AIS Ch 1", rawAiSString) {
						glogger.GLogger.Info(rawAiSString)
						continue
					}
					if strings.HasPrefix("AIS Ch 2", rawAiSString) {
						glogger.GLogger.Info(rawAiSString)
						continue
					}
				}

				// 如果不需要解析,直接原文透传
				if !*aism.mainConfig.CommonConfig.ParseAis {
					// {
					//
					//     "gwsn":"%s"
					//     "ais_data":"%s"
					// }
					ds := `{"gwsn":"%s","ais_data":"%s"}`
					lens := len(rawAiSString)
					if lens > 2 {
						aism.RuleEngine.WorkDevice(aism.Details(),
							fmt.Sprintf(ds, aism.mainConfig.CommonConfig.GwSN, rawAiSString))
					}
				} else {
					errParseAisToJson := aism.ParseAisToJson(rawAiSString)
					if errParseAisToJson != nil {
						glogger.GLogger.Error("ParseAisToJson error:", errParseAisToJson)
						continue
					}

				}
			}
		}()
		aism.status = typex.SOURCE_UP
		return nil
	}
	aism.status = typex.SOURCE_DOWN
	return fmt.Errorf("Invalid work mode:%s", aism.mainConfig.CommonConfig.Mode)
}

// 设备当前状态
func (aism *AISDeviceMaster) Status() typex.SourceState {
	return aism.status
}

// 停止设备
func (aism *AISDeviceMaster) Stop() {
	aism.status = typex.SOURCE_DOWN
	if aism.CancelCTX != nil {
		aism.CancelCTX()
	}
	if aism.tcpListener != nil {
		aism.tcpListener.Close()
	}
	// release serial port
	if aism.mainConfig.CommonConfig.Mode == "UART" {
		if aism.serialPort != nil {
			aism.serialPort.Close()
		}
	}
}

// 真实设备
func (aism *AISDeviceMaster) Details() *typex.Device {
	return aism.RuleEngine.GetDevice(aism.PointId)
}

// 状态
func (aism *AISDeviceMaster) SetState(status typex.SourceState) {
	aism.status = status

}

func (aism *AISDeviceMaster) OnDCACall(UUID string, Command string, Args any) typex.DCAResult {
	return typex.DCAResult{}
}

/*
*
* OnCtrl 接口可以用来向外广播数据
*
 */
func (aism *AISDeviceMaster) OnCtrl(cmd []byte, args []byte) ([]byte, error) {
	return []byte{}, nil
}

//--------------------------------------------------------------------------------------------------
// 内部
//--------------------------------------------------------------------------------------------------
/*
*
* 处理连接
*
 */
func (aism *AISDeviceMaster) handleTcpConnect(listener net.Listener) {
	for {
		select {
		case <-aism.Ctx.Done():
			{
				return
			}
		default:
			{
			}
		}
		tcpcon, err := listener.Accept()
		if err != nil {
			glogger.GLogger.Error(err)
			continue
		}
		ctx, cancel := context.WithCancel(aism.Ctx)
		go aism.handleTcpAuth(ctx, cancel, &__AISDeviceSession{
			Transport: tcpcon,
		})

	}

}

/*
*
* 等待认证: 传感器发送的第一个包必须为ID, 最大不能超过64字节
* 注意：Auth只针对AIS主机，来自AIS的数据只解析不做验证
*
 */
type __AISDeviceSession struct {
	ctx       context.Context
	cancel    context.CancelFunc
	SN        string   // 注册包里的序列号, 必须是:SN-$AA-$BB-$CC-$DD
	Ip        string   // 注册包里的序列号
	Transport net.Conn // TCP连接
}

func (aism *AISDeviceMaster) handleTcpAuth(ctx context.Context,
	cancel context.CancelFunc, session *__AISDeviceSession) {
	// 5秒内读一个SN
	session.ctx = ctx
	session.cancel = cancel
	session.Transport.SetDeadline(time.Now().Add(5 * time.Second))
	reader := bufio.NewReader(session.Transport)
	registerPkt, err := reader.ReadString('$')
	session.Transport.SetDeadline(time.Time{})
	//
	if err != nil {
		glogger.GLogger.Error(session.Transport.RemoteAddr(), err)
		session.Transport.Close()
		return
	}
	// 对SN有要求, 必须不少于4个字符
	if len(registerPkt) < 4 {
		glogger.GLogger.Error("Must have register packet and can not less than 4 character")
		session.Transport.Close()
		return
	}
	sn := registerPkt[:len(registerPkt)-1] // 去除$
	glogger.GLogger.Debug("AIS Device ready to auth:", sn)
	if aism.DevicesSessionMap[sn] != nil {
		glogger.GLogger.Error("SN Already Have Been Registered:", sn)
		session.Transport.Close()
		return
	}
	session.SN = sn
	session.Ip = session.Transport.RemoteAddr().String()
	aism.DevicesSessionMap[sn] = session
	go aism.handleIO(session)
}

/*
*
* 数据处理
*
 */
func (aism *AISDeviceMaster) handleIO(session *__AISDeviceSession) {
	reader := bufio.NewReader(session.Transport)
	for {
		rawAiSString, err := reader.ReadString('\n')
		if err != nil {
			glogger.GLogger.Error(err)
			delete(aism.DevicesSessionMap, session.SN)
			session.Transport.Close()
			aism.status = typex.SOURCE_DOWN
			return
		}
		// 可能会收到心跳包: !HRT710,Q,003,0*06
		if strings.HasPrefix(rawAiSString, "!HRT") {
			glogger.GLogger.Debug("Heart beat from:", session.SN, session.Transport.RemoteAddr())
			continue
		}
		if strings.HasPrefix(rawAiSString, "!DYA") {
			glogger.GLogger.Debug("DYA Message from:", session.SN, session.Transport.RemoteAddr())
			continue
		}
		// 如果不需要解析,直接原文透传
		if !*aism.mainConfig.CommonConfig.ParseAis {
			// {
			//     "ais_data":"%s"
			//     "gwsn":"%s"
			// }
			ds := `{"gwsn":"%s","ais_data":"%s"}`
			Size := len(rawAiSString)
			if Size > 2 {
				aism.RuleEngine.WorkDevice(aism.Details(),
					fmt.Sprintf(ds, aism.mainConfig.CommonConfig.GwSN, rawAiSString[:Size-2]))
			}
			continue
		} else {
			errParseAisToJson := aism.ParseAisToJson(rawAiSString)
			if errParseAisToJson != nil {
				glogger.GLogger.Error("ParseAisToJson error:", errParseAisToJson)
				continue
			}
		}

	}

}

/*
*
* 将AIS解析成JSON
CREATE STABLE IF NOT EXISTS ais_transmitter (

	`ts` TIMESTAMP,
	`mmsi` BINARY(20),
	`name` BINARY(20),
	`call_num` BINARY(20),
	`length` FLOAT,
	`width` FLOAT,
	`draft` FLOAT,
	`main_angle` FLOAT,
	`trace_angle` FLOAT,
	`latitude` FLOAT,
	`longitude` FLOAT,
	`speed` FLOAT

) TAGS (

	ais_transmitter_id BINARY(64),
	ais_transmitter_area BINARY(64),
	ais_transmitter_bznz_master BINARY(64)

);
*/

func (aism *AISDeviceMaster) ParseAisToJson(rawAiSString string) error {
	if rawAiSString == "" {
		return fmt.Errorf("empty ais string")
	}
	sentence, err := nmea.Parse(rawAiSString)
	if err != nil {
		return err
	}
	DataType := sentence.DataType()
	if DataType == nmea.TypeRMC {
		rmc1 := sentence.(nmea.RMC)
		rmc := RMC{
			GwID:      aism.mainConfig.CommonConfig.GwSN,
			Type:      rmc1.Type,
			Validity:  rmc1.Validity,
			Latitude:  rmc1.Latitude,
			Longitude: rmc1.Longitude,
			Speed:     rmc1.Speed,
			Course:    rmc1.Course,
			Variation: rmc1.Variation,
			FFAMode:   rmc1.FFAMode,
			NavStatus: rmc1.NavStatus,
		}
		Date := fmt.Sprintf("%d-%02d-%02d", rmc1.Date.DD, rmc1.Date.MM, rmc1.Date.YY)
		seconds := float64(rmc1.Time.Second) + float64(rmc1.Time.Millisecond)/1000
		Time := fmt.Sprintf("%02d:%02d:%07.4f", rmc1.Time.Hour, rmc1.Time.Minute, seconds)
		rmc.DateTime = fmt.Sprintf("%s %s", Date, Time)
		if bytes, err := json.Marshal(rmc); err != nil {
			return err
		} else {
			aism.RuleEngine.WorkDevice(aism.Details(), string(bytes))
		}
		return nil
	}
	// GNS 是GPS定位信息
	if DataType == nmea.TypeGNS {
		gns1 := sentence.(nmea.GNS)
		gns := GNS{
			Type:       gns1.Type,
			GwID:       aism.mainConfig.CommonConfig.GwSN,
			Latitude:   gns1.Latitude,
			Longitude:  gns1.Longitude,
			Mode:       gns1.Mode,
			SVs:        gns1.SVs,
			HDOP:       gns1.HDOP,
			Altitude:   gns1.Altitude,
			Separation: gns1.Separation,
			Age:        gns1.Age,
			Station:    gns1.Station,
			NavStatus:  gns1.NavStatus,
		}
		if bytes, err := json.Marshal(gns); err != nil {
			return err
		} else {
			aism.RuleEngine.WorkDevice(aism.Details(), string(bytes))
		}
		return nil

	}
	// VDM 是AIS报文
	if DataType == nmea.TypeVDM {
		vdmo1 := sentence.(nmea.VDMVDO)
		ParsedAis := Parse_AIVDM_VDO_PayloadInfo(aism.mainConfig.CommonConfig.GwSN,
			vdmo1.DataType(), vdmo1.Payload)
		if ParsedAis != "" {
			aism.RuleEngine.WorkDevice(aism.Details(), ParsedAis)
		}
		return nil

	}
	if DataType == nmea.TypeVDO {
		vdmo1 := sentence.(nmea.VDMVDO)
		ParsedAis := Parse_AIVDM_VDO_PayloadInfo(aism.mainConfig.CommonConfig.GwSN,
			vdmo1.DataType(), vdmo1.Payload)
		if ParsedAis != "" {
			aism.RuleEngine.WorkDevice(aism.Details(), ParsedAis)
		}
		return nil
	}
	return fmt.Errorf("unsupported AIS Message Type:%s", DataType)
}

//--------------------------------------------------------------------------------------------------
// AIS 结构, 下面这些结构是从nema包里面拿过来的，删除了一些无用字段，主要为了方便JSON编码操作
//--------------------------------------------------------------------------------------------------

type RMC struct {
	// Talker    string  `json:"talker"`     // The talker id (e.g GP)
	Type      string  `json:"type"` // The data type (e.g GSA)
	GwID      string  `json:"gwid"`
	Validity  string  `json:"validity"`   // validity - A-ok, V-invalid
	Latitude  float64 `json:"latitude"`   // Latitude
	Longitude float64 `json:"longitude"`  // Longitude
	Speed     float64 `json:"speed"`      // Speed in knots
	Course    float64 `json:"course"`     // True course
	DateTime  string  `json:"date"`       // Date
	Variation float64 `json:"variation"`  // Magnetic variation
	FFAMode   string  `json:"ffa_mode"`   // FAA mode indicator (filled in NMEA 2.3 and later)
	NavStatus string  `json:"nav_status"` // Nav Status (NMEA 4.1 and later)
}

func (s RMC) String() string {
	bytes, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(bytes)
}

/*
*
* AIS消息结构体
*
 */
type VDMVDO struct {
	MessageID      int64  `json:"message_id"`
	GwID           string `json:"gwid"`
	Type           string `json:"type"` // The data type (e.g GSA)
	NumFragments   int64  `json:"numFragments"`
	FragmentNumber int64  `json:"fragmentNumber"`
	Channel        string `json:"channel"`
	Payload        []byte `json:"-"`
}
type __PositionReport struct {
	GwID             string  `json:"gwid"`
	Type             string  `json:"type"` // The data type (e.g GSA)
	MessageID        uint8   `json:"message_id"`
	UserID           uint32  `json:"user_id"`
	Valid            bool    `json:"valid"`
	Spare1           uint8   `json:"spare_1"`
	PositionAccuracy bool    `json:"position_accuracy"`
	Longitude        float64 `json:"longitude"`
	Latitude         float64 `json:"latitude"`
	Cog              float64 `json:"cog"`
	TrueHeading      uint16  `json:"true_heading"`
	Timestamp        uint8   `json:"timestamp"`
}

// AIS 规范的一个扩展报文
type __AIVDM_ExtendedClassBPositionReport struct {
	Type        string  `json:"type"` // The data type (e.g GSA)
	GwID        string  `json:"gwid"`
	MessageID   uint8   `json:"message_id"`
	UserID      uint32  `json:"user_id"`
	Name        string  `json:"name"`
	Sog         float64 `json:"sog"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	Cog         float64 `json:"cog"`
	TrueHeading uint16  `json:"true_heading"`
	Timestamp   uint8   `json:"timestamp"`
}
type __AIVDM_StaticDataReport struct {
	Type           string `json:"type"` // The data type (e.g GSA)
	GwID           string `json:"gwid"`
	MessageID      uint8  `json:"message_id"`
	UserID         uint32 `json:"user_id"`
	PartNumber     bool   `json:"part_number"`
	Valid          bool   `json:"valid"`
	ShipType       uint8  `json:"ship_type"`
	VendorIDName   string `json:"vendor_id_name"`
	VenderIDModel  uint8  `json:"vender_id_model"`
	VenderIDSerial uint32 `json:"vender_id_serial"`
	CallSign       string `json:"call_sign"`
}

func Parse_AIVDM_VDO_PayloadInfo(GwID, Type string, Payload []byte) string {
	__AisCodec.DropSpace = true
	pkt := __AisCodec.DecodePacket(Payload)
	var _Type reflect.Type
	if _Type = reflect.TypeOf(pkt); _Type == nil {
		return ""
	}
	// 上报位置
	TypeName := _Type.Name()
	if TypeName == "StandardClassBPositionReport" {
		spr := pkt.(aislib.StandardClassBPositionReport)
		PositionReport := __PositionReport{
			Type:             Type,
			GwID:             GwID,
			MessageID:        spr.MessageID,
			UserID:           spr.UserID,
			Valid:            spr.Valid,
			Spare1:           spr.Spare1,
			PositionAccuracy: spr.PositionAccuracy,
			Longitude:        float64(spr.Latitude),
			Latitude:         float64(spr.Latitude),
			Cog:              float64(spr.Cog),
			TrueHeading:      spr.TrueHeading,
			Timestamp:        spr.Timestamp,
		}
		bytes, _ := json.Marshal(PositionReport)
		return string(bytes)
	}
	// "StaticDataReport"
	if TypeName == "StaticDataReport" {
		spr := pkt.(aislib.StaticDataReport)
		data := __AIVDM_StaticDataReport{
			Type:           Type,
			GwID:           GwID,
			UserID:         spr.UserID,
			MessageID:      spr.MessageID,
			PartNumber:     spr.PartNumber,
			Valid:          spr.Valid,
			ShipType:       spr.ReportB.ShipType,
			VendorIDName:   spr.ReportB.VendorIDName,
			VenderIDModel:  spr.ReportB.VenderIDModel,
			VenderIDSerial: spr.ReportB.VenderIDSerial,
			CallSign:       spr.ReportB.CallSign,
		}
		bytes, _ := json.Marshal(data)
		return string(bytes)
	}
	if TypeName == "ExtendedClassBPositionReport" {
		spr := pkt.(aislib.ExtendedClassBPositionReport)
		data := __AIVDM_ExtendedClassBPositionReport{
			Type:        Type,
			UserID:      spr.UserID,
			GwID:        GwID,
			MessageID:   spr.MessageID,
			Sog:         float64(spr.Sog),
			Longitude:   float64(spr.Longitude),
			Latitude:    float64(spr.Latitude),
			Cog:         float64(spr.Cog),
			TrueHeading: spr.TrueHeading,
			Timestamp:   spr.Timestamp,
			Name:        spr.Name,
		}
		bytes, _ := json.Marshal(data)
		return string(bytes)
	}
	return ""
}

type Time struct {
	Valid       bool `json:"valid"`
	Hour        int  `json:"hour"`
	Minute      int  `json:"minute"`
	Second      int  `json:"second"`
	Millisecond int  `json:"millisecond"`
}

// String representation of Time
func (t Time) String() string {
	seconds := float64(t.Second) + float64(t.Millisecond)/1000
	return fmt.Sprintf("%02d:%02d:%07.4f", t.Hour, t.Minute, seconds)
}

// Date type
type Date struct {
	Valid bool `json:"valid"`
	DD    int  `json:"dd"`
	MM    int  `json:"mm"`
	YY    int  `json:"yy"`
}

// String representation of date
func (d Date) String() string {
	return fmt.Sprintf("%02d-%02d-%02d", d.DD, d.MM, d.YY)
}

/*
*
*经纬度
*
 */
type GNS struct {
	Type       string   `json:"type"`
	GwID       string   `json:"gwid"`
	Latitude   float64  `json:"latitude"`
	Longitude  float64  `json:"longitude"`
	Mode       []string `json:"mode"`
	SVs        int64    `json:"s_vs"`
	HDOP       float64  `json:"hdop"`
	Altitude   float64  `json:"altitude"`
	Separation float64  `json:"separation"`
	Age        float64  `json:"age"`
	Station    int64    `json:"station"`
	NavStatus  string   `json:"nav_status"`
}
