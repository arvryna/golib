package utils

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func GetUserInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	data, _ := reader.ReadString('\n')
	return strings.TrimSpace(data)
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func Date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func Wait() {
	var sig = make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	<-sig
}
