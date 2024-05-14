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
	core "github.com/hootrhino/rhilex/config"
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
			"ToHttp":     rhilexlib.DataToHttp(e, r.UUID),
			"ToMqtt":     rhilexlib.DataToMqtt(e, r.UUID),
			"ToUdp":      rhilexlib.DataToUdp(e, r.UUID),
			"ToTcp":      rhilexlib.DataToTcp(e, r.UUID),
			"ToTdEngine": rhilexlib.DataToTdEngine(e, r.UUID),
			"ToMongo":    rhilexlib.DataToMongo(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "data", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Debug": rhilexlib.DebugRule(e, r.UUID),
			"Throw": rhilexlib.Throw(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "_G", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"VSet":             rhilexlib.StoreSet(e, r.UUID),
			"VSetWithDuration": rhilexlib.StoreSetWithDuration(e, r.UUID),
			"VGet":             rhilexlib.StoreGet(e, r.UUID),
			"VDel":             rhilexlib.StoreDelete(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "kv", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Time":       rhilexlib.Time(e, r.UUID),
			"TimeMs":     rhilexlib.TimeMs(e, r.UUID),
			"TsUnix":     rhilexlib.TsUnix(e, r.UUID),
			"TsUnixNano": rhilexlib.TsUnixNano(e, r.UUID),
			"NtpTime":    rhilexlib.NtpTime(e, r.UUID),
			"Sleep":      rhilexlib.Sleep(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "time", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"HToN":       rhilexlib.HToN(e, r.UUID),
			"HsubToN":    rhilexlib.HsubToN(e, r.UUID),
			"MatchHex":   rhilexlib.MatchHex(e, r.UUID),
			"MatchUInt":  rhilexlib.MatchUInt(e, r.UUID),
			"Bytes2Hexs": rhilexlib.Bytes2Hexs(e, r.UUID),
			"Hexs2Bytes": rhilexlib.Hexs2Bytes(e, r.UUID),
			"ABCD":       rhilexlib.ABCD(e, r.UUID),
			"DCBA":       rhilexlib.DCBA(e, r.UUID),
			"BADC":       rhilexlib.BADC(e, r.UUID),
			"CDAB":       rhilexlib.CDAB(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "hex", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"MB":            rhilexlib.MatchBinary(e, r.UUID),
			"MBHex":         rhilexlib.MatchBinaryHex(e, r.UUID),
			"B2BS":          rhilexlib.ByteToBitString(e, r.UUID),
			"Bit":           rhilexlib.GetABitOnByte(e, r.UUID),
			"B2I64":         rhilexlib.ByteToInt64(e, r.UUID),
			"B64S2B":        rhilexlib.B64S2B(e, r.UUID),
			"BS2B":          rhilexlib.BitStringToBytes(e, r.UUID),
			"Bin2F32":       rhilexlib.BinToFloat32(e, r.UUID),
			"Bin2F64":       rhilexlib.BinToFloat64(e, r.UUID),
			"Bin2F32Big":    rhilexlib.BinToFloat32(e, r.UUID),
			"Bin2F64Big":    rhilexlib.BinToFloat64(e, r.UUID),
			"Bin2F32Little": rhilexlib.BinToFloat32Little(e, r.UUID),
			"Bin2F64Little": rhilexlib.BinToFloat64Little(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "binary", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"T2J": rhilexlib.JSONE(e, r.UUID),
			"J2T": rhilexlib.JSOND(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "json", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"ReadDevice":  rhilexlib.ReadDevice(e, r.UUID),
			"WriteDevice": rhilexlib.WriteDevice(e, r.UUID),
			"CtrlDevice":  rhilexlib.CtrlDevice(e, r.UUID),
			"ReadSource":  rhilexlib.ReadSource(e, r.UUID),
			"WriteSource": rhilexlib.WriteSource(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "device", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"T2Str":   rhilexlib.T2Str(e, r.UUID),
			"Bin2Str": rhilexlib.Bin2Str(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "string", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"F5":  rhilexlib.F5(e, r.UUID),
			"F6":  rhilexlib.F6(e, r.UUID),
			"F15": rhilexlib.F15(e, r.UUID),
			"F16": rhilexlib.F16(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "modbus", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"DO1Set":  rhilexlib.RHILEXG1_DO1Set(e, r.UUID),
			"DO1Get":  rhilexlib.RHILEXG1_DO1Get(e, r.UUID),
			"DO2Set":  rhilexlib.RHILEXG1_DO2Set(e, r.UUID),
			"DO2Get":  rhilexlib.RHILEXG1_DO2Get(e, r.UUID),
			"DI1Get":  rhilexlib.H3DI1Get(e, r.UUID),
			"DI2Get":  rhilexlib.H3DI2Get(e, r.UUID),
			"DI3Get":  rhilexlib.H3DI3Get(e, r.UUID),
			"Led1On":  rhilexlib.Led1On(e, r.UUID),
			"Led1Off": rhilexlib.Led1Off(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "rhinopi", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"XOR":   rhilexlib.XOR(e, r.UUID),
			"CRC16": rhilexlib.CRC16(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "misc", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"GPIOGet": rhilexlib.RASPI4_GPIOGet(e, r.UUID),
			"GPIOSet": rhilexlib.RASPI4_GPIOSet(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "raspi4b", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"GPIOGet": rhilexlib.WKYWS1608_GPIOGet(e, r.UUID),
			"GPIOSet": rhilexlib.WKYWS1608_GPIOSet(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "ws1608", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"TFloat": rhilexlib.TruncateFloat(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "math", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"PlayMusic": rhilexlib.PlayMusic(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "audio", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Request": rhilexlib.Request(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "rpc", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Execute": rhilexlib.JqSelect(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "jq", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Ping": rhilexlib.PingIp(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "network", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Get":  rhilexlib.HttpGet(e, r.UUID),
			"Post": rhilexlib.HttpPost(e, r.UUID),
		}
		AddRuleLibToGroup(r, e, "http", Funcs)
	}
	{
		Funcs := map[string]func(l *lua.LState) int{
			"Update": rhilexlib.DataSchemaValueUpdate(e, r.UUID),
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
