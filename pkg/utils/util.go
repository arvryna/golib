package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func GetUserInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	data, _ := reader.ReadString('\n')
	return strings.TrimSpace(data)
}
