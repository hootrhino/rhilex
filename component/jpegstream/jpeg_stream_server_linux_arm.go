package jpegstream

import (
	"fmt"

	"github.com/hootrhino/rhilex/component"
	"github.com/hootrhino/rhilex/typex"
)

var __DefaultJpegStreamServer *JpegStreamServer

type JpegStreamServer struct {
}

func InitJpegStreamServer(rhilex typex.Rhilex) {
	__DefaultJpegStreamServer = &JpegStreamServer{}
}

func (s *JpegStreamServer) Init(cfg map[string]any) error {

	return nil
}
func (s *JpegStreamServer) Start(r typex.Rhilex) error {

	return nil
}
func (s *JpegStreamServer) Stop() error {
	return nil
}
func (s *JpegStreamServer) PluginMetaInfo() component.XComponentMetaInfo {
	return component.XComponentMetaInfo{}
}

/*
*
* Manage API
*
 */

func (s *JpegStreamServer) RegisterJpegStreamSource(liveId string) error {

	return fmt.Errorf("stream already exists")
}

func (s *JpegStreamServer) GetJpegStreamSource(liveId string) (*JpegStream, error) {

	return nil, nil

}

func (s *JpegStreamServer) Exists(liveId string) bool {
	return true
}
func (s *JpegStreamServer) DeleteJpegStreamSource(liveId string) {

}

func (s *JpegStreamServer) JpegStreamSourceList() []JpegStream {
	return nil
}
func (s *JpegStreamServer) JpegStreamFlush() {
}
