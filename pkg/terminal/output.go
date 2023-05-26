package terminal

import (
	"fmt"
)

func Output(msg string) {
	fmt.Print(msg)
}

func Read() interface{} {
	var res interface{}
	fmt.Scanln(&res)
	return res
}
