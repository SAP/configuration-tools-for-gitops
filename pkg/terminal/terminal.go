package terminal

import (
	"fmt"
)

func Output(msg string) {
	fmt.Print(msg)
}

func ReadInt() (int, error) {
	var res int
	_, err := fmt.Scanln(&res)
	return res, err
}

func ReadStr() (string, error) {
	var res string
	_, err := fmt.Scanln(&res)
	return res, err
}
