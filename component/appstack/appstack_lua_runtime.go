// Copyright (C) 2023 wwhai
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
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package appstack

import (
	"fmt"

	lua "github.com/hootrhino/gopher-lua"
	"github.com/hootrhino/rhilex/rhilexlib"
	"github.com/hootrhino/rhilex/typex"
)

// 临时校验语法
func ValidateLuaSyntax(bytes []byte) error {
	// 把虚拟机参数全部设置为0是为了防止误操作产生垃圾数据
	tempVm := lua.NewState(lua.Options{
		SkipOpenLibs:     true,
		RegistrySize:     0,
		RegistryMaxSize:  0,
		RegistryGrowStep: 0,
	})
	if err := tempVm.DoString(string(bytes)); err != nil {
		return err
	}
	// 检查函数入口
	AppMain := tempVm.GetGlobal("Main")
	if AppMain == nil {
		return fmt.Errorf("'Main' field not exists")
	}
	if AppMain.Type() != lua.LTFunction {
		return fmt.Errorf("'Main' must be function(arg)")
	}
	tempVm.Close()
	tempVm = nil
	return nil
}

/*
*
  - 分组加入函数
*/
func AddAppLibToGroup(app *Application, rx typex.Rhilex,
	ModuleName string, funcs map[string]func(l *lua.LState) int) {
	var table *lua.LTable
	if ModuleName == "_G" {
		table = app.vm.G.Global
	} else {
		table = app.vm.NewTable()
	}
	app.vm.SetGlobal(ModuleName, table)
	for funcName, f := range funcs {
		table.RawSetString(funcName, app.vm.NewClosure(f))
	}
	app.vm.Push(table)
}

func LoadAppLibGroup(app *Application, e typex.Rhilex) {
	{
		Funcs := map[string]func(l *lua.LState) int{
			"ToHttp":     rhilexlib.DataToHttp(e, app.UUID),
			"ToMqtt":     rhilexlib.DataToMqtt(e, app.UUID),
			"ToUdp":      rhilexlib.DataToUdp(e, app.UUID),
			"ToTcp":      rhilexlib.DataToTcp(e, app.UUID),
			"ToTdEngine": rhilexlib.DataToTdEngine(e, app.UUID),
			"ToMongo":    rhilexlib.DataToMongo(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "data", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Debug": rhilexlib.DebugAPP(e, app.UUID),
			"Throw": rhilexlib.Throw(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "_G", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"VSet":             rhilexlib.StoreSet(e, app.UUID),
			"VSetWithDuration": rhilexlib.StoreSetWithDuration(e, app.UUID),
			"VGet":             rhilexlib.StoreGet(e, app.UUID),
			"VDel":             rhilexlib.StoreDelete(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "kv", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Time":       rhilexlib.Time(e, app.UUID),
			"TimeMs":     rhilexlib.TimeMs(e, app.UUID),
			"TsUnix":     rhilexlib.TsUnix(e, app.UUID),
			"TsUnixNano": rhilexlib.TsUnixNano(e, app.UUID),
			"NtpTime":    rhilexlib.NtpTime(e, app.UUID),
			"Sleep":      rhilexlib.Sleep(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "time", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"HToN":       rhilexlib.HToN(e, app.UUID),
			"HsubToN":    rhilexlib.HsubToN(e, app.UUID),
			"MatchHex":   rhilexlib.MatchHex(e, app.UUID),
			"MatchUInt":  rhilexlib.MatchUInt(e, app.UUID),
			"Bytes2Hexs": rhilexlib.Bytes2Hexs(e, app.UUID),
			"Hexs2Bytes": rhilexlib.Hexs2Bytes(e, app.UUID),
			"ABCD":       rhilexlib.ABCD(e, app.UUID),
			"DCBA":       rhilexlib.DCBA(e, app.UUID),
			"BADC":       rhilexlib.BADC(e, app.UUID),
			"CDAB":       rhilexlib.CDAB(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "hex", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"MB":            rhilexlib.MatchBinary(e, app.UUID),
			"MBHex":         rhilexlib.MatchBinaryHex(e, app.UUID),
			"B2BS":          rhilexlib.ByteToBitString(e, app.UUID),
			"Bit":           rhilexlib.GetABitOnByte(e, app.UUID),
			"B2I64":         rhilexlib.ByteToInt64(e, app.UUID),
			"B64S2B":        rhilexlib.B64S2B(e, app.UUID),
			"BS2B":          rhilexlib.BitStringToBytes(e, app.UUID),
			"Bin2F32":       rhilexlib.BinToFloat32(e, app.UUID),
			"Bin2F64":       rhilexlib.BinToFloat64(e, app.UUID),
			"Bin2F32Big":    rhilexlib.BinToFloat32(e, app.UUID),
			"Bin2F64Big":    rhilexlib.BinToFloat64(e, app.UUID),
			"Bin2F32Little": rhilexlib.BinToFloat32Little(e, app.UUID),
			"Bin2F64Little": rhilexlib.BinToFloat64Little(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "binary", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"T2J": rhilexlib.JSONE(e, app.UUID),
			"J2T": rhilexlib.JSOND(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "json", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"ReadDevice":  rhilexlib.ReadDevice(e, app.UUID),
			"WriteDevice": rhilexlib.WriteDevice(e, app.UUID),
			"CtrlDevice":  rhilexlib.CtrlDevice(e, app.UUID),
			"ReadSource":  rhilexlib.ReadSource(e, app.UUID),
			"WriteSource": rhilexlib.WriteSource(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "device", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"T2Str":   rhilexlib.T2Str(e, app.UUID),
			"Bin2Str": rhilexlib.Bin2Str(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "string", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"F5":  rhilexlib.F5(e, app.UUID),
			"F6":  rhilexlib.F6(e, app.UUID),
			"F15": rhilexlib.F15(e, app.UUID),
			"F16": rhilexlib.F16(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "modbus", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"DO1Set":  rhilexlib.RHILEXG1_DO1Set(e, app.UUID),
			"DO1Get":  rhilexlib.RHILEXG1_DO1Get(e, app.UUID),
			"DO2Set":  rhilexlib.RHILEXG1_DO2Set(e, app.UUID),
			"DO2Get":  rhilexlib.RHILEXG1_DO2Get(e, app.UUID),
			"DI1Get":  rhilexlib.H3DI1Get(e, app.UUID),
			"DI2Get":  rhilexlib.H3DI2Get(e, app.UUID),
			"DI3Get":  rhilexlib.H3DI3Get(e, app.UUID),
			"Led1On":  rhilexlib.Led1On(e, app.UUID),
			"Led1Off": rhilexlib.Led1Off(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "rhinopi", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"XOR":   rhilexlib.XOR(e, app.UUID),
			"CRC16": rhilexlib.CRC16(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "misc", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"GPIOGet": rhilexlib.RASPI4_GPIOGet(e, app.UUID),
			"GPIOSet": rhilexlib.RASPI4_GPIOSet(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "raspi4b", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"GPIOGet": rhilexlib.WKYWS1608_GPIOGet(e, app.UUID),
			"GPIOSet": rhilexlib.WKYWS1608_GPIOSet(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "ws1608", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"TFloat":    rhilexlib.TruncateFloat(e, app.UUID),
			"RandomInt": rhilexlib.RandomInt(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "math", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"PlayMusic": rhilexlib.PlayMusic(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "audio", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Request": rhilexlib.Request(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "rpc", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Execute": rhilexlib.JqSelect(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "jq", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Ping": rhilexlib.PingIp(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "network", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Get":  rhilexlib.HttpGet(e, app.UUID),
			"Post": rhilexlib.HttpPost(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "http", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"LedOn":  rhilexlib.EN6400_LedOn(e, app.UUID),
			"LedOff": rhilexlib.EN6400_LedOff(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "en6400", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Update": rhilexlib.DataSchemaValueUpdate(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "dataschema", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Save":      rhilexlib.InsertToDataCenterTable(e, app.UUID),
			"QueryList": rhilexlib.QueryDataCenterList(e, app.UUID),
			"QueryLast": rhilexlib.QueryDataCenterLast(e, app.UUID),
		}
		AddAppLibToGroup(app, e, "rds", Funcs)
	}
}
