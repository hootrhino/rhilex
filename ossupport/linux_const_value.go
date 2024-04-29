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

package ossupport

/*
*
* Linux系统下的一些和应用交互的系统级路径
*
 */
const (
	// rhilex 工作目录
	MainWorkDir = "/usr/local"
	// RHILEX 的配置数据库
	RunDbPath = "/usr/local/rhilex.db"
	// 固件保存路径
	FirmwarePath = "/usr/local/upload/Firmware/Firmware.zip"
	// 升级日志
	UpgradeLogPath = "/usr/local/local-upgrade-log.txt"
	// 运行时日志
	RunningLogPath = "/usr/local/rhilex-log.txt"
	// 数据恢复日志
	RecoverLogPath = "/usr/local/local-recover-log.txt"
	// 备份锁
	BackupLockPath = "/var/run/rhilex-upgrade.lock"
	// 升级锁
	UpgradeLockPath = BackupLockPath
	// 备份数据库
	RecoveryDbPath = "/usr/local/upload/Backup/recovery.db"
)
