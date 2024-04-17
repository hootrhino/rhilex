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
			"ToHttp":     rhilexlib.DataToHttp(e),
			"ToMqtt":     rhilexlib.DataToMqtt(e),
			"ToUdp":      rhilexlib.DataToUdp(e),
			"ToTcp":      rhilexlib.DataToTcp(e),
			"ToTdEngine": rhilexlib.DataToTdEngine(e),
			"ToMongo":    rhilexlib.DataToMongo(e),
		}
		AddAppLibToGroup(app, e, "data", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Debug": rhilexlib.DebugAPP(e, app.UUID),
			"Throw": rhilexlib.Throw(e),
		}
		AddAppLibToGroup(app, e, "_G", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"VSet":             rhilexlib.StoreSet(e),
			"VSetWithDuration": rhilexlib.StoreSetWithDuration(e),
			"VGet":             rhilexlib.StoreGet(e),
			"VDel":             rhilexlib.StoreDelete(e),
		}
		AddAppLibToGroup(app, e, "kv", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Time":       rhilexlib.Time(e),
			"TimeMs":     rhilexlib.TimeMs(e),
			"TsUnix":     rhilexlib.TsUnix(e),
			"TsUnixNano": rhilexlib.TsUnixNano(e),
			"NtpTime":    rhilexlib.NtpTime(e),
			"Sleep":      rhilexlib.Sleep(e),
		}
		AddAppLibToGroup(app, e, "time", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"HToN":       rhilexlib.HToN(e),
			"HsubToN":    rhilexlib.HsubToN(e),
			"MatchHex":   rhilexlib.MatchHex(e),
			"MatchUInt":  rhilexlib.MatchUInt(e),
			"Bytes2Hexs": rhilexlib.Bytes2Hexs(e),
			"Hexs2Bytes": rhilexlib.Hexs2Bytes(e),
			"ABCD":       rhilexlib.ABCD(e),
			"DCBA":       rhilexlib.DCBA(e),
			"BADC":       rhilexlib.BADC(e),
			"CDAB":       rhilexlib.CDAB(e),
		}
		AddAppLibToGroup(app, e, "hex", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"MB":            rhilexlib.MatchBinary(e),
			"MBHex":         rhilexlib.MatchBinaryHex(e),
			"B2BS":          rhilexlib.ByteToBitString(e),
			"Bit":           rhilexlib.GetABitOnByte(e),
			"B2I64":         rhilexlib.ByteToInt64(e),
			"B64S2B":        rhilexlib.B64S2B(e),
			"BS2B":          rhilexlib.BitStringToBytes(e),
			"Bin2F32":       rhilexlib.BinToFloat32(e),
			"Bin2F64":       rhilexlib.BinToFloat64(e),
			"Bin2F32Big":    rhilexlib.BinToFloat32(e),
			"Bin2F64Big":    rhilexlib.BinToFloat64(e),
			"Bin2F32Little": rhilexlib.BinToFloat32Little(e),
			"Bin2F64Little": rhilexlib.BinToFloat64Little(e),
		}
		AddAppLibToGroup(app, e, "binary", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"T2J": rhilexlib.JSONE(e),
			"J2T": rhilexlib.JSOND(e),
		}
		AddAppLibToGroup(app, e, "json", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"ReadDevice":  rhilexlib.ReadDevice(e),
			"WriteDevice": rhilexlib.WriteDevice(e),
			"CtrlDevice":  rhilexlib.CtrlDevice(e),
			"ReadSource":  rhilexlib.ReadSource(e),
			"WriteSource": rhilexlib.WriteSource(e),
		}
		AddAppLibToGroup(app, e, "device", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"T2Str":   rhilexlib.T2Str(e),
			"Bin2Str": rhilexlib.Bin2Str(e),
		}
		AddAppLibToGroup(app, e, "string", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"F5":  rhilexlib.F5(e),
			"F6":  rhilexlib.F6(e),
			"F15": rhilexlib.F15(e),
			"F16": rhilexlib.F16(e),
		}
		AddAppLibToGroup(app, e, "modbus", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"DO1Set":  rhilexlib.H3DO1Set(e),
			"DO1Get":  rhilexlib.H3DO1Get(e),
			"DO2Set":  rhilexlib.H3DO2Set(e),
			"DO2Get":  rhilexlib.H3DO2Get(e),
			"DI1Get":  rhilexlib.H3DI1Get(e),
			"DI2Get":  rhilexlib.H3DI2Get(e),
			"DI3Get":  rhilexlib.H3DI3Get(e),
			"Led1On":  rhilexlib.Led1On(e),
			"Led1Off": rhilexlib.Led1Off(e),
		}
		AddAppLibToGroup(app, e, "rhinopi", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"XOR":   rhilexlib.XOR(e),
			"CRC16": rhilexlib.CRC16(e),
		}
		AddAppLibToGroup(app, e, "misc", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"GPIOGet": rhilexlib.RASPI4_GPIOGet(e),
			"GPIOSet": rhilexlib.RASPI4_GPIOSet(e),
		}
		AddAppLibToGroup(app, e, "raspi4b", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"GPIOGet": rhilexlib.WKYWS1608_GPIOGet(e),
			"GPIOSet": rhilexlib.WKYWS1608_GPIOSet(e),
		}
		AddAppLibToGroup(app, e, "ws1608", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"TFloat": rhilexlib.TruncateFloat(e),
		}
		AddAppLibToGroup(app, e, "math", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"PlayMusic": rhilexlib.PlayMusic(e),
		}
		AddAppLibToGroup(app, e, "audio", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Request": rhilexlib.Request(e),
		}
		AddAppLibToGroup(app, e, "rpc", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Execute": rhilexlib.JqSelect(e),
		}
		AddAppLibToGroup(app, e, "jq", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Ping": rhilexlib.PingIp(e),
		}
		AddAppLibToGroup(app, e, "network", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Get":  rhilexlib.HttpGet(e),
			"Post": rhilexlib.HttpPost(e),
		}
		AddAppLibToGroup(app, e, "http", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"LedOn":  rhilexlib.EN6400_LedOn(e),
			"LedOff": rhilexlib.EN6400_LedOff(e),
		}
		AddAppLibToGroup(app, e, "en6400", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Update": rhilexlib.DataSchemaValueUpdate(e),
		}
		AddAppLibToGroup(app, e, "dataschema", Funcs)
	}
}
