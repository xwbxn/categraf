package w_aviation

import (
	"regexp"
	"strconv"
	"strings"
)

//	ID  | SDRType | Type            |SNum| Name             |Status| Reading
//
// 0001 | Full    | Voltage         | 01 | BMC +0.75V       | OK   | 0.74 V
// 0002 | Full    | Voltage         | 02 | BMC +1V          | OK   | 1.01 V
// 0003 | Full    | Voltage         | 03 | BMC +1.5V        | OK   | 1.52 V
// 0004 | Full    | Voltage         | 04 | BMC +1.8V        | OK   | 1.82 V
// extractFieldsFromRegex consumes a regex with named capture groups and returns a kvp map of strings with the results
var re = regexp.MustCompile(`^\s*(?P<id>[^I|^D]*?)\s*\|\s*(?P<sdrtype>.*?)\s*\|\s*(?P<sdr>.*?)\s*\|\s*(?P<snum>.*?)\s*\|\s*(?P<name>.*?)\s*\|\s*(?P<status>.*?)\s*\|\s*(?P<reading>.*?)\s*$`)

func ExtractFieldsFromRegex(input string) map[string]string {
	submatches := re.FindStringSubmatch(input)
	results := make(map[string]string)
	subexpNames := re.SubexpNames()
	if len(subexpNames) > len(submatches) {
		return results
	}
	for i, name := range subexpNames {
		if name != input && name != "" && input != "" {
			results[name] = strings.Trim(submatches[i], "")
		}
	}
	return results
}

func transformReading(val string) float64 {
	arr := strings.Split(val, " ")
	v, err := strconv.ParseFloat(arr[0], 32)
	if err != nil {
		return float64(0)
	}
	return v
}
