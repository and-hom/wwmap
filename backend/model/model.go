package model

import (
	"fmt"
	"strings"
	"regexp"
	"strconv"
)

type SportCategory struct {
	Category int
	Sub      string
}

func (this SportCategory) MarshalJSON() ([]byte, error) {
	if this.Category == 0 {
		return []byte("\"\""), nil
	}
	return []byte(fmt.Sprintf("\"%d%s\"", this.Category, this.Sub)), nil
}

func (this *SportCategory) UnmarshalJSON(data []byte) error {
	dataStr := string(data)
	if len(strings.TrimSpace(dataStr)) == 0 {
		// no category specified
		return nil
	}

	re := regexp.MustCompile("^(\\d+)([A-Za-z]+)?$")
	var err error

	match := re.FindStringSubmatch(dataStr)
	if match == nil {
		return fmt.Errorf("Can not parse route category: %s", dataStr)
	}
	this.Category, err = strconv.Atoi(match[1])
	if err != nil {
		return err
	}
	if len(match) >= 3 {
		this.Sub = match[2]
	} else {
		this.Sub = ""
	}
	return nil
}
