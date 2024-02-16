package poh

import "fmt"

var ErrInUse = fmt.Errorf("conection in use")
var ErrConnTerminate = fmt.Errorf("conection terminate fail")

var ErrInternalLockCP = fmt.Errorf("conection pool internal lock error")
var ErrOverflowCP = fmt.Errorf("conection pool overflow")
var ErrConnectionCreationErrorCP = fmt.Errorf("conection create fail")
