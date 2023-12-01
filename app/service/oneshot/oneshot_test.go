package oneshot

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSets(t *testing.T) {
	// GIVEN some storage and a flag to store that an action happened
	var storage map[string]bool
	storage = make(map[string]bool)
	var fired bool

	// WHEN the condition to set the oneshot is true
	err := SetOrReset("someEvent", storage,
		Config{
			SetIf: func() bool {
				return true
			},
			OnSet: func() error {
				fired = true
				return nil
			},
			ResetIf: nil,
			OnReset: nil,
		})

	// THEN we expect the set action to have happened
	require.Nil(t, err)
	require.True(t, fired)

}

func TestItDoesntSetIfConditionIsntTrue(t *testing.T) {
	// GIVEN some storage and a flag to store that an action happened
	var storage map[string]bool
	storage = make(map[string]bool)
	var fired bool

	// WHEN the condition to set the oneshot is NOT true
	err := SetOrReset("someEvent", storage,
		Config{
			SetIf: func() bool {
				return false
			},
			OnSet: func() error {
				fired = true
				return nil
			},
			ResetIf: nil,
			OnReset: nil,
		})

	// THEN we expect the set action to NOT have happened
	require.Nil(t, err)
	require.False(t, fired)

}

func TestItDoesntSetIfActionFails(t *testing.T) {
	// GIVEN some storage and a flag to store that an action happened
	var storage map[string]bool
	storage = make(map[string]bool)
	var firedOnce bool
	var firedTwice bool

	// WHEN the condition to set the oneshot is true but the action fails
	err := SetOrReset("someEvent", storage,
		Config{
			SetIf: func() bool {
				return true
			},
			OnSet: func() error {
				firedOnce = true
				return fmt.Errorf("some error")
			},
			ResetIf: nil,
			OnReset: nil,
		})

	// AND WHEN we attempt to set the oneshot again
	err = SetOrReset("someEvent", storage,
		Config{
			SetIf: func() bool {
				return true
			},
			OnSet: func() error {
				firedTwice = true
				return nil
			},
			ResetIf: nil,
			OnReset: nil,
		})

	// THEN we expect the set action to have happened again
	require.Nil(t, err)
	require.True(t, firedOnce)
	require.True(t, firedTwice)

}

func TestItDoesntFireSetActionTwice(t *testing.T) {
	// GIVEN some storage and two flags to store that actions happened
	var firedOnce bool
	var firedTwice bool
	var storage map[string]bool
	storage = make(map[string]bool)

	// WHEN the condition to set the oneshot is true
	err := SetOrReset("someEvent", storage,
		Config{
			SetIf: func() bool {
				return true
			}, OnSet: func() error {
				firedOnce = true
				return nil
			},
			ResetIf: nil,
			OnReset: nil,
		})
	require.Nil(t, err)
	require.True(t, firedOnce)

	// AND we attempt to set the oneshot a second time
	err = SetOrReset("someEvent", storage,
		Config{
			SetIf: func() bool {
				return true
			},
			OnSet: func() error {
				firedTwice = true
				return nil
			},
			ResetIf: nil,
			OnReset: nil,
		})

	// THEN the set action fired once but not the second time
	require.Nil(t, err)
	require.True(t, firedOnce)
	require.False(t, firedTwice)

}

func TestReset(t *testing.T) {
	// GIVEN some storage and a flag to store that an action happened
	var storage map[string]bool
	storage = make(map[string]bool)
	var resetFired bool

	// AND the oneshot is set
	err := SetOrReset("someEvent", storage,
		Config{
			SetIf: func() bool {
				return true
			},
			OnSet:   nil,
			ResetIf: nil,
			OnReset: nil,
		})

	// WHEN we reset the oneshot
	err = SetOrReset("someEvent", storage,
		Config{
			SetIf: nil,
			OnSet: nil,
			ResetIf: func() bool {
				return true
			},
			OnReset: func() error {
				resetFired = true
				return nil
			},
		})

	// THEN we expect the reset action to have happened
	require.Nil(t, err)
	require.True(t, resetFired)

}

func TestItDoesntFireResetActionTwice(t *testing.T) {
	// GIVEN some storage and flags to store that actions happened
	var storage map[string]bool
	storage = make(map[string]bool)
	var resetFiredOnce bool
	var resetFiredTwice bool

	// AND the oneshot is set
	err := SetOrReset("someEvent", storage,
		Config{
			SetIf: func() bool {
				return true
			},
			OnSet:   nil,
			ResetIf: nil,
			OnReset: nil,
		})

	// WHEN we reset the oneshot
	err = SetOrReset("someEvent", storage,
		Config{
			SetIf: nil,
			OnSet: nil,
			ResetIf: func() bool {
				return true
			},
			OnReset: func() error {
				resetFiredOnce = true
				return nil
			},
		})

	// WHEN we reset the oneshot
	err = SetOrReset("someEvent", storage,
		Config{
			SetIf: nil,
			OnSet: nil,
			ResetIf: func() bool {
				return true
			},
			OnReset: func() error {
				resetFiredTwice = true
				return nil
			},
		})

	// THEN we expect the reset action to have happened the first time but not the second
	require.Nil(t, err)
	require.True(t, resetFiredOnce)
	require.False(t, resetFiredTwice)
}

func TestItCanBeSetAndReset(t *testing.T) {
	// GIVEN some storage and flags to store that actions happened
	var storage map[string]bool
	storage = make(map[string]bool)
	var setFiredSecondTime bool

	// AND the oneshot is set
	err := SetOrReset("someEvent", storage,
		Config{
			SetIf: func() bool {
				return true
			},
			OnSet:   nil,
			ResetIf: nil,
			OnReset: nil,
		})

	// AND we reset the oneshot
	err = SetOrReset("someEvent", storage,
		Config{
			SetIf: nil,
			OnSet: nil,
			ResetIf: func() bool {
				return true
			},
			OnReset: nil,
		})

	// WHEN we set the oneshot again
	err = SetOrReset("someEvent", storage,
		Config{
			SetIf: func() bool {
				return true
			},
			OnSet: func() error {
				setFiredSecondTime = true
				return nil
			},
			ResetIf: nil,
			OnReset: nil,
		})

	// THEN we expect the second set action to have happened
	require.Nil(t, err)
	require.True(t, setFiredSecondTime)
}

// TODO: Test that events are tracked separately (how?), ie. the map is a map. Necessary?
