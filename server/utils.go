package server

import (
	"strconv"
	"strings"
)

func modifyTime(timeStr string) int {
	timeList := strings.Split(timeStr, ":")
	hour := timeList[0]
	minute := timeList[1]
	second := timeList[2]
	hourInt, err := strconv.Atoi(hour)
	minuteInt, err := strconv.Atoi(minute)
	secondInt, err := strconv.Atoi(second)
	if err != nil {
		return 0
	}
	return hourInt*3600 + minuteInt*60 + secondInt
}
