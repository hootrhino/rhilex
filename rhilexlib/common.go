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

package rhilexlib

import (
	"errors"

	"github.com/hootrhino/rhilex/component/interqueue"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
)

func handleDataFormat(e typex.Rhilex, uuid string, incoming string) error {
	outEnd := e.GetOutEnd(uuid)
	if outEnd != nil {
		return interqueue.PushOutQueue(outEnd, incoming)
	}
	msg := "target not found:" + uuid
	glogger.GLogger.Error(msg)
	return errors.New(msg)

}
