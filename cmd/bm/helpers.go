package bm

import (
	"strings"
)

const lorem = "Lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua Ut enim ad minim veniam quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat"

var l = strings.Split(lorem, " ")

func MakeRecord(numKeys int) map[string]interface{} {
	r := make(map[string]interface{}, numKeys)
	for i := 0; i < (2*numKeys)-1; i += 2 {
		r[l[i]] = l[i+1]
	}

	if len(r) == 0 {
		panic("empty record")
	}

	return r
}
