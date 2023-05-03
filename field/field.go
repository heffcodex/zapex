package zf

import (
	"go.uber.org/zap"
)

type Field interface {
	Field() zap.Field
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}

	return b
}

func errsToStrings(errs ...error) []string {
	strs := make([]string, len(errs))

	for i, err := range errs {
		strs[i] = err.Error()
	}

	return strs
}
