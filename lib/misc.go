package lib

import (
	"fmt"
	"strconv"
	"strings"
)

// string could be  -10.54-36.42
func GetFloatTemperature(s string) float64 {
	var index int = 0
	var result float64 = 0
	var err error
	if strings.HasPrefix(s, "-") {
		index = 1
	}
	if result, err = strconv.ParseFloat(strings.Split(s, "-")[index], 64); err == nil {
		fmt.Printf("TEST %v\n", result)
		if index == 1 {
			result *= -1
		}
	}
	return result
}
