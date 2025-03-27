// Copyright (C) 2025 wwhai
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

package engine

import (
	"github.com/hootrhino/rhilex/alarmcenter"
	"github.com/hootrhino/rhilex/applet"
	"github.com/hootrhino/rhilex/component/aibase"
	"github.com/hootrhino/rhilex/component/crontask"
	"github.com/hootrhino/rhilex/component/eventbus"
	intercache "github.com/hootrhino/rhilex/component/intercache"
	"github.com/hootrhino/rhilex/component/interdb"
	"github.com/hootrhino/rhilex/component/interkv"
	"github.com/hootrhino/rhilex/component/intermetric"
	"github.com/hootrhino/rhilex/component/internotify"
	"github.com/hootrhino/rhilex/component/interqueue"
	"github.com/hootrhino/rhilex/component/lostcache"
	"github.com/hootrhino/rhilex/component/security"
	supervisor "github.com/hootrhino/rhilex/component/supervisor"
	core "github.com/hootrhino/rhilex/config"
	datacenter "github.com/hootrhino/rhilex/datacenter"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/plugin"
	"github.com/hootrhino/rhilex/registry"
	transceiver "github.com/hootrhino/rhilex/transceiver"
	"github.com/hootrhino/rhilex/typex"
)

func InitAllComponent(__DefaultRuleEngine typex.Rhilex) {
	// Init Security License
	security.InitSecurityLicense()
	// Init EventBus
	eventbus.InitEventBus(__DefaultRuleEngine)
	// Init Internal DB
	interdb.InitAll(__DefaultRuleEngine)
	// Init Alarm Center
	alarmcenter.InitAll(__DefaultRuleEngine)
	// Init Notify Center
	internotify.InitAll(__DefaultRuleEngine)
	// Init Data Center
	datacenter.InitAll(__DefaultRuleEngine)
	// Init Lost Cache
	lostcache.InitAll(__DefaultRuleEngine)
	// Init Alarm Center
	alarmcenter.InitAlarmCenter(__DefaultRuleEngine)
	// Data center: future version maybe support
	datacenter.InitDataCenter(__DefaultRuleEngine)
	// Internal kv Store
	interkv.InitInterKVStore(core.GlobalConfig.MaxKvStoreSize)
	// SuperVisor Admin
	supervisor.InitResourceSuperVisorAdmin(__DefaultRuleEngine)
	// Init Global Value Registry
	intercache.InitGlobalValueRegistry(__DefaultRuleEngine)
	// Internal Metric
	intermetric.InitInternalMetric(__DefaultRuleEngine)
	// lua applet manager
	applet.InitAppletRuntime(__DefaultRuleEngine)
	// current only support Internal ai
	aibase.InitAlgorithmRuntime(__DefaultRuleEngine)
	// Internal Queue
	interqueue.InitXQueue(__DefaultRuleEngine, core.GlobalConfig.MaxQueueSize)
	// Init Transceiver Communicator Manager
	transceiver.InitTransceiverManager(__DefaultRuleEngine)
	// Init Device Registry
	registry.InitDeviceRegistry(__DefaultRuleEngine)
	// Init Source Registry
	registry.InitSourceRegistry(__DefaultRuleEngine)
	// Init Target Registry
	registry.InitTargetRegistry(__DefaultRuleEngine)
	// Init Plugin Registry
	plugin.InitPluginRegistry(__DefaultRuleEngine)
}
func StartAllComponent() {
	// Internal BUS
	interqueue.StartXQueue()
}
func StopAllComponent() {
	crontask.StopCronRebootExecutor()
	supervisor.StopSupervisorAdmin()
	applet.Stop()
	intercache.Flush()
	aibase.Stop()
	transceiver.Stop()
	alarmcenter.StopAlarmCenter()
	plugin.Stop()
	interdb.StopAll()
	alarmcenter.StopAll()
	datacenter.StopAll()
	lostcache.StopAll()
	internotify.StopAll()
	eventbus.Stop()
	glogger.Close()
}
