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

package engine

import (
	lua "github.com/hootrhino/gopher-lua"
	"github.com/hootrhino/rhilex/core"
	"github.com/hootrhino/rhilex/rhilexlib"
	"github.com/hootrhino/rhilex/typex"
)

/*
*
  - 分组加入函数
*/
func AddRuleLibToGroup(r *typex.Rule, rx typex.Rhilex,
	ModuleName string, funcs map[string]func(l *lua.LState) int) {
	var table *lua.LTable
	if ModuleName == "_G" {
		table = r.LuaVM.G.Global
	} else {
		table = r.LuaVM.NewTable()
	}
	r.LuaVM.SetGlobal(ModuleName, table)
	for funcName, f := range funcs {
		table.RawSetString(funcName, r.LuaVM.NewClosure(f))
	}
	r.LuaVM.Push(table)
}

func LoadRuleLibGroup(r *typex.Rule, e typex.Rhilex) {
	{
		Funcs := map[string]func(l *lua.LState) int{
			"ToHttp":     rhilexlib.DataToHttp(e),
			"ToMqtt":     rhilexlib.DataToMqtt(e),
			"ToUdp":      rhilexlib.DataToUdp(e),
			"ToTcp":      rhilexlib.DataToTcp(e),
			"ToTdEngine": rhilexlib.DataToTdEngine(e),
			"ToMongo":    rhilexlib.DataToMongo(e),
		}
		AddRuleLibToGroup(r, e, "data", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Debug": rhilexlib.DebugRule(e, r.UUID),
			"Throw": rhilexlib.Throw(e),
		}
		AddRuleLibToGroup(r, e, "_G", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"VSet":             rhilexlib.StoreSet(e),
			"VSetWithDuration": rhilexlib.StoreSetWithDuration(e),
			"VGet":             rhilexlib.StoreGet(e),
			"VDel":             rhilexlib.StoreDelete(e),
		}
		AddRuleLibToGroup(r, e, "kv", Funcs)
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
		AddRuleLibToGroup(r, e, "time", Funcs)
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
		AddRuleLibToGroup(r, e, "hex", Funcs)
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
		AddRuleLibToGroup(r, e, "binary", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"T2J": rhilexlib.JSONE(e),
			"J2T": rhilexlib.JSOND(e),
		}
		AddRuleLibToGroup(r, e, "json", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"ReadDevice":  rhilexlib.ReadDevice(e),
			"WriteDevice": rhilexlib.WriteDevice(e),
			"CtrlDevice":  rhilexlib.CtrlDevice(e),
			"ReadSource":  rhilexlib.ReadSource(e),
			"WriteSource": rhilexlib.WriteSource(e),
		}
		AddRuleLibToGroup(r, e, "device", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"T2Str":   rhilexlib.T2Str(e),
			"Bin2Str": rhilexlib.Bin2Str(e),
		}
		AddRuleLibToGroup(r, e, "string", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"F5":  rhilexlib.F5(e),
			"F6":  rhilexlib.F6(e),
			"F15": rhilexlib.F15(e),
			"F16": rhilexlib.F16(e),
		}
		AddRuleLibToGroup(r, e, "modbus", Funcs)
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
		AddRuleLibToGroup(r, e, "rhinopi", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"XOR":   rhilexlib.XOR(e),
			"CRC16": rhilexlib.CRC16(e),
		}
		AddRuleLibToGroup(r, e, "misc", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"GPIOGet": rhilexlib.RASPI4_GPIOGet(e),
			"GPIOSet": rhilexlib.RASPI4_GPIOSet(e),
		}
		AddRuleLibToGroup(r, e, "raspi4b", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"GPIOGet": rhilexlib.WKYWS1608_GPIOGet(e),
			"GPIOSet": rhilexlib.WKYWS1608_GPIOSet(e),
		}
		AddRuleLibToGroup(r, e, "ws1608", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"TFloat": rhilexlib.TruncateFloat(e),
		}
		AddRuleLibToGroup(r, e, "math", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"PlayMusic": rhilexlib.PlayMusic(e),
		}
		AddRuleLibToGroup(r, e, "audio", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Request": rhilexlib.Request(e),
		}
		AddRuleLibToGroup(r, e, "rpc", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Execute": rhilexlib.JqSelect(e),
		}
		AddRuleLibToGroup(r, e, "jq", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Ping": rhilexlib.PingIp(e),
		}
		AddRuleLibToGroup(r, e, "network", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Get":  rhilexlib.HttpGet(e),
			"Post": rhilexlib.HttpPost(e),
		}
		AddRuleLibToGroup(r, e, "http", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Update": rhilexlib.DataSchemaValueUpdate(e),
		}
		AddRuleLibToGroup(r, e, "dataschema", Funcs)
	}
}

/*
*
* 加载外部扩展库
*
 */
func LoadExtLuaLib(e typex.Rhilex, r *typex.Rule) error {
	for _, s := range core.GlobalConfig.Extlibs.Value {
		err := r.LoadExternLuaLib(s)
		if err != nil {
			return err
		}
	}
	return nil
}
