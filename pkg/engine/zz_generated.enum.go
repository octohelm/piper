/*
Package engine GENERATED BY gengo:enum
DON'T EDIT THIS FILE
*/
package engine

import (
	errors "errors"
	fmt "fmt"
)

var InvalidLogLevel = errors.New("invalid LogLevel")

func (LogLevel) EnumValues() []any {
	return []any{
		DebugLevel, ErrorLevel, InfoLevel, WarnLevel,
	}
}

func ParseLogLevelLabelString(label string) (LogLevel, error) {
	switch label {
	case "debug":
		return DebugLevel, nil
	case "error":
		return ErrorLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn":
		return WarnLevel, nil

	default:
		return "", InvalidLogLevel
	}
}

func (v LogLevel) Label() string {
	switch v {
	case DebugLevel:
		return "debug"
	case ErrorLevel:
		return "error"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"

	default:
		return fmt.Sprint(v)
	}
}
