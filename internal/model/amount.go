package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)


type Amount float64

var amountNumberRe = regexp.MustCompile(`-?\d[\d\s\x{00A0}.,]*`)

func (a *Amount) UnmarshalJSON(data []byte) error {
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*a = Amount(f)
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		if string(data) == "null" {
			*a = 0
			return nil
		}
		return fmt.Errorf("amount: unsupported value %s", data)
	}

	f, err := ParseAmount(s)
	if err != nil {
		return err
	}
	*a = Amount(f)
	return nil
}


func ParseAmount(s string) (float64, error) {
	if strings.TrimSpace(s) == "" {
		return 0, nil
	}
	num := amountNumberRe.FindString(s)
	if num == "" {
		return 0, fmt.Errorf("amount: no number in %q", s)
	}

	num = strings.NewReplacer(" ", "", " ", "", "\t", "", "\n", "").Replace(num)
	num = strings.TrimRight(num, ".,")

	lastComma := strings.LastIndex(num, ",")
	lastDot := strings.LastIndex(num, ".")
	switch {
	case lastComma >= 0 && lastDot >= 0:

		if lastComma > lastDot {
			num = strings.ReplaceAll(num, ".", "")
			num = strings.Replace(num, ",", ".", 1)
		} else {
			num = strings.ReplaceAll(num, ",", "")
		}
	case lastComma >= 0:
		if strings.Count(num, ",") == 1 && len(num)-lastComma-1 <= 2 {
			num = strings.Replace(num, ",", ".", 1)
		} else {
			num = strings.ReplaceAll(num, ",", "")
		}
	case lastDot >= 0:
		if strings.Count(num, ".") > 1 {
			num = strings.ReplaceAll(num, ".", "")
		}
	}

	f, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, fmt.Errorf("amount: parse %q: %w", s, err)
	}
	return f, nil
}
