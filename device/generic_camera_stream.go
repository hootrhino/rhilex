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

package device

import (
	"github.com/hootrhino/rhilex/typex"
)

type videoCamera struct {
	typex.XStatus
	status typex.SourceState
}

/*
*
* ARM32不支持
*
 */
func NewVideoCamera(e typex.Rhilex) typex.XDevice {
	hd := new(videoCamera)
	hd.RuleEngine = e
	return hd
}

func (hd *videoCamera) Init(devId string, configMap map[string]any) error {
	hd.PointId = devId
	return nil
}

func (hd *videoCamera) Start(cctx typex.CCTX) error {
	hd.Ctx = cctx.Ctx
	hd.CancelCTX = cctx.CancelCTX

	hd.status = typex.SOURCE_UP
	return nil
}

// 设备当前状态
func (hd *videoCamera) Status() typex.SourceState {
	return hd.status
}

// 停止设备
func (hd *videoCamera) Stop() {
	hd.status = typex.SOURCE_DOWN
	hd.CancelCTX()
}

// 真实设备
func (hd *videoCamera) Details() *typex.Device {
	return hd.RuleEngine.GetDevice(hd.PointId)
}

// 状态
func (hd *videoCamera) SetState(status typex.SourceState) {
	hd.status = status

}

// --------------------------------------------------------------------------------------------------
//
// --------------------------------------------------------------------------------------------------
func (hd *videoCamera) OnDCACall(UUID string, Command string, Args any) typex.DCAResult {
	return typex.DCAResult{}
}
func (hd *videoCamera) OnCtrl(cmd []byte, args []byte) ([]byte, error) {
	return []byte{}, nil
}
