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

package intercache

import (
	"sync"

	"github.com/hootrhino/rhilex/typex"
)

type CacheValue struct {
	UUID          string
	Status        int // 0 正常；1 错误，填充 ErrMsg
	ErrMsg        string
	LastFetchTime uint64
	Value         string
}

/*
*
* 内部缓存器
*
 */
type InterCache interface {
	RegisterSlot(Slot string)      // 存储槽位, 释放资源的时候调用
	UnRegisterSlot(Slot string)    // 注销存储槽位, 释放资源的时候调用
	Size() uint64                  // 存储器当前长度
	Flush()                        // 释放存储器空间
	ClearSlot()                    // 清空槽位
	List() []map[string]CacheValue // 列表
}

var __DefaultHnc8CnCPointCache *GlobalValueRegistry

func RegisterSlot(Slot string) {
	__DefaultHnc8CnCPointCache.RegisterSlot(Slot)
}
func UnRegisterSlot(Slot string) {
	__DefaultHnc8CnCPointCache.UnRegisterSlot(Slot)
}
func GetSlot(Slot string) map[string]CacheValue {
	return __DefaultHnc8CnCPointCache.GetSlot(Slot)
}
func SetValue(Slot, K string, V CacheValue) {
	__DefaultHnc8CnCPointCache.SetValue(Slot, K, V)
}
func GetValue(Slot, K string) CacheValue {
	return __DefaultHnc8CnCPointCache.GetValue(Slot, K)
}
func DeleteValue(Slot, K string) {
	__DefaultHnc8CnCPointCache.DeleteValue(Slot, K)
}

func Size() uint64 {
	return __DefaultHnc8CnCPointCache.Size()
}
func Flush() {
	__DefaultHnc8CnCPointCache.Flush()
}

type GlobalValueRegistry struct {
	Slots      map[string]map[string]CacheValue
	slotKeys   []string
	ruleEngine typex.Rhilex
	locker     sync.RWMutex
}

func InitGlobalValueRegistry(ruleEngine typex.Rhilex) InterCache {
	__DefaultHnc8CnCPointCache = &GlobalValueRegistry{
		ruleEngine: ruleEngine,
		Slots:      map[string]map[string]CacheValue{},
		locker:     sync.RWMutex{},
		slotKeys:   []string{},
	}
	return __DefaultHnc8CnCPointCache
}
func (M *GlobalValueRegistry) RegisterSlot(Slot string) {
	M.locker.Lock()
	defer M.locker.Unlock()
	M.Slots[Slot] = map[string]CacheValue{}
}
func (M *GlobalValueRegistry) GetSlot(Slot string) map[string]CacheValue {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		return S
	}
	return nil
}
func (M *GlobalValueRegistry) SetValue(Slot, K string, V CacheValue) {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		S[K] = V
		M.Slots[Slot] = S
	}
	found := false
	for i, k := range M.slotKeys {
		if k == K {
			M.slotKeys[i] = K
			found = true
			break
		}
	}
	if !found {
		M.slotKeys = append(M.slotKeys, K)
	}
}
func (M *GlobalValueRegistry) GetValue(Slot, K string) CacheValue {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		return S[K]
	}
	return CacheValue{}
}
func (M *GlobalValueRegistry) DeleteValue(Slot, K string) {
	M.locker.Lock()
	defer M.locker.Unlock()
	if S, ok := M.Slots[Slot]; ok {
		delete(S, Slot)
	}
	for i, k := range M.slotKeys {
		if k == K {
			M.slotKeys = append(M.slotKeys[:i], M.slotKeys[i+1:]...)
			break
		}
	}
}
func (M *GlobalValueRegistry) List() []map[string]CacheValue {
	CacheValues := []map[string]CacheValue{}
	M.locker.RLock()
	defer M.locker.RUnlock()
	for _, k := range M.slotKeys {
		V := M.Slots[k]
		if V != nil {
			CacheValues = append(CacheValues, V)
		}
	}
	return CacheValues
}
func (M *GlobalValueRegistry) UnRegisterSlot(SlotName string) {
	M.locker.Lock()
	defer M.locker.Unlock()
	Slot := M.Slots[SlotName]
	for ks := range Slot {
		delete(Slot, ks)
	}
	delete(M.Slots, SlotName)
	for i := 0; i < len(M.slotKeys); i++ {
		M.slotKeys = append(M.slotKeys[:i], M.slotKeys[i+1:]...)
	}
}
func (M *GlobalValueRegistry) Size() uint64 {
	return uint64(len(M.Slots))
}

func (M *GlobalValueRegistry) ClearSlot() {
	M.locker.Lock()
	defer M.locker.Unlock()
	for _, sk := range M.slotKeys {
		for ks := range M.Slots[sk] {
			delete(M.Slots[sk], ks)
		}
		delete(M.Slots, sk)
	}
	for i := 0; i < len(M.slotKeys); i++ {
		M.slotKeys = append(M.slotKeys[:i], M.slotKeys[i+1:]...)
	}
}

func (M *GlobalValueRegistry) Flush() {
	for slotName, slot := range M.Slots {
		for k := range slot {
			delete(slot, k)
		}
		delete(M.Slots, slotName)
	}
}
