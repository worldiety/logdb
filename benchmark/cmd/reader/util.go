package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

func niceNumber(v interface{}) string {
	switch t := v.(type) {
	case int64:
	case uint64:
		return fmt.Sprintf("%d", t)
	case float64:
		const epsilon = 1e-9
		if _, frac := math.Modf(math.Abs(t)); frac < epsilon || frac > 1.0-epsilon {
			return fmt.Sprintf("%.f", t)
		}
		return fmt.Sprintf("%.4f", t)
	}

	return fmt.Sprintf("%v", v)
}

func toJson(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func json2MarkdownTable(str string) string {
	obj := make(map[string]interface{})
	if err := json.Unmarshal([]byte(str), &obj); err != nil {
		panic(err)
	}

	sb := &strings.Builder{}
	for k, v := range obj {
		if v == nil {
			continue
		}

		var val string
		if otherObj, ok := v.(map[string]interface{}); ok {
			val = json2MarkdownList(" ", otherObj)
		} else {
			val = niceNumber(v)
		}
		sb.WriteString(fmt.Sprintf("|%s|%v|\n", k, val))
	}
	return sb.String()
}

func json2MarkdownList(indent string, obj map[string]interface{}) string {
	sb := &strings.Builder{}
	for k, v := range obj {
		if v == nil {
			continue
		}
		sb.WriteString(indent)
		var val string
		if otherObj, ok := v.(map[string]interface{}); ok {
			val = json2MarkdownList(indent+" ", otherObj)
		} else {
			val = niceNumber(v)
		}
		sb.WriteString(fmt.Sprintf("* %s:%s\n", k, val))
	}
	return sb.String()
}
