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

package utils

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

/*
#include <sys/utsname.h>
#include <stdio.h>

struct utsname ReleaseInfo() {
    struct utsname utsname1;
    int rv = uname(&utsname1);
    if (rv == -1) {
        return utsname1;
    }
    return utsname1;
}
*/
import "C"

/*
*
* Cgo 实现, 用来获取Linux的系统参数
*
 */
type Utsname struct {
	sysname  string
	nodename string
	release  string
	version  string
	machine  string
}

func ReleaseInfo() Utsname {
	CS := C.ReleaseInfo()
	sysname := [65]C.char(CS.sysname)
	nodename := [65]C.char(CS.nodename)
	release := [65]C.char(CS.release)
	version := [65]C.char(CS.version)
	machine := [65]C.char(CS.machine)

	return Utsname{
		sysname:  C.GoStringN(&sysname[0], 65),
		nodename: C.GoStringN(&nodename[0], 65),
		release:  C.GoStringN(&release[0], 65),
		version:  C.GoStringN(&version[0], 65),
		machine:  C.GoStringN(&machine[0], 65),
	}
}

/*
*
* Get Local IP
*
 */
func HostNameI() ([]string, error) {
	dist, _ := GetOSDistribution()
	if dist == "openwrt" {
		line := `ip addr show | awk '/inet / {print $2}' | awk 'BEGIN{FS="/"} {split($0, arr, "/"); print arr[1]}'`
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "sh", "-c", line)
		output, err := cmd.Output()
		if err != nil {
			return []string{}, err
		}
		result := strings.TrimSpace(string(output))
		ips := []string{}
		for _, v := range strings.Split(result, "\n") {
			if v != "127.0.0.1" {
				ips = append(ips, v)
			}
		}
		return ips, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "hostname", "-I")
	data, err1 := cmd.Output()
	if err1 != nil {
		return []string{}, err1
	}
	ss := []string{}
	for _, s := range strings.Split(string(data), " ") {
		if s != "\n" {
			ss = append(ss, s)
		}
	}
	return ss, nil
}

/*
*
* 获取设备树
*
 */
func GetSystemDevices() (SystemDevices, error) {
	SystemDevices := SystemDevices{
		Uarts:  []string{},
		Videos: []string{},
		Audios: []string{},
	}
	f, err := os.Open("/dev/")
	if err != nil {
		return SystemDevices, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return SystemDevices, err
	}

	for _, d := range list {
		if !d.IsDir() {
			if strings.Contains(d.Name(), "ttyS") {
				SystemDevices.Uarts = append(SystemDevices.Uarts, d.Name())
			}
			if strings.Contains(d.Name(), "ttyACM") {
				SystemDevices.Uarts = append(SystemDevices.Uarts, d.Name())
			}
			if strings.Contains(d.Name(), "ttyUSB") {
				SystemDevices.Uarts = append(SystemDevices.Uarts, d.Name())
			}
			if strings.Contains(d.Name(), "video") {
				SystemDevices.Videos = append(SystemDevices.Videos, d.Name())
			}
			if strings.Contains(d.Name(), "audio") {
				SystemDevices.Audios = append(SystemDevices.Audios, d.Name())
			}
		}
	}
	return SystemDevices, nil
}

/*
*
* Linux: cat /etc/os-release
*
 */
func CatOsRelease() (map[string]string, error) {
	returnMap := map[string]string{}
	if runtime.GOOS == "linux" {
		if PathExists("/etc/os-release") {
			cfg, err := ini.ShadowLoad("/etc/os-release")
			if err != nil {
				return nil, err
			}
			DefaultSection, err := cfg.GetSection("DEFAULT")
			if err != nil {
				return nil, err
			}
			for _, Key := range DefaultSection.KeyStrings() {
				V, _ := DefaultSection.GetKey(Key)
				returnMap[Key] = V.String()
			}
		} else {
			returnMap["OS Version"] = "UNKNOWN"
		}
	}
	return returnMap, nil

}
