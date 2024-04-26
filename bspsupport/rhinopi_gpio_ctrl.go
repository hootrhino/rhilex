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

package archsupport

//-----------------------------------------------
// 这是RHILEXG1网关的DI-DO支持库
//-----------------------------------------------
/*
    pins map

	DO1 -> PA6
	DO2 -> PA7
	DI1 -> PA8
	DI2	-> PA9
	DI3 -> PA10
	USER_GPIO -> PA20
*/
import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

//-----------------------------------------------
// 这是RHILEXG1网关的DI-DO支持库
//-----------------------------------------------
/*
    pins map

	DO1 -> PA6
	DO2 -> PA7
	DI1 -> PA8
	DI2	-> PA9
	DI3 -> PA10
	USER_GPIO -> PA20
*/
const (
	// DO
	rhinopi_DO1 string = "6"
	rhinopi_DO2 string = "7"
	// DI
	rhinopi_DI1 string = "8"
	rhinopi_DI2 string = "9"
	rhinopi_DI3 string = "10"
	// Use LED
	rhinopi_USER_GPIO string = "20"
)

const (
	rhinopi_Out string = "out"
	rhinopi_In  string = "in"
)

func init() {
	env := os.Getenv("ARCHSUPPORT")
	if env == "RHILEXG1" {
		_RHILEXG1_GPIOAllInit()
	}
}

/*
explain:init all gpio
*/
func _RHILEXG1_GPIOAllInit() int {
	gpio6 := "/sys/class/gpio/gpio6/value"
	gpio7 := "/sys/class/gpio/gpio7/value"
	gpio8 := "/sys/class/gpio/gpio8/value"
	gpio9 := "/sys/class/gpio/gpio9/value"
	gpio10 := "/sys/class/gpio/gpio10/value"
	gpio20 := "/sys/class/gpio/gpio20/value"
	_, err1 := os.Stat(gpio6)
	_, err2 := os.Stat(gpio7)
	_, err3 := os.Stat(gpio8)
	_, err4 := os.Stat(gpio9)
	_, err5 := os.Stat(gpio10)
	_, err6 := os.Stat(gpio20)
	if err1 != nil {
		if strings.Contains(err1.Error(), "no such file or directory") {
			_RHILEXG1_GPIOInit(rhinopi_DO1, rhinopi_Out)
			fmt.Println("RHILEXG1_GPIOAllInit DO1 Out Mode Ok")
		}
	}
	if err2 != nil {
		if strings.Contains(err2.Error(), "no such file or directory") {
			_RHILEXG1_GPIOInit(rhinopi_DO2, rhinopi_Out)
			fmt.Println("RHILEXG1_GPIOAllInit DO2 Out Mode Ok")
		}
	}
	if err3 != nil {
		if strings.Contains(err3.Error(), "no such file or directory") {
			_RHILEXG1_GPIOInit(rhinopi_DI1, rhinopi_In)
			fmt.Println("RHILEXG1_GPIOAllInit DI1 In Mode Ok")
		}
	}
	if err4 != nil {
		if strings.Contains(err4.Error(), "no such file or directory") {
			_RHILEXG1_GPIOInit(rhinopi_DI2, rhinopi_In)
			fmt.Println("RHILEXG1_GPIOAllInit DI2 In Mode Ok")
		}
	}
	if err5 != nil {
		if strings.Contains(err5.Error(), "no such file or directory") {
			_RHILEXG1_GPIOInit(rhinopi_DI3, rhinopi_In)
			fmt.Println("RHILEXG1_GPIOAllInit DI3 In Mode Ok")
		}
	}
	if err6 != nil {
		if strings.Contains(err5.Error(), "no such file or directory") {
			_RHILEXG1_GPIOInit(rhinopi_USER_GPIO, rhinopi_Out)
			fmt.Println("RHILEXG1_GPIOAllInit USER_GPIO Out Mode Ok")
		}
	}
	// 返回值无用
	return 1
}

/*
explain:init gpio
Pin: gpio pin
direction:gpio direction in or out
*/
func _RHILEXG1_GPIOInit(Pin string, direction string) {
	//gpio export
	cmd := fmt.Sprintf("echo %s > /sys/class/gpio/export", Pin)
	output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Println("[RHILEXG1_GPIOInit] error", err, string(output))
		return
	}
	//gpio set direction
	cmd = fmt.Sprintf("echo %s > /sys/class/gpio/gpio%s/direction", direction, Pin)
	output, err = exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Println("[RHILEXG1_GPIOInit] error", err, string(output))
	}
}

/*
*
* GPIO 表
*
 */
const (
	__h3_GPIO_PATH = "/sys/class/gpio/gpio%v/value"
)

/*
*
* 新版本的文件读取形式获取GPIO状态
*
 */
func RHILEXG1_GPIOGetDO1() (int, error) {
	return RHILEXG1_GPIOGetByFile(6)
}
func RHILEXG1_GPIOGetDO2() (int, error) {
	return RHILEXG1_GPIOGetByFile(7)
}
func RHILEXG1_GPIOGetDI1() (int, error) {
	return RHILEXG1_GPIOGetByFile(8)
}
func RHILEXG1_GPIOGetDI2() (int, error) {
	return RHILEXG1_GPIOGetByFile(9)
}
func RHILEXG1_GPIOGetDI3() (int, error) {
	return RHILEXG1_GPIOGetByFile(10)
}
func RHILEXG1_GPIOGetUserGpio() (int, error) {
	return RHILEXG1_GPIOGetByFile(20)
}
func RHILEXG1_GPIOGetByFile(pin byte) (int, error) {
	return __GPIOGet(fmt.Sprintf(__h3_GPIO_PATH, pin))
}

func __GPIOGet(gpioPath string) (int, error) {
	bites, err := os.ReadFile(gpioPath)
	if err != nil {
		return 0, err
	}
	if len(bites) > 0 {
		if bites[0] == '0' || bites[0] == 48 {
			return 0, nil
		}
		if bites[1] == '1' || bites[0] == 49 {
			return 1, nil
		}
	}
	return 0, fmt.Errorf("read gpio value failed: %s, value: %v", gpioPath, bites)
}

// Set

func RHILEXG1_GPIOSetDO1(value int) error {
	return RHILEXG1_GPIOSetByFile(6, value)
}
func RHILEXG1_GPIOSetDO2(value int) error {
	return RHILEXG1_GPIOSetByFile(7, value)
}
func RHILEXG1_GPIOSetDI1(value int) error {
	return RHILEXG1_GPIOSetByFile(8, value)
}
func RHILEXG1_GPIOSetDI2(value int) error {
	return RHILEXG1_GPIOSetByFile(9, value)
}
func RHILEXG1_GPIOSetDI3(value int) error {
	return RHILEXG1_GPIOSetByFile(10, value)
}
func RHILEXG1_GPIOSetUserGpio(value int) error {
	return RHILEXG1_GPIOSetByFile(20, value)
}

func RHILEXG1_GPIOSetByFile(pin, value int) error {
	return __GPIOSet(fmt.Sprintf(__h3_GPIO_PATH, pin), value)
}

func __GPIOSet(gpioPath string, value int) error {
	if value == 1 {
		err := os.WriteFile(gpioPath, []byte{'1'}, 0644)
		if err != nil {
			return err
		}
	}
	if value == 0 {
		err := os.WriteFile(gpioPath, []byte{'0'}, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
