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

package engine

import (
	"errors"

	core "github.com/hootrhino/rhilex/config"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

// ┌──────┐    ┌──────┐    ┌──────┐
// │ Init ├───►│ Load ├───►│ Stop │
// └──────┘    └──────┘    └──────┘
func (e *RuleEngine) LoadPlugin(sectionK string, p typex.XPlugin) error {
	section := utils.GetINISection(core.GlobalConfig.IniPath, sectionK)
	if err := p.Init(section); err != nil {
		return err
	}
	_, ok := e.Plugins.Load(p.PluginMetaInfo().UUID)
	if ok {
		return errors.New("plugin already installed:" + p.PluginMetaInfo().Name)
	}
	if err := p.Start(e); err != nil {
		return err
	}
	// Skip License Manager
	if p.PluginMetaInfo().UUID != "LicenseManager" {
		e.Plugins.Store(p.PluginMetaInfo().UUID, p)
		glogger.GLogger.Infof("Plugin start successfully:[%v]", p.PluginMetaInfo().Name)
	}
	return nil

}
