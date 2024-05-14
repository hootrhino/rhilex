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

package target

import (
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

type smtpConfig struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Subject  string `json:"subject"`
	From     string `json:"from"`
	To       string `json:"to"`
}
type SmtpTarget struct {
	typex.XStatus
	mainConfig smtpConfig
	status     typex.SourceState
}

// {
//     Server:   "smtp.example.com",       // 您的SMTP服务器地址
//     Port:     587,                      // SMTP端口，通常是587或465
//     User:     "your-email@example.com", // 您的SMTP用户名
//     Password: "your-password",          // 您的SMTP密码
//     Subject:  "Hello from Go",          // 邮件主题
//     From:     "sender@example.com",     // 发件人地址
//     To:       "recipient@example.com",  // 收件人地址
// }

func NewSmtpTarget(e typex.Rhilex) typex.XTarget {
	ht := new(SmtpTarget)
	ht.RuleEngine = e
	ht.mainConfig = smtpConfig{}
	ht.status = typex.SOURCE_DOWN
	return ht
}

func (ht *SmtpTarget) Init(outEndId string, configMap map[string]interface{}) error {
	ht.PointId = outEndId
	if err := utils.BindSourceConfig(configMap, &ht.mainConfig); err != nil {
		return err
	}
	return nil
}
func (ht *SmtpTarget) Start(cctx typex.CCTX) error {
	ht.Ctx = cctx.Ctx
	ht.CancelCTX = cctx.CancelCTX
	ht.status = typex.SOURCE_UP
	glogger.GLogger.Info("Smtp Target started")
	return nil
}

func (ht *SmtpTarget) Status() typex.SourceState {
	if err := CheckSMTPConnection(ht.mainConfig.Server,
		ht.mainConfig.Port); err != nil {
		glogger.GLogger.Error(err)
		return typex.SOURCE_DOWN
	}
	return typex.SOURCE_UP
}
func (ht *SmtpTarget) To(data interface{}) (interface{}, error) {
	switch data.(type) {
	case string:
		err := sendMail(ht.mainConfig.From, ht.mainConfig.To, ht.mainConfig.Subject, "",
			ht.mainConfig.Server, ht.mainConfig.Port, ht.mainConfig.User, ht.mainConfig.Password)
		return nil, err
	default:
		return nil, fmt.Errorf("email content must plain txt type")
	}
}

func (ht *SmtpTarget) Stop() {
	ht.status = typex.SOURCE_DOWN
	if ht.CancelCTX != nil {
		ht.CancelCTX()
	}
}
func (ht *SmtpTarget) Details() *typex.OutEnd {
	return ht.RuleEngine.GetOutEnd(ht.PointId)
}

func sendMail(from, to, subject, body string,
	smtpHost string, smtpPort int, smtpUser, smtpPass string) error {
	msg := []byte("From: " + from + "\nTo: " + to + "\nSubject: " + subject + "\n\n" + body)
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(fmt.Sprintf("%s%d", smtpHost, smtpPort), auth, from, []string{to}, msg)
	if err != nil {
		return err
	}
	return nil
}

// CheckSMTPConnection
func CheckSMTPConnection(smtpServer string, smtpPort int) error {
	timeout := 3 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", smtpServer, smtpPort), timeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
