package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

func promptYesNo(msg string) (bool, error) {
	fmt.Printf("%s [yes/no]\n", msg)
	r := bufio.NewReader(os.Stdin)
	response, err := r.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return false, context.Canceled
		}
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

func promptEnter(msg string) (bool, error) {
	fmt.Printf("%s [press enter to continue or CTRL+C to exit]\n", msg)
	r := bufio.NewReader(os.Stdin)
	_, err := r.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
