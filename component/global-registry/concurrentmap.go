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
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package globalregistry

import (
	"sync"
)

type ConcurrentMap struct {
	sync.RWMutex
	data map[string]interface{}
	keys []string
}

func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{
		data: make(map[string]interface{}),
		keys: make([]string, 0),
	}
}

func (m *ConcurrentMap) Set(key string, value interface{}) {
	m.Lock()
	defer m.Unlock()
	m.data[key] = value

	// 更新keys的顺序
	found := false
	for i, k := range m.keys {
		if k == key {
			m.keys[i] = key
			found = true
			break
		}
	}
	if !found {
		m.keys = append(m.keys, key)
	}
}

func (m *ConcurrentMap) Get(key string) (interface{}, bool) {
	m.RLock()
	defer m.RUnlock()
	value, ok := m.data[key]
	return value, ok
}

func (m *ConcurrentMap) Delete(key string) {
	m.Lock()
	defer m.Unlock()
	delete(m.data, key)

	for i, k := range m.keys {
		if k == key {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

func (m *ConcurrentMap) Keys() []string {
	m.RLock()
	defer m.RUnlock()
	keys := make([]string, len(m.keys))
	copy(keys, m.keys)
	return keys
}
func (m *ConcurrentMap) List() []interface{} {
	List := []interface{}{}
	m.RLock()
	defer m.RUnlock()
	for _, k := range m.keys {
		V := m.data[k]
		if V != nil {
			List = append(List, V)
		}
	}
	return List
}

func (m *ConcurrentMap) Clear() {
	m.RLock()
	defer m.RUnlock()
	for _, k := range m.keys {
		delete(m.data, k)
	}
	for i := 0; i < len(m.keys); i++ {
		m.keys = append(m.keys[:i], m.keys[i+1:]...)
	}
}
