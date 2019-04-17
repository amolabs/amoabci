package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func PromptUsername() (string, error) {
	fmt.Printf("\nInput username of the signing key: ")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.Trim(username, "\r\n"), nil
}

func PromptPassphrase() (string, error) {
	fmt.Printf("Type passphrase: ")
	b, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(b), nil
}
