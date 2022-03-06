package aravia_test

import (
	"testing"

	"github.com/Pauloo27/aravia"
	"github.com/stretchr/testify/require"
)

func TestGetWords(t *testing.T) {
	require.Equal(t, []string{"Post", "Stuff", "More", "Stuff"}, aravia.GetWords("PostStuffMoreStuff"))
	require.Equal(t, []string{"Post", "Stuff", "_More", "Stuff"}, aravia.GetWords("PostStuff_MoreStuff"))
}
