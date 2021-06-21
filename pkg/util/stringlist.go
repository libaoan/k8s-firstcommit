package util

import (
	"fmt"
	"strings"
)

type StringList []string

func (sl *StringList) String() string {
	return fmt.Sprint(*sl)
}

func (sl *StringList) Set(value string) error {
	for _, s := range strings.Split(value, ",") {
		if len(s) == 0 {
			return fmt.Errorf("value should not be an empty string")
		}
		*sl = append(*sl, s)
	}
	return nil
}
