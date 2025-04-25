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
package rhilex

import (
	"encoding/json"

	"gopkg.in/ini.v1"
)

type RhilexConfig struct {
	AppId         string `ini:"app_id" json:"appId"`                  // Application ID
	MaxQueueSize  int    `ini:"max_queue_size" json:"maxQueueSize"`   // Max size of the queue for incoming messages
	GomaxProcs    int    `ini:"gomax_procs" json:"gomaxProcs"`        // Number of OS threads to use for goroutines
	EnablePProf   bool   `ini:"enable_pprof" json:"enablePProf"`      // Enable pprof: true or false
	DebugMode     bool   `ini:"debug_mode" json:"appDebugMode"`       // Debug mode: true or false
	LogLevel      string `ini:"log_level" json:"logLevel"`            // Log level: debug, info, warn, error, fatal
	LogMaxSize    int    `ini:"log_max_size" json:"logMaxSize"`       // Max size of log file in MB
	LogMaxBackups int    `ini:"log_max_backups" json:"logMaxBackups"` // Max number of backup log files
	LogMaxAge     int    `ini:"log_max_age" json:"logMaxAge"`         // Max age of log files in days
}

// LoadConfig loads the configuration from the specified ini file
func LoadConfig(iniPath string) (RhilexConfig, error) {
	cfg, err := ini.Load(iniPath)
	if err != nil {
		return RhilexConfig{}, err
	}

	var rhilexConfig RhilexConfig
	err = cfg.MapTo(&rhilexConfig)
	if err != nil {
		return RhilexConfig{}, err
	}

	return rhilexConfig, nil
}

// json string
func (c *RhilexConfig) ToJson() string {
	jsonString, err := json.Marshal(c)
	if err != nil {
		return `{"error": "failed to convert to JSON"}`
	}
	return string(jsonString)
}
