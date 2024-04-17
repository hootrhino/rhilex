// Copyright (C) 2023 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) ShellyDevice later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ShellyDevice WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package shellymanager

import (
	"encoding/json"
	"sync"

	"github.com/hootrhino/rhilex/typex"
)

type IOPort struct {
	Name   string
	Status bool
}
type ShellyDevice struct {
	Ip         string   `json:"ip"` // 扫描出来的IP
	Name       *string  `json:"name"`
	ID         string   `json:"id"`
	Mac        string   `json:"mac"`
	Slot       int      `json:"slot"`
	Model      string   `json:"model"`
	Gen        int      `json:"gen"`
	FwID       string   `json:"fw_id"`
	Ver        string   `json:"ver"`
	App        string   `json:"app"`
	AuthEn     bool     `json:"auth_en"`
	AuthDomain *string  `json:"auth_domain"`
	Input      []IOPort `json:"input"`
	Switch     []IOPort `json:"switch"`
}

func (device ShellyDevice) String() string {
	if bytes, err := json.Marshal(device); err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}

var __DefaultShellyDeviceRegistry *ShellyDeviceRegistry

func RegisterSlot(Slot string) {
	__DefaultShellyDeviceRegistry.RegisterSlot(Slot)
}
func GetSlot(Slot string) map[string]ShellyDevice {
	return __DefaultShellyDeviceRegistry.GetSlot(Slot)
}
func SetValue(Slot, K string, V ShellyDevice) {
	__DefaultShellyDeviceRegistry.SetValue(Slot, K, V)
}
func GetValue(Slot, K string) ShellyDevice {
	return __DefaultShellyDeviceRegistry.GetValue(Slot, K)
}
func Exists(Slot, K string) bool {
	if Slots, ok1 := __DefaultShellyDeviceRegistry.Slots[Slot]; ok1 {
		if _, ok2 := Slots[K]; ok2 {
			return true
		}
	}
	return false
}
func DeleteValue(Slot, K string) {
	__DefaultShellyDeviceRegistry.DeleteValue(Slot, K)
}
func UnRegisterSlot(Slot string) {
	__DefaultShellyDeviceRegistry.UnRegisterSlot(Slot)
}
func Size() uint64 {
	return __DefaultShellyDeviceRegistry.Size()
}
func Flush() {
	__DefaultShellyDeviceRegistry.Flush()
}

//Modbus 点位运行时存储器

type ShellyDeviceRegistry struct {
	Slots      map[string]map[string]ShellyDevice // MAC : {MAC, Info}
	keys       []string
	ruleEngine typex.RuleX
	locker     sync.RWMutex
	status     string // SCANNING | DONE
}

func InitShellyDeviceRegistry(ruleEngine typex.RuleX) *ShellyDeviceRegistry {
	__DefaultShellyDeviceRegistry = &ShellyDeviceRegistry{
		ruleEngine: ruleEngine,
		Slots:      map[string]map[string]ShellyDevice{},
		locker:     sync.RWMutex{},
		status:     "DONE",
		keys:       make([]string, 0),
	}
	__DefaultShellyDeviceRegistry.TestAlive()
	return __DefaultShellyDeviceRegistry
}
func (M *ShellyDeviceRegistry) GetStatus() string {
	M.locker.Lock()
	defer M.locker.Unlock()
	return M.status
}
func (M *ShellyDeviceRegistry) RegisterSlot(Slot string) {
	M.locker.Lock()
	defer M.locker.Unlock()
	M.Slots[Slot] = map[string]ShellyDevice{}
}
func (M *ShellyDeviceRegistry) GetSlot(Slot string) map[string]ShellyDevice {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		return S
	}
	return nil
}
func (M *ShellyDeviceRegistry) SetValue(Slot, K string, V ShellyDevice) {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		S[K] = V
		M.Slots[Slot] = S
	}
	found := false
	for i, k := range M.keys {
		if k == K {
			M.keys[i] = K
			found = true
			break
		}
	}
	if !found {
		M.keys = append(M.keys, K)
	}
}
func (M *ShellyDeviceRegistry) GetValue(Slot, K string) ShellyDevice {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		return S[K]
	}
	return ShellyDevice{}
}
func (M *ShellyDeviceRegistry) DeleteValue(Slot, K string) {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		delete(S, Slot)
	}
	for i, k := range M.keys {
		if k == K {
			M.keys = append(M.keys[:i], M.keys[i+1:]...)
			break
		}
	}
}

func (M *ShellyDeviceRegistry) UnRegisterSlot(Slot string) {
	M.locker.Lock()
	defer M.locker.Unlock()
	delete(M.Slots, Slot)
	M.Flush()
}
func (M *ShellyDeviceRegistry) Size() uint64 {
	return uint64(len(M.Slots))
}
func (M *ShellyDeviceRegistry) Flush() {
	for slotName, slot := range M.Slots {
		for k := range slot {
			delete(slot, k)
		}
		delete(M.Slots, slotName)
	}
}

/*
*
* 顺序Map
*
 */
func (M *ShellyDeviceRegistry) ListAllSlots() []map[string]ShellyDevice {
	M.locker.Lock()
	defer M.locker.Unlock()
	List := []map[string]ShellyDevice{}
	for _, k := range M.keys {
		V := M.Slots[k]
		if V != nil {
			List = append(List, V)
		}
	}
	return List
}
func (M *ShellyDeviceRegistry) ListAllValues() []ShellyDevice {
	M.locker.Lock()
	defer M.locker.Unlock()
	List := []ShellyDevice{}
	for _, k := range M.keys {
		Slot := M.Slots[k]
		for _, Value := range Slot {
			List = append(List, Value)
		}
	}
	return List
}
func (M *ShellyDeviceRegistry) Keys() []string {
	M.locker.Lock()
	defer M.locker.Unlock()
	keys := make([]string, len(M.keys))
	copy(keys, M.keys)
	return keys
}
