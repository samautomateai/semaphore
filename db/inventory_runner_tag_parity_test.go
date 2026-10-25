package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInventoryValidateRunnerTag(t *testing.T) {
	t.Run("absent tag allowed", func(t *testing.T) {
		require.NoError(t, (Inventory{}).Validate())
	})
	t.Run("empty explicit tag rejected", func(t *testing.T) {
		empty := ""
		require.Error(t, (Inventory{RunnerTag: &empty}).Validate())
	})
	t.Run("nonempty tag allowed", func(t *testing.T) {
		tag := "linux"
		require.NoError(t, (Inventory{RunnerTag: &tag}).Validate())
	})
}
