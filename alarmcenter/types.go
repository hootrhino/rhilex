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

package alarmcenter

import (
	"time"

	"github.com/hootrhino/rhilex/typex"
	"gorm.io/gorm"
)

/*
*
* Sqlite 数据持久层
*
 */
type SqliteDAO struct {
	engine typex.Rhilex
	name   string   // 框架可以根据名称来选择不同的数据库驱动,为以后扩展准备
	db     *gorm.DB // Sqlite 驱动
}

// AlarmRule 告警规则
type AlarmRule struct {
	Threshold    uint64        // 单次触发的日志数量阈值
	Interval     time.Duration // 最小触发时间间隔
	ExprDefines  []ExprDefine
	HandleId     string    // 事件处理器，目前是北向ID
	lastAlarm    time.Time // 上次告警触发的时间
	pendingCount uint64    // 当前累计的告警数量
}

// NewAlarmRule 创建一个告警规则
func NewAlarmRule(threshold uint64, interval time.Duration,
	HandleId string, ExprDefines []ExprDefine) *AlarmRule {
	return &AlarmRule{
		Threshold:    threshold,
		Interval:     interval,
		ExprDefines:  ExprDefines,
		HandleId:     HandleId,
		lastAlarm:    time.Time{},
		pendingCount: 0,
	}
}

// AddLog 添加告警日志，返回是否需要触发
func (ar *AlarmRule) AddLog() bool {
	ar.pendingCount++
	now := time.Now()

	// 检查是否满足触发条件
	if ar.pendingCount >= ar.Threshold ||
		now.Sub(ar.lastAlarm) >= ar.Interval {
		ar.lastAlarm = now  // 更新最后触发时间
		ar.pendingCount = 0 // 清空计数器
		return true         // 触发告警
	}

	return false
}

// Reset 重置告警状态
func (ar *AlarmRule) Reset() {
	ar.ResetLastAlarm()
	ar.ResetPendingCount()
}

// Reset 重置告警状态
func (ar *AlarmRule) ResetPendingCount() {
	ar.pendingCount = 0
}

// Reset 重置告警状态
func (ar *AlarmRule) ResetLastAlarm() {
	ar.lastAlarm = time.Time{}
}
