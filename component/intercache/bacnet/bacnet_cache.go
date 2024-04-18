// Copyright (C) 2023 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) BacnetCacheValue later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT BacnetCacheValue WARRANTY; without even the implied warranty of
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

var __DefaultBacnetCache *BacnetCache

type BacnetCacheValue struct {
	UUID          string
	Status        int // 0 正常；1 错误，填充 ErrMsg
	ErrMsg        string
	LastFetchTime uint64
	Value         string
}

func RegisterSlot(Slot string) {
	__DefaultBacnetCache.RegisterSlot(Slot)
}
func GetSlot(Slot string) map[string]BacnetCacheValue {
	return __DefaultBacnetCache.GetSlot(Slot)
}
func SetValue(Slot, K string, V BacnetCacheValue) {
	__DefaultBacnetCache.SetValue(Slot, K, V)
}
func GetValue(Slot, K string) BacnetCacheValue {
	return __DefaultBacnetCache.GetValue(Slot, K)
}
func DeleteValue(Slot, K string) {
	__DefaultBacnetCache.DeleteValue(Slot, K)
}
func UnRegisterSlot(Slot string) {
	__DefaultBacnetCache.UnRegisterSlot(Slot)
}
func Size() uint64 {
	return __DefaultBacnetCache.Size()
}
func Flush() {
	__DefaultBacnetCache.Flush()
}

type BacnetCache struct {
	Slots      map[string]map[string]BacnetCacheValue
	ruleEngine typex.Rhilex
	locker     sync.Mutex
}

func InitBacnetCache(ruleEngine typex.Rhilex) intercache.InterCache {
	__DefaultBacnetCache = &BacnetCache{
		ruleEngine: ruleEngine,
		Slots:      map[string]map[string]BacnetCacheValue{},
		locker:     sync.Mutex{},
	}
	return __DefaultBacnetCache
}
func (M *BacnetCache) RegisterSlot(Slot string) {
	M.locker.Lock()
	defer M.locker.Unlock()
	M.Slots[Slot] = map[string]BacnetCacheValue{}
}
func (M *BacnetCache) GetSlot(Slot string) map[string]BacnetCacheValue {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		return S
	}
	return nil
}
func (M *BacnetCache) SetValue(Slot, K string, V BacnetCacheValue) {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		S[K] = V
		M.Slots[Slot] = S
	}
}
func (M *BacnetCache) GetValue(Slot, K string) BacnetCacheValue {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		return S[K]
	}
	return BacnetCacheValue{}
}
func (M *BacnetCache) DeleteValue(Slot, K string) {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		delete(S, Slot)
	}
}
func (M *BacnetCache) UnRegisterSlot(Slot string) {
	M.locker.Lock()
	defer M.locker.Unlock()
	delete(M.Slots, Slot)
	M.Flush()
}
func (M *BacnetCache) Size() uint64 {
	return uint64(len(M.Slots))
}
func (M *BacnetCache) Flush() {
	for slotName, slot := range M.Slots {
		for k := range slot {
			delete(slot, k)
		}
		delete(M.Slots, slotName)
	}
}
