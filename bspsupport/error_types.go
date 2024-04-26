package archsupport

import (
	"errors"
	"runtime"
)

var errArchNotSupport = errors.New("not support current OS:" + runtime.GOOS)
var errInvalidLen = errors.New("invalid len")
var errInvalidValue = errors.New("invalid value")
