package gpio

import (
	"fmt"
	"os"
	"strconv"
)

type GPIO struct {
	PinNumber int
	Exported  bool
	Direction string
	Value     int
}

func (g *GPIO) Export() error {
	if g.Exported {
		return nil
	}

	err := os.WriteFile("/sys/class/gpio/export", []byte(strconv.Itoa(g.PinNumber)), 0644)
	if err != nil {
		return err
	}

	g.Exported = true
	return nil
}

func (g *GPIO) Unexport() error {
	if !g.Exported {
		return nil
	}

	err := os.WriteFile("/sys/class/gpio/unexport", []byte(strconv.Itoa(g.PinNumber)), 0644)
	if err != nil {
		return err
	}

	g.Exported = false
	return nil
}

func (g *GPIO) SetDirection(direction string) error {
	if !g.Exported {
		return fmt.Errorf("GPIO %d is not exported", g.PinNumber)
	}

	if direction != "in" && direction != "out" {
		return fmt.Errorf("invalid direction: %s", direction)
	}

	err := os.WriteFile(fmt.Sprintf("/sys/class/gpio/gpio%d/direction", g.PinNumber), []byte(direction), 0644)
	if err != nil {
		return err
	}

	g.Direction = direction
	return nil
}

func (g *GPIO) SetValue(value int) error {
	if !g.Exported {
		return fmt.Errorf("GPIO %d is not exported", g.PinNumber)
	}

	if g.Direction != "out" {
		return fmt.Errorf("GPIO %d is not set to output mode", g.PinNumber)
	}

	if value != 0 && value != 1 {
		return fmt.Errorf("invalid value: %d", value)
	}

	err := os.WriteFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", g.PinNumber), []byte(strconv.Itoa(value)), 0644)
	if err != nil {
		return err
	}

	g.Value = value
	return nil
}

func (g *GPIO) GetValue() (int, error) {
	if !g.Exported {
		return 0, fmt.Errorf("GPIO %d is not exported", g.PinNumber)
	}

	data, err := os.ReadFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", g.PinNumber))
	if err != nil {
		return 0, err
	}

	value, err := strconv.Atoi(string(data[:len(data)-1])) // 去除换行符
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (g *GPIO) GetInfo() string {
	return fmt.Sprintf("GPIO %d: Exported=%t, Direction=%s, Value=%d", g.PinNumber, g.Exported, g.Direction, g.Value)
}

func NewGPIO(pinNumber int) *GPIO {
	return &GPIO{
		PinNumber: pinNumber,
		Exported:  false,
		Direction: "",
		Value:     0,
	}
}
