package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Util to easily get input from user and trim input
func GetUserInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	data, _ := reader.ReadString('\n')
	return strings.TrimSpace(data)
}
