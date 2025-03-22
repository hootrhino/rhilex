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

package typex

import (
	lua "github.com/hootrhino/gopher-lua"
)

type RuleStatus int

// error: index out of range [_VM_Registry_MaxSize] with length
const _VM_Registry_Size int = 1024 * 1024    // 默认堆栈大小
const _VM_Registry_MaxSize int = 1024 * 1024 // 默认最大堆栈
const _VM_Registry_GrowStep int = 32         // 默认CPU消耗

// RULE_STOP 0: 停止
// RULE_RUNNING 1: 运行中
const RULE_STOP RuleStatus = 0
const RULE_RUNNING RuleStatus = 1

// 规则描述
type Rule struct {
	Id          string      `json:"id"`
	UUID        string      `json:"uuid"`
	Type        string      `json:"type"` // 脚本类型，目前支持"lua"
	Status      RuleStatus  `json:"status"`
	Name        string      `json:"name"`
	FromSource  string      `json:"fromSource"` // 来自数据源
	FromDevice  string      `json:"fromDevice"` // 来自设备
	Actions     string      `json:"actions"`
	Success     string      `json:"success"`
	Failed      string      `json:"failed"`
	Description string      `json:"description"`
	LuaVM       *lua.LState `json:"-"` // Lua VM
}

func NewLuaRule(e Rhilex,
	uuid string,
	name string,
	description string,
	fromSource string,
	fromDevice string,
	success string,
	actions string,
	failed string) *Rule {
	rule := NewRule(e,
		uuid,
		name,
		description,
		fromSource,
		fromDevice,
		success,
		actions,
		failed)
	return rule
}

// New
func NewRule(e Rhilex,
	uuid string,
	name string,
	description string,
	fromSource string,
	fromDevice string,
	success string,
	actions string,
	failed string) *Rule {
	return &Rule{
		UUID:        uuid,
		Name:        name,
		Type:        "lua", // 默认执行lua脚本
		Description: description,
		FromSource:  fromSource,
		FromDevice:  fromDevice,
		Status:      RULE_RUNNING, // 默认为启用
		Actions:     actions,
		Success:     success,
		Failed:      failed,
		LuaVM: lua.NewState(lua.Options{
			// IncludeGoStackTrace: true,
			RegistrySize:     _VM_Registry_Size,
			RegistryMaxSize:  _VM_Registry_MaxSize,
			RegistryGrowStep: _VM_Registry_GrowStep,
		}),
	}
}

/*
*
* AddLib: 根据 KV形式加载库(推荐)
*  - Global: 命名空间
*   - funcName: 函数名称
 */
func (r *Rule) AddLib(rx Rhilex, ModuleName string, funcName string,
	f func(l *lua.LState) int) {

	r.LuaVM.PreloadModule(ModuleName, func(L *lua.LState) int {
		table := r.LuaVM.NewTable()
		table.RawSetString(funcName, r.LuaVM.NewClosure(f))
		L.Push(table)
		return 1
	})
}
