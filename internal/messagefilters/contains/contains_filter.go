package contains

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	FilterName              = "contains"
	configKeywordContains   = "contains"
	configKeywordIgnoreCase = "ignoreCase"
)

type ContainsFilter struct {
	contains   string
	ignoreCase bool
}

func NewContainsFilterFromMap(args map[string]interface{}) (*ContainsFilter, error) {
	expectedTypes := map[string]reflect.Kind{
		configKeywordContains:   reflect.String,
		configKeywordIgnoreCase: reflect.Bool,
	}
	for keyword, valueType := range expectedTypes {
		t, ok := args[keyword]
		if !ok {
			return nil, fmt.Errorf("no '%s' in args", keyword)
		}

		switch v := reflect.ValueOf(t); v.Kind() {
		case valueType:
		default:
			return nil, fmt.Errorf("expected type %v for keyword '%s'", valueType, keyword)
		}
	}

	return NewContainsFilter(args[configKeywordContains].(string), args[configKeywordIgnoreCase].(bool))
}

func NewContainsFilter(keyword string, ignoreCase bool) (*ContainsFilter, error) {
	if len(keyword) == 0 {
		return nil, errors.New("empty contains given")
	}

	if ignoreCase {
		keyword = strings.ToLower(keyword)
	}

	return &ContainsFilter{
		contains:   keyword,
		ignoreCase: ignoreCase,
	}, nil
}

func (c *ContainsFilter) Accept(msg string) bool {
	if c.ignoreCase {
		msg = strings.ToLower(msg)
	}
	return strings.Contains(msg, c.contains)
}
