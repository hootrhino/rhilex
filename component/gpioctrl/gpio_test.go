package gpio

import (
	"fmt"
	"testing"
	"time"
)

func Test_Gpio(t *testing.T) {
	// 创建一个 GPIO 实例，例如 GPIO 17
	gpio17 := NewGPIO(17)

	// 导出 GPIO 引脚
	err := gpio17.Export()
	if err != nil {
		fmt.Println("Error exporting GPIO:", err)
		return
	}
	defer gpio17.Unexport() // 在程序结束时取消导出

	// 设置 GPIO 引脚为输出模式
	err = gpio17.SetDirection("out")
	if err != nil {
		fmt.Println("Error setting direction:", err)
		return
	}

	// 设置 GPIO 引脚为高电平
	err = gpio17.SetValue(1)
	if err != nil {
		fmt.Println("Error setting value:", err)
		return
	}

	fmt.Println(gpio17.GetInfo())

	time.Sleep(2 * time.Second)

	// 设置 GPIO 引脚为低电平
	err = gpio17.SetValue(0)
	if err != nil {
		fmt.Println("Error setting value:", err)
		return
	}

	fmt.Println(gpio17.GetInfo())

	// 设置为输入模式
	err = gpio17.SetDirection("in")
	if err != nil {
		fmt.Println("Error setting direction:", err)
		return
	}

	value, err := gpio17.GetValue()
	if err != nil {
		fmt.Println("Error reading value:", err)
		return
	}

	fmt.Println("GPIO Value:", value)
}
