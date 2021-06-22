package core

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStatement(t *testing.T) {
	cmd := Statement("go build -o {{ bin }} main.go")
	res, err := cmd.Render(map[string]string{
		"bin": "main",
	})
	require.NoError(t, err)
	require.Equal(t, "go build -o main main.go", res)
}
