package main

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

import (
	"context"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	archsupport "github.com/hootrhino/rhilex/bspsupport"
	"github.com/hootrhino/rhilex/engine"
	"github.com/hootrhino/rhilex/ossupport"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
	"github.com/urfave/cli/v2"
)

func init() {

	go func() {
		for {
			select {
			case <-context.Background().Done():
				return
			default:
				time.Sleep(30 * time.Second)
				runtime.GC()
			}
		}
	}()
	env := os.Getenv("ARCHSUPPORT")
	typex.DefaultVersionInfo.Product = archsupport.CheckVendor(env)
	dist, err := utils.GetOSDistribution()
	if err != nil {
		utils.CLog("Failed to Get OS Distribution:%s", err)
		os.Exit(1)
	}
	typex.DefaultVersionInfo.Dist = dist
	arch := fmt.Sprintf("%s-%s", typex.DefaultVersionInfo.Dist, runtime.GOARCH)
	typex.DefaultVersionInfo.Arch = arch
}

//
//go:generate bash ./gen_info.sh
func main() {
	app := &cli.App{
		Name:  "RHILEX Gateway FrameWork",
		Usage: "Homepage: http://rhilex.hootrhino.com",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Start rhilex, Must with config: -config=path/rhilex.ini",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "db",
						Usage: "Database of rhilex",
						Value: "rhilex.db",
					},
					&cli.StringFlag{
						Name:  "config",
						Usage: "Config of rhilex",
						Value: "rhilex.ini",
					},
				},
				Action: func(c *cli.Context) error {
					utils.CLog(typex.Banner)
					utils.ShowGGpuAndCpuInfo()
					engine.RunRhilex(c.String("config"))
					fmt.Printf("[RHILEX UPGRADE] Run rhilex successfully.")
					return nil
				},
			},
			{
				Name:   "upgrade",
				Hidden: true,
				Usage:  "! JUST FOR Upgrade FirmWare",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "upgrade",
						Usage: "! THIS PARAMENT IS JUST FOR Upgrade FirmWare",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					file, err := os.Create(ossupport.UpgradeLogPath)
					if err != nil {
						utils.CLog(err.Error())
						return nil
					}
					defer file.Close()
					os.Stdout = file
					os.Stderr = file
					// upgrade lock
					if err := os.WriteFile(ossupport.UpgradeLockPath, []byte{48}, 0755); err != nil {
						utils.CLog("[RHILEX UPGRADE] Write Upgrade Lock File error:%s", err.Error())
						return nil
					}
					defer func() {
						// upgrade lock
						if err := os.Remove(ossupport.UpgradeLockPath); err != nil {
							utils.CLog("[RHILEX UPGRADE] Remove Upgrade Lock File error:%s", err.Error())
							return
						}
						utils.CLog("[RHILEX UPGRADE] Remove Upgrade Lock File Finished")
					}()
					if runtime.GOOS != "linux" {
						utils.CLog("[RHILEX UPGRADE] Only Support Linux")
						return nil
					}
					if !c.Bool("upgrade") {
						utils.CLog("[RHILEX UPGRADE] Nothing todo")
						return nil
					}
					// unzip Firmware
					utils.CLog("[RHILEX UPGRADE] Unzip Firmware")
					if err := ossupport.UnzipFirmware(
						ossupport.FirmwarePath, ossupport.MainWorkDir); err != nil {
						utils.CLog("[RHILEX UPGRADE] Unzip error:%s", err.Error())
						return nil
					}
					utils.CLog("[RHILEX UPGRADE] Unzip Firmware finished")
					// Remove old package
					utils.CLog("[RHILEX UPGRADE] Remove Firmware")
					if err := os.Remove(ossupport.FirmwarePath); err != nil {
						utils.CLog("[RHILEX UPGRADE] Remove Firmware error:%s", err.Error())
						return nil
					}
					utils.CLog("[RHILEX UPGRADE] Remove Firmware finished")
					//
					utils.CLog("[RHILEX UPGRADE] Restart rhilex")
					if err := ossupport.RestartRhilex(); err != nil {
						utils.CLog("[RHILEX UPGRADE] Restart rhilex error:%s", err.Error())
						return nil
					}
					utils.CLog("[RHILEX UPGRADE] Restart rhilex finished, Upgrade Process Exited")
					os.Exit(0)
					return nil
				},
			},
			// 数据恢复
			{
				Name:   "recover",
				Usage:  "! JUST FOR Recover Data",
				Hidden: true,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "recover",
						Usage: "! THIS PARAMENT IS JUST FOR Recover Data",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					file, err := os.Create(ossupport.RecoverLogPath)
					if err != nil {
						utils.CLog(err.Error())
						return nil
					}
					defer file.Close()
					os.Stdout = file
					os.Stderr = file
					// upgrade lock
					if err := os.WriteFile(ossupport.UpgradeLockPath, []byte{48}, 0755); err != nil {
						utils.CLog("[DATA RECOVER] Write Recover Lock File error:%s", err.Error())
						return nil
					}
					defer func() {
						// upgrade lock
						if err := os.Remove(ossupport.UpgradeLockPath); err != nil {
							utils.CLog("[DATA RECOVER] Remove Recover Lock File error:%s", err.Error())
							return
						}
						utils.CLog("[DATA RECOVER] Remove Recover Lock File Finished")
					}()
					if runtime.GOOS != "linux" {
						utils.CLog("[DATA RECOVER] Only Support Linux")
						return nil
					}

					if !c.Bool("recover") {
						utils.CLog("[DATA RECOVER] Nothing todo")
						return nil
					}
					utils.CLog("[DATA RECOVER] Remove Old Db File")
					if err := os.Remove(ossupport.RunDbPath); err != nil {
						utils.CLog("[DATA RECOVER] Remove Old Db File error:%s", err.Error())
						return nil
					}
					utils.CLog("[DATA RECOVER] Remove Old Db File Finished")
					utils.CLog("[DATA RECOVER] Move New Db File")
					if err := ossupport.MoveFile(ossupport.RecoveryDbPath,
						ossupport.RunDbPath); err != nil {
						utils.CLog("[DATA RECOVER] Move New Db File error:%s", err.Error())
						return nil
					}
					utils.CLog("[DATA RECOVER] Move New Db File Finished")
					utils.CLog("[DATA RECOVER] Try to Restart rhilex")
					if err := ossupport.RestartRhilex(); err != nil {
						utils.CLog("[DATA RECOVER] Restart rhilex error:%s", err.Error())
					} else {
						utils.CLog("[DATA RECOVER] Restart rhilex success, Recover Process Exited")
					}
					os.Exit(0)
					return nil
				},
			},
			{
				Name:   "active",
				Usage:  "active -H host -U rhino -P hoot",
				Hidden: true,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "H",
						Usage: "active server ip",
					},
					&cli.StringFlag{
						Name:  "U",
						Usage: "active admin username",
					},
					&cli.StringFlag{
						Name:  "P",
						Usage: "active admin password",
					},
				},

				Action: func(c *cli.Context) error {
					host := c.String("H")
					if host == "" {
						return fmt.Errorf("[LICENCE ACTIVE]: missing 'host' parameter")
					}
					username := c.String("U")
					if username == "" {
						return fmt.Errorf("[LICENCE ACTIVE]: missing admin 'username' parameter")
					}
					password := c.String("P")
					if password == "" {
						return fmt.Errorf("[LICENCE ACTIVE]: missing admin 'password' parameter")
					}
					// linux
					if runtime.GOOS == "linux" {
						macAddr, err := ossupport.GetLinuxMacAddr("eth0")
						if err != nil {
							return fmt.Errorf("[LICENCE ACTIVE]: Get Local Mac Address error: %s", err)
						}
						// Commercial version will implement it
						// rhilex active -H https://127.0.0.1/api/v1/device-active -U admin -P 123456
						// - H: Active Server Host
						// - U: Active Server Account
						// - P: Active Server Password
						err1 := utils.FetchLoadLicense(host, username, password, macAddr)
						if err1 != nil {
							return fmt.Errorf("[LICENCE ACTIVE]: Fetch license failed, error: %s", err1)
						}
						return nil
					}
					if runtime.GOOS == "windows" {
						// Just for test
						macAddr, err0 := ossupport.GetWindowsMACAddress()
						if err0 != nil {
							return fmt.Errorf("[LICENCE ACTIVE]: Get Local Mac Address error: %s", err0)
						}
						err1 := utils.FetchLoadLicense(host, username, password, macAddr)
						if err1 != nil {
							return fmt.Errorf("[LICENCE ACTIVE]: Fetch license failed, error: %s", err1)
						}
						return nil
					}
					return fmt.Errorf("[LICENCE ACTIVE]: Active not supported on current distribution.")
				},
			},
			// version
			{
				Name:  "version",
				Usage: "Show rhilex Current Version",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "version",
						Usage: "rhilex version",
					},
				},
				Action: func(*cli.Context) error {
					version := fmt.Sprintf("[%v-%v-%v]",
						runtime.GOOS, runtime.GOARCH, typex.MainVersion)
					utils.CLog("[*] rhilex Version: " + version)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
