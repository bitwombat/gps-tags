package sub

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetsContents(t *testing.T) {
	contents, err := GetContents("test.html", nil)
	require.Nil(t, err)
	require.NotEmpty(t, contents)
}

func Test_DoesSub(t *testing.T) {
	subs := map[string]string{
		"tuckerPositions": "THIS-SHOULD-BE-THERE",
	}
	contents, err := GetContents("test.html", subs)
	require.Nil(t, err)
	require.Contains(t, contents, "THIS-SHOULD-BE-THERE")
}
