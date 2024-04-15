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

package shellymanager

type ShellyDeviceInfo struct {
	Ip         string  `json:"ip"` // 扫描出来的IP
	Name       *string `json:"name"`
	ID         string  `json:"id"`
	Mac        string  `json:"mac"`
	Slot       int     `json:"slot"`
	Model      string  `json:"model"`
	Gen        int     `json:"gen"`
	FwID       string  `json:"fw_id"`
	Ver        string  `json:"ver"`
	App        string  `json:"app"`
	AuthEn     bool    `json:"auth_en"`
	AuthDomain *string `json:"auth_domain"`
}
type ShellyDeviceStatus struct {
	Mac              string `json:"mac"`
	RestartRequired  bool   `json:"restart_required"`
	Time             string `json:"time"`
	Unixtime         int64  `json:"unixtime"`
	Uptime           int    `json:"uptime"`
	RamSize          int    `json:"ram_size"`
	RamFree          int    `json:"ram_free"`
	FsSize           int    `json:"fs_size"`
	FsFree           int    `json:"fs_free"`
	CfgRev           int    `json:"cfg_rev"`
	KvsRev           int    `json:"kvs_rev"`
	ScheduleRev      int    `json:"schedule_rev"`
	WebhookRev       int    `json:"webhook_rev"`
	AvailableUpdates struct {
		Stable struct {
			Version string `json:"version"`
		} `json:"stable"`
	} `json:"available_updates"`
}
