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

var (
	AffirmationOptions = []string{"y", "yes"}
)

func IsYes() (bool, error) {
	var res string
	_, err := fmt.Scanln(&res)
	if err != nil {
		return false, err
	}
	for _, o := range AffirmationOptions {
		if res == o {
			return true, nil
		}
	}
	return false, nil
}
