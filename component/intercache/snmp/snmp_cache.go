// Copyright (C) 2023 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) SnmpOid later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT SnmpOid WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package snmp

import (
	"sync"

	"github.com/hootrhino/rhilex/component/intercache"
	"github.com/hootrhino/rhilex/typex"
)

var __DefaultSnmpOidCache *SnmpOidCache

// 点位表
type SnmpOid struct {
	UUID          string
	Status        int // 0 正常；1 错误，填充 ErrMsg
	ErrMsg        string
	LastFetchTime uint64
	Value         string
}

func RegisterSlot(Slot string) {
	__DefaultSnmpOidCache.RegisterSlot(Slot)
}
func GetSlot(Slot string) map[string]SnmpOid {
	return __DefaultSnmpOidCache.GetSlot(Slot)
}
func SetValue(Slot, K string, V SnmpOid) {
	__DefaultSnmpOidCache.SetValue(Slot, K, V)
}
func GetValue(Slot, K string) SnmpOid {
	return __DefaultSnmpOidCache.GetValue(Slot, K)
}
func DeleteValue(Slot, K string) {
	__DefaultSnmpOidCache.DeleteValue(Slot, K)
}
func UnRegisterSlot(Slot string) {
	__DefaultSnmpOidCache.UnRegisterSlot(Slot)
}
func Size() uint64 {
	return __DefaultSnmpOidCache.Size()
}
func Flush() {
	__DefaultSnmpOidCache.Flush()
}

type SnmpOidCache struct {
	Slots      map[string]map[string]SnmpOid
	ruleEngine typex.Rhilex
	locker     sync.Mutex
}

func InitSnmpOidCache(ruleEngine typex.Rhilex) intercache.InterCache {
	__DefaultSnmpOidCache = &SnmpOidCache{
		ruleEngine: ruleEngine,
		Slots:      map[string]map[string]SnmpOid{},
		locker:     sync.Mutex{},
	}
	return __DefaultSnmpOidCache
}
func (M *SnmpOidCache) RegisterSlot(Slot string) {
	M.locker.Lock()
	defer M.locker.Unlock()
	M.Slots[Slot] = map[string]SnmpOid{}
}
func (M *SnmpOidCache) GetSlot(Slot string) map[string]SnmpOid {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		return S
	}
	return nil
}
func (M *SnmpOidCache) SetValue(Slot, K string, V SnmpOid) {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		S[K] = V
		M.Slots[Slot] = S
	}
}
func (M *SnmpOidCache) GetValue(Slot, K string) SnmpOid {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		return S[K]
	}
	return SnmpOid{}
}
func (M *SnmpOidCache) DeleteValue(Slot, K string) {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		delete(S, Slot)
	}
}
func (M *SnmpOidCache) UnRegisterSlot(Slot string) {
	M.locker.Lock()
	defer M.locker.Unlock()
	delete(M.Slots, Slot)
	M.Flush()
}
func (M *SnmpOidCache) Size() uint64 {
	return uint64(len(M.Slots))
}
func (M *SnmpOidCache) Flush() {
	for slotName, slot := range M.Slots {
		for k := range slot {
			delete(slot, k)
		}
		delete(M.Slots, slotName)
	}
}
