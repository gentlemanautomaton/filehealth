package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func prompt(msg string) (bool, error) {
	fmt.Printf("%s [yes/no]\n", msg)
	r := bufio.NewReader(os.Stdin)
	response, err := r.ReadString('\n')
	if err != nil {
		return false, err
	}
	response = strings.ToLower(strings.TrimSpace(response))
	switch response {
	case "yes":
		return true, nil
	case "no", "n":
		return false, nil
	default:
		return false, fmt.Errorf("\"%s\" is not \"yes\" or \"no\"", response)
	}
}
