package zapex

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDefault(t *testing.T) {
	l, err := New(zap.DebugLevel.String())
	require.NoError(t, err)

	SetDefault(l)

	Default().Error("error",
		zap.NamedError("named", errors.New("named-error")),
		zap.Namespace("ns"),
		zap.Error(errors.New("error")),
		zap.Errors("errors", []error{errors.New("error1"), errors.New("error2")}),
	)

	_ = Default().Sync()
}
