package mqttserver

import (
	"encoding/json"

	"github.com/hootrhino/rhilex/typex"
)

func (s *MqttServer) ListClients(offset, count int) []Client {
	result := []Client{}
	for _, v := range s.mqttServer.Clients.GetAll() {
		c := Client{
			ID:           v.ID,
			Remote:       v.Net.Remote,
			Username:     string(v.Properties.Username),
			CleanSession: v.Properties.Clean,
			Listener:     v.Net.Listener,
		}
		topics := s.topics[v.ID]
		c.Topics = topics
		result = append(result, c)
	}
	return result
}

/*
*
* 把某个客户端给踢下线
*
 */
func (s *MqttServer) KickOut(clientid string) bool {
	C, ok := s.mqttServer.Clients.Get(clientid)
	if ok {
		C.Stop(nil)
		s.mqttServer.Clients.Delete(clientid)
	}
	return true
}

/*
*
* 服务调用接口
*
 */
func (s *MqttServer) Service(arg typex.ServiceArg) typex.ServiceResult {
	// 老API，新版会删除
	if arg.Name == "clients" {
		return typex.ServiceResult{Out: s.ListClients(0, 10)}
	}
	// 新版本API：需要分页查询
	if arg.Name == "PageQueryClients" {
		switch cmd := arg.Args.(type) {
		case string:
			{
				Page := PageRequest{}
				if err := json.Unmarshal([]byte(cmd), &Page); err != nil {
					return typex.ServiceResult{Out: PageResult{
						Current: 1,
						Size:    0,
						Total:   s.mqttServer.Clients.Len(),
						Records: []Client{},
					}}
				}
				return typex.ServiceResult{Out: PageResult{
					Current: Page.Current,
					Size:    Page.Size,
					Total:   s.mqttServer.Clients.Len(),
					Records: s.ListClients(Page.Current, Page.Size),
				}}
			}
		}
		// 默认返回10条数据
		return typex.ServiceResult{Out: PageResult{
			Current: 1,
			Size:    10,
			Total:   s.mqttServer.Clients.Len(),
			Records: s.ListClients(1, 10),
		}}
	}
	if arg.Name == "kickout" {
		switch tt := arg.Args.(type) {
		case []interface{}:
			for _, id := range tt {
				switch idt := id.(type) {
				case string:
					{
						s.KickOut(idt)
					}
				}
			}
		}
	}
	// 向客户端发送消息
	if arg.Name == "publish" {
		switch cmd := arg.Args.(type) {
		case string:
			publishMsg := publishMsg{}
			if err := json.Unmarshal([]byte(cmd), &publishMsg); err != nil {
				return typex.ServiceResult{Out: err}
			}
			if err := s.mqttServer.Publish(publishMsg.Topic,
				[]byte(publishMsg.Msg), false, 1); err != nil {
				return typex.ServiceResult{Out: err}
			}
		}
	}
	return typex.ServiceResult{Out: []Client{Client{}}}
}

// 发布的消息
type publishMsg struct {
	Topic string
	Msg   string
}
