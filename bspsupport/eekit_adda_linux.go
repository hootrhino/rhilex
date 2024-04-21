package archsupport

/*
*
* Linux 特定实现
*
 */

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

//-----------------------------------------------
// 这是EEKIT网关的DI-DO支持库
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
	if env == "RHINOPI" {
		_RHINOPI_GPIOAllInit()
	}
}

/*
explain:init all gpio
*/
func _RHINOPI_GPIOAllInit() int {
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
			_RHINOPI_GPIOInit(rhinopi_DO1, rhinopi_Out)
			fmt.Println("EEKIT_GPIOAllInit DO1 Out Mode Ok")
		}
	}
	if err2 != nil {
		if strings.Contains(err2.Error(), "no such file or directory") {
			_RHINOPI_GPIOInit(rhinopi_DO2, rhinopi_Out)
			fmt.Println("EEKIT_GPIOAllInit DO2 Out Mode Ok")
		}
	}
	if err3 != nil {
		if strings.Contains(err3.Error(), "no such file or directory") {
			_RHINOPI_GPIOInit(rhinopi_DI1, rhinopi_In)
			fmt.Println("EEKIT_GPIOAllInit DI1 In Mode Ok")
		}
	}
	if err4 != nil {
		if strings.Contains(err4.Error(), "no such file or directory") {
			_RHINOPI_GPIOInit(rhinopi_DI2, rhinopi_In)
			fmt.Println("EEKIT_GPIOAllInit DI2 In Mode Ok")
		}
	}
	if err5 != nil {
		if strings.Contains(err5.Error(), "no such file or directory") {
			_RHINOPI_GPIOInit(rhinopi_DI3, rhinopi_In)
			fmt.Println("EEKIT_GPIOAllInit DI3 In Mode Ok")
		}
	}
	if err6 != nil {
		if strings.Contains(err5.Error(), "no such file or directory") {
			_RHINOPI_GPIOInit(rhinopi_USER_GPIO, rhinopi_Out)
			fmt.Println("EEKIT_GPIOAllInit USER_GPIO Out Mode Ok")
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
func _RHINOPI_GPIOInit(Pin string, direction string) {
	//gpio export
	cmd := fmt.Sprintf("echo %s > /sys/class/gpio/export", Pin)
	output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Println("[EEKIT_GPIOInit] error", err, string(output))
		return
	}
	//gpio set direction
	cmd = fmt.Sprintf("echo %s > /sys/class/gpio/gpio%s/direction", direction, Pin)
	output, err = exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Println("[EEKIT_GPIOInit] error", err, string(output))
	}
}

/*
explain:set gpio
Pin: gpio pin
Value:gpio level 1 is high 0 is low
*/
func EEKIT_GPIOSet(pin, value int) (bool, error) {
	cmd := fmt.Sprintf("echo %d > /sys/class/gpio/gpio%d/value", value, pin)
	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		log.Println("[EEKIT_GPIOSet] error", err, string(output))
		return false, err
	}
	v, e := EEKIT_GPIOGet(pin)
	if e != nil {
		return false, e
	}
	return v == value, nil
}

/*
explain:read gpio
Pin: gpio pin
return:1 is high 0 is low
*/
func EEKIT_GPIOGet(pin int) (int, error) {
	cmd := fmt.Sprintf("cat /sys/class/gpio/gpio%d/value", pin)
	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return -1, err
	}
	if len(output) < 1 {
		return -1, errInvalidLen
	}
	if output[0] == '0' {
		return 0, nil
	}
	if output[0] == '1' {
		return 1, nil
	}
	return -1, errInvalidValue
}
