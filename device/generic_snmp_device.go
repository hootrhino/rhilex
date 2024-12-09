package device

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/hootrhino/rhilex/component/alarmcenter"
	"github.com/hootrhino/rhilex/component/intercache"
	"github.com/hootrhino/rhilex/component/interdb"
	"github.com/hootrhino/rhilex/resconfig"

	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

// snmpOids: SNMP点位表
type snmpOid struct {
	UUID      string `json:"-"`
	Oid       string `json:"oid"`   // .1.3.6.1.2.1.25.1.6.0
	Tag       string `json:"tag"`   // temp
	Alias     string `json:"alias"` // 温度
	Frequency *int   `json:"-"`     // 请求频率
}
type _SNMPCommonConfig struct {
	AutoRequest  *bool `json:"autoRequest" validate:"required"`  // 自动请求
	BatchRequest *bool `json:"batchRequest" validate:"required"` // 批量采集
	EnableGroup  *bool `json:"enableGroup" validate:"required"`  // 并发请求, 注意: 这个开关可能会把目标设备搞挂了, 根据设备性能量力而行
	Timeout      *int  `json:"timeout" validate:"required"`      // 请求超时
	Frequency    *int  `json:"frequency" validate:"required"`    // 请求频率
}

type _GSNMPConfig struct {
	SchemaId      string                      `json:"schemaId"`
	CommonConfig  _SNMPCommonConfig           `json:"commonConfig" validate:"required"`
	SNMPConfig    resconfig.GenericSnmpConfig `json:"snmpConfig" validate:"required"`
	CecollaConfig resconfig.CecollaConfig     `json:"cecollaConfig"`
	AlarmConfig   resconfig.AlarmConfig       `json:"alarmConfig"`
}

type genericSnmpDevice struct {
	typex.XStatus
	status     typex.DeviceState
	RuleEngine typex.Rhilex
	locker     sync.Locker
	mainConfig _GSNMPConfig
	snmpOids   map[string]snmpOid
	client     *gosnmp.GoSNMP
}

// Example: 0x02 0x92 0xFF 0x98
/*
*
* 温湿度传感器
*
 */
func NewGenericSnmpDevice(e typex.Rhilex) typex.XDevice {
	sd := new(genericSnmpDevice)
	sd.RuleEngine = e
	sd.locker = &sync.Mutex{}
	sd.mainConfig = _GSNMPConfig{
		CommonConfig: _SNMPCommonConfig{
			EnableGroup: func() *bool {
				b := false
				return &b
			}(),
			AutoRequest: func() *bool {
				b := true
				return &b
			}(),
			Timeout: func() *int {
				a := 5000
				return &a
			}(), // ms
			Frequency: func() *int {
				a := 5000
				return &a
			}(), // ms
			BatchRequest: func() *bool {
				b := false
				return &b
			}(),
		},
		CecollaConfig: resconfig.CecollaConfig{
			Enable: func() *bool {
				b := false
				return &b
			}(),
			EnableCreateSchema: func() *bool {
				b := true
				return &b
			}(),
		},
		AlarmConfig: resconfig.AlarmConfig{
			Enable: func() *bool {
				b := false
				return &b
			}(),
		},
	}
	sd.snmpOids = map[string]snmpOid{}
	return sd
}

// 数据模型
type SnmpSchemaProperty struct {
	UUID          string
	Status        int    // 0 正常；1 错误，填充 ErrMsg
	ErrMsg        string // 错误信息
	LastFetchTime uint64 // 最后更新时间
	Name          string // 变量关联名
	Value         any    // 运行时值
}

//  初始化
func (sd *genericSnmpDevice) Init(devId string, configMap map[string]interface{}) error {
	sd.PointId = devId
	intercache.RegisterSlot(devId)
	if err := utils.BindSourceConfig(configMap, &sd.mainConfig); err != nil {
		return err
	}
	snmpOids := []snmpOid{}
	snmpOidLoadErr := interdb.InterDb().Table("m_snmp_oids").
		Where("device_uuid=?", devId).Find(&snmpOids).Error
	if snmpOidLoadErr != nil {
		return snmpOidLoadErr
	}
	for _, oid := range snmpOids {
		sd.snmpOids[oid.UUID] = snmpOid{
			UUID:      oid.UUID,
			Oid:       oid.Oid,
			Tag:       oid.Tag,
			Alias:     oid.Alias,
			Frequency: oid.Frequency,
		}
		LastFetchTime := uint64(time.Now().UnixMilli())
		intercache.SetValue(sd.PointId, oid.UUID, intercache.CacheValue{
			UUID:          oid.UUID,
			Status:        0,
			LastFetchTime: LastFetchTime,
			Value:         "",
			ErrMsg:        "--",
		})
	}
	if sd.mainConfig.SchemaId != "" {
		var SchemaProperties []SnmpSchemaProperty
		dataSchemaLoadError := interdb.InterDb().Table("m_iot_properties").
			Where("schema_id=?", sd.mainConfig.SchemaId).Find(&SchemaProperties).Error
		if dataSchemaLoadError != nil {
			return dataSchemaLoadError
		}
		LastFetchTime := uint64(time.Now().UnixMilli())
		for _, MSnmpSchemaProperty := range SchemaProperties {
			intercache.SetValue(sd.PointId, MSnmpSchemaProperty.Name,
				intercache.CacheValue{
					UUID:          MSnmpSchemaProperty.UUID,
					LastFetchTime: LastFetchTime,
					Value:         "-",
				})
		}
	}
	return nil
}

// Version1  SnmpVersion = 0x0
// Version2c SnmpVersion = 0x1
// Version3  SnmpVersion = 0x3
// 启动
func (sd *genericSnmpDevice) Start(cctx typex.CCTX) error {
	sd.Ctx = cctx.Ctx
	sd.CancelCTX = cctx.CancelCTX
	sd.client = &gosnmp.GoSNMP{
		Target:             sd.mainConfig.SNMPConfig.Target,
		Port:               sd.mainConfig.SNMPConfig.Port,
		Transport:          "udp",
		Community:          sd.mainConfig.SNMPConfig.Community,
		Version:            0x1,
		Timeout:            time.Duration(3) * time.Second,
		Retries:            1,
		ExponentialTimeout: false,
		MaxOids:            60,
	}
	err := sd.client.Connect()
	if err != nil {
		glogger.GLogger.Errorf("Connect err: %v", err)
		return err
	}

	//---------------------------------------------------------------------------------
	// Start
	//---------------------------------------------------------------------------------
	if !*sd.mainConfig.CommonConfig.AutoRequest {
		sd.status = typex.DEV_UP
		return nil
	}
	go func(sd *genericSnmpDevice) {
		ticker := time.NewTicker(5000 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-sd.Ctx.Done():
				return
			case <-time.After(4 * time.Millisecond):
				// Continue loop
			}
			snmpOids, err := sd.readData()
			if err != nil {
				glogger.GLogger.Error(err)
				goto END
			}
			if len(snmpOids) < 1 {
				goto END
			}
			if len(snmpOids) < 1 {
				time.Sleep(50 * time.Second)
				continue
			}
			if *sd.mainConfig.CommonConfig.BatchRequest {
				if len(snmpOids) > 0 {
					if bytes, err := json.Marshal(snmpOids); err != nil {
						glogger.GLogger.Error(err)
					} else {
						glogger.GLogger.Debug(string(bytes))
						sd.RuleEngine.WorkDevice(sd.Details(), string(bytes))
					}
				}
			}
			// 是否预警
			if *sd.mainConfig.AlarmConfig.Enable {
				Input := map[string]any{}
				Input["data"] = snmpOids
				_, err := alarmcenter.Input(sd.mainConfig.AlarmConfig.AlarmRuleId, sd.PointId, Input)
				if err != nil {
					glogger.GLogger.Error(err)
				}
			}
		END:
			<-ticker.C
		}

	}(sd)
	sd.status = typex.DEV_UP
	return nil
}

// 设备当前状态
func (sd *genericSnmpDevice) Status() typex.DeviceState {
	return sd.status
}

// 停止设备
func (sd *genericSnmpDevice) Stop() {
	sd.status = typex.DEV_DOWN
	if sd.CancelCTX != nil {
		sd.CancelCTX()
	}
	intercache.UnRegisterSlot(sd.PointId)
}

// 真实设备
func (sd *genericSnmpDevice) Details() *typex.Device {
	return sd.RuleEngine.GetDevice(sd.PointId)
}

// 状态
func (sd *genericSnmpDevice) SetState(status typex.DeviceState) {
	sd.status = status

}

func (sd *genericSnmpDevice) OnDCACall(UUID string, Command string, Args interface{}) typex.DCAResult {
	return typex.DCAResult{}
}

func (sd *genericSnmpDevice) OnCtrl(cmd []byte, args []byte) ([]byte, error) {
	return []byte{}, nil
}

/*
*
* 读数据
*
 */

//	EndOfContents     Asn1BER = 0x00
//	UnknownType       Asn1BER = 0x00
//	Boolean           Asn1BER = 0x01
//	Integer           Asn1BER = 0x02
//	BitString         Asn1BER = 0x03
//	OctetString       Asn1BER = 0x04
//	Null              Asn1BER = 0x05
//	ObjectIdentifier  Asn1BER = 0x06
//	ObjectDescription Asn1BER = 0x07
//	IPAddress         Asn1BER = 0x40
//	Counter32         Asn1BER = 0x41
//	Gauge32           Asn1BER = 0x42
//	TimeTicks         Asn1BER = 0x43
//	Opaque            Asn1BER = 0x44
//	NsapAddress       Asn1BER = 0x45
//	Counter64         Asn1BER = 0x46
//	Uinteger32        Asn1BER = 0x47
//	OpaqueFloat       Asn1BER = 0x78
//	OpaqueDouble      Asn1BER = 0x79
//	NoSuchObject      Asn1BER = 0x80
//	NoSuchInstance    Asn1BER = 0x81
//	EndOfMibView      Asn1BER = 0x82
//

type ReadSnmpOidValue struct {
	Tag   string `json:"tag"`   // temp
	Alias string `json:"alias"` // 温度
	Value any    `json:"value"`
}

func (sd *genericSnmpDevice) readData() ([]ReadSnmpOidValue, error) {
	if err1 := sd.client.Connect(); err1 != nil {
		return nil, err1
	}
	result := []ReadSnmpOidValue{}

	if !*sd.mainConfig.CommonConfig.EnableGroup {
		for _, oid := range sd.snmpOids {
			R, ok := sd.walk(oid)
			if ok {
				result = append(result, R)
			}
		}
		return result, nil
	}
	wg := sync.WaitGroup{}
	wg.Add(len(sd.snmpOids))
	for _, oid := range sd.snmpOids {
		go func(snmpOid snmpOid) {
			defer wg.Done()
			Value := ""
			glogger.GLogger.Debug("SNMP Walk:", snmpOid.Oid)
			err2 := sd.client.Walk(snmpOid.Oid, func(variable gosnmp.SnmpPDU) error {
				// 目前先考虑这么多类型，其他的好像没见过
				if variable.Type == gosnmp.OctetString {
					Value = fmt.Sprintf("%v", string(variable.Value.([]byte)))
				}
				if variable.Type == gosnmp.Integer {
					Value = fmt.Sprintf("%v", int64(variable.Value.(int)))
				}
				if variable.Type == gosnmp.Boolean {
					Value = fmt.Sprintf("%v", bool(variable.Value.(bool)))
				}
				if variable.Type == gosnmp.IPAddress {
					Value = fmt.Sprintf("%v", string(variable.Value.([]byte)))
				}
				if variable.Type == gosnmp.Null {
					Value = "Null"
				}
				return nil
			})

			lastTimes := uint64(time.Now().UnixMilli())
			NewValue := intercache.CacheValue{
				UUID:          snmpOid.UUID,
				Status:        0,
				LastFetchTime: lastTimes,
				Value:         "",
				ErrMsg:        "",
			}
			if err2 != nil {
				glogger.GLogger.Error(err2)
				NewValue.ErrMsg = err2.Error()
			} else {
				NewValue.ErrMsg = ""
				NewValue.Value = Value
				NewValue.Status = 1
				result = append(result, ReadSnmpOidValue{
					Tag:   snmpOid.Tag,
					Alias: snmpOid.Alias,
					Value: Value,
				})
			}
			intercache.SetValue(sd.PointId, snmpOid.UUID, NewValue)
			if !*sd.mainConfig.CommonConfig.BatchRequest {
				if bytes, errMarshal := json.Marshal(snmpOid); errMarshal != nil {
					glogger.GLogger.Error(errMarshal)
				} else {
					sd.RuleEngine.WorkDevice(sd.Details(), string(bytes))
				}
			}
		}(oid)
	}
	wg.Wait()

	return result, nil
}
func (sd *genericSnmpDevice) walk(snmpOid snmpOid) (ReadSnmpOidValue, bool) {
	result := ReadSnmpOidValue{}
	Value := ""
	glogger.GLogger.Debug("SNMP Walk:", snmpOid.Oid)
	err2 := sd.client.Walk(snmpOid.Oid, func(variable gosnmp.SnmpPDU) error {
		// 目前先考虑这么多类型，其他的好像没见过
		if variable.Type == gosnmp.OctetString {
			Value = fmt.Sprintf("%v", string(variable.Value.([]byte)))
		}
		if variable.Type == gosnmp.Integer {
			Value = fmt.Sprintf("%v", int64(variable.Value.(int)))
		}
		if variable.Type == gosnmp.Boolean {
			Value = fmt.Sprintf("%v", bool(variable.Value.(bool)))
		}
		if variable.Type == gosnmp.IPAddress {
			Value = fmt.Sprintf("%v", string(variable.Value.([]byte)))
		}
		if variable.Type == gosnmp.Null {
			Value = "Null"
		}
		return nil
	})

	lastTimes := uint64(time.Now().UnixMilli())
	NewValue := intercache.CacheValue{
		UUID:          snmpOid.UUID,
		Status:        0,
		LastFetchTime: lastTimes,
		Value:         "",
		ErrMsg:        "",
	}
	if err2 != nil {
		glogger.GLogger.Error(err2)
		NewValue.ErrMsg = err2.Error()
	} else {
		NewValue.ErrMsg = ""
		NewValue.Value = Value
		NewValue.Status = 1
	}
	intercache.SetValue(sd.PointId, snmpOid.UUID, NewValue)
	if err2 != nil {
		return result, false
	}
	result.Tag = snmpOid.Tag
	result.Alias = snmpOid.Alias
	result.Value = Value
	return result, true
}
