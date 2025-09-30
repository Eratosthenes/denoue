package denoue

import (
	"fmt"
	"sort"
	"strings"
)

const DEFAULT_TIME_LAYOUT string = "2006-01-02 3:04:05.000pm Z07"

// log levels
const (
	INFO  string = "INFO"
	WARN  string = "WARN"
	ERROR string = "ERROR"
)

// log keys
const (
	LEVEL_KEY string = "level"
	TIME_KEY  string = "time"
	MSG_KEY   string = "msgs"
	ERR_KEY   string = "error"
)

const (
	OC string = "{"
	CC string = "}"
	OB string = "["
	CB string = "]"
	QM string = "\""
)

func wrap(s string, tokens ...string) string {
	switch len(tokens) {
	case 1:
		return tokens[0] + s + tokens[0]
	case 2:
		return tokens[0] + s + tokens[1]
	default:
		panic("cannot parse")
	}
}

func (d JDict) String() string {
	keys := make([]string, 0, len(d.objects))
	default_keys := []string{TIME_KEY, LEVEL_KEY, ERR_KEY}

	// add default keys first
	for _, k := range default_keys {
		if _, found := d.objects[k]; found {
			keys = append(keys, k)
		}
	}

	// sort the non-default keys and add them
	nonDefaultKeys := make([]string, 0, len(d.objects))
	for k := range d.objects {
		if !in(default_keys, k) {
			nonDefaultKeys = append(nonDefaultKeys, k)
		}
	}
	sort.Strings(nonDefaultKeys)
	keys = append(keys, nonDefaultKeys...)

	// create the output
	out := ""
	for _, k := range keys {
		out += d.objects[k].String() + ", "
	}
	return wrap(out[:len(out)-2], OC, CC)
}

func in(vals []string, s string) bool {
	for _, v := range vals {
		if s == v {
			return true
		}
	}
	return false
}

func (g JGroup) String() string {
	return fmt.Sprintf("%v: %v", wrap(g.Key, QM), g.Dict)
}

func (a JArray) String() string {
	var sb strings.Builder

	sb.WriteString(wrap(a.Key, QM))
	sb.WriteString(": ")
	sb.WriteString(OB) // opening bracket

	first := true
	for _, val := range a.Vals {
		if first {
			first = false
			sb.WriteString(wrap(val, QM))
			continue
		}
		sb.WriteString(", ")
		sb.WriteString(wrap(val, QM))
	}

	if len(a.ByteVals) > 0 {
		for _, b := range a.ByteVals {
			if first {
				first = false
				sb.WriteString(QM)
				sb.Write(b)
				sb.WriteString(QM)
				continue
			}
			sb.WriteString(", ")
			sb.WriteString(QM)
			sb.Write(b)
			sb.WriteString(QM)
		}
	}

	sb.WriteString(CB) // closing bracket
	return sb.String()
}

func (p JPair) String() string {
	return fmt.Sprintf("%v: %v", wrap(p.Key, QM), wrap(p.Val, QM))
}
