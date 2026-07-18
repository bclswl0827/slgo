package handlers

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
)

type TIME struct{}

// Callback of "TIME <...>" command, implements handler interface
func (t *TIME) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	if len(args) < 1 || len(args) > 2 {
		_, err := client.Write([]byte(RES_ERR))
		return err
	}

	startTime, err := parseSeedLinkTime(args[0])
	if err != nil || startTime.After(provider.GetCurrentTime()) {
		_, writeErr := client.Write([]byte(RES_ERR))
		return writeErr
	}
	endTime := provider.GetCurrentTime()
	if len(args) == 2 {
		endTime, err = parseSeedLinkTime(args[1])
		if err != nil || endTime.Before(startTime) {
			_, writeErr := client.Write([]byte(RES_ERR))
			return writeErr
		}
	}

	client.StartTime = startTime
	client.EndTime = endTime
	_, err = client.Write([]byte(RES_OK))
	return err
}

// Fallback of "TIME <...>" command, implements handler interface
func (*TIME) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}

func (*TIME) getTimeFromArg(timeStr string) (time.Time, error) {
	return parseSeedLinkTime(timeStr)
}

func parseSeedLinkTime(timeStr string) (time.Time, error) {
	splitTimeStr := strings.Split(timeStr, ",")
	if len(splitTimeStr) != 6 {
		return time.Time{}, errors.New("time string must have 6 comma-separated values")
	}

	// Format:  YYYY,MM,DD,hh,mm,ss
	// Example: 2024,01,16,07,15,16
	year, err := strconv.Atoi(splitTimeStr[0])
	if err != nil {
		return time.Time{}, err
	}

	monthInt, err := strconv.Atoi(splitTimeStr[1])
	if err != nil {
		return time.Time{}, err
	}

	month := time.Month(monthInt)
	day, err := strconv.Atoi(splitTimeStr[2])
	if err != nil {
		return time.Time{}, err
	}

	hour, err := strconv.Atoi(splitTimeStr[3])
	if err != nil {
		return time.Time{}, err
	}

	minute, err := strconv.Atoi(splitTimeStr[4])
	if err != nil {
		return time.Time{}, err
	}

	secondValue, err := strconv.ParseFloat(splitTimeStr[5], 64)
	if err != nil {
		return time.Time{}, err
	}
	if secondValue < 0 || secondValue >= 60 {
		return time.Time{}, errors.New("seconds must be in the range [0, 60)")
	}
	second, fractional := math.Modf(secondValue)
	nanosecond := int(math.Round(fractional * float64(time.Second)))
	if nanosecond == int(time.Second) {
		second++
		nanosecond = 0
	}

	parsed := time.Date(year, month, day, hour, minute, int(second), nanosecond, time.UTC)
	if parsed.Year() != year || parsed.Month() != month || parsed.Day() != day ||
		parsed.Hour() != hour || parsed.Minute() != minute || parsed.Second() != int(second) {
		return time.Time{}, errors.New("time contains an out-of-range value")
	}
	return parsed, nil
}
