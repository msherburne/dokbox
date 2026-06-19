package config

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func ConfirmSave(reader io.Reader, writer io.Writer, prompt string) (bool, error) {
	if writer == nil {
		writer = io.Discard
	}
	if reader == nil {
		reader = strings.NewReader("")
	}

	if _, err := fmt.Fprintf(writer, "%s [Y/n] ", prompt); err != nil {
		return false, err
	}

	response, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}

	answer := strings.TrimSpace(strings.ToLower(response))
	return answer == "" || answer == "y" || answer == "yes", nil
}
