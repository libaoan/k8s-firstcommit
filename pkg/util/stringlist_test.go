package util

import (
	"reflect"
	"testing"
)

func TestStringListSet(t *testing.T) {
	var s1 StringList
	s1.Set("foo,bar")
	s1.Set("hop")
	expected := []string{"foo", "bar", "hop"}
	if reflect.DeepEqual(expected, []string(s1)) == false {
		t.Errorf("expected: %v, got: %v", expected, s1)
	}
}

func TestStringListSetErr(t *testing.T) {
	var sl StringList
	if err := sl.Set(""); err == nil {
		t.Errorf("expected error for empty string")
	}
	if err := sl.Set(","); err == nil {
		t.Errorf("expected error for list of empty strings")
	}
}
