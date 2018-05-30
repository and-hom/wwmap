package model

import (
	"fmt"
	"strings"
	"regexp"
	"strconv"
)

const UNDEFINED_CATEGORY = 0;

const IMPASSABLE = -1

type SportCategory struct {
	Category int
	Sub      string
}

func (this SportCategory) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", this.String())), nil
}

func (this SportCategory) String() string {
	return this.Serialize();
}

func (this SportCategory) Serialize() string {
	if this.Category == UNDEFINED_CATEGORY {
		return "0"
	}
	return fmt.Sprintf("%d%s", this.Category, this.Sub)
}

func (this *SportCategory) UnmarshalJSON(data []byte) error {
	dataStr := string(data)
	if len(strings.TrimSpace(dataStr)) == 0 || dataStr == "\"\"" {
		// no category specified
		return nil
	}

	re := regexp.MustCompile("^\"?(-?\\d+)([A-Za-z]+)?\"?$")
	var err error

	match := re.FindStringSubmatch(dataStr)
	if match == nil {
		return fmt.Errorf("Can not parse sport category: %s", dataStr)
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
