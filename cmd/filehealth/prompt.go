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
	// Bracket the prompt in a horizontal rule
	fmt.Printf("----\n")
	defer fmt.Printf("----\n")

	// Prompt the user
	fmt.Printf("%s [yes/no]\n", msg)

	// Collect a response
	r := bufio.NewReader(os.Stdin)
	response, err := r.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			// The reader closed, probably due to CTRL+C
			return false, context.Canceled
		}
		return false, err
	}

	// Validate the response
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
	// Bracket the prompt in a horizontal rule
	fmt.Printf("----\n")
	defer fmt.Printf("----\n")

	// Collect a response
	fmt.Printf("%s [press enter to continue or CTRL+C to exit]\n", msg)
	r := bufio.NewReader(os.Stdin)
	_, err := r.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			// The reader closed, probably due to CTRL+C
			return false, nil
		}
		return false, err
	}
	return true, nil
}
