package oneshot

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type stepType struct {
	config                   Config
	expectedSetActionCount   int
	expectedResetActionCount int
	expectedError            error
}

func TestOneShot(t *testing.T) {
	uut := NewOneShot()

	someError := fmt.Errorf("some error")

	var setActionCount int
	var resetActionCount int

	var SetFn = func() error {
		setActionCount++
		return nil
	}

	var SetFnWithError = func() error {
		setActionCount++
		return someError
	}

	var ResetFn = func() error {
		resetActionCount++
		return nil
	}

	var ResetFnWithError = func() error {
		resetActionCount++
		return someError
	}

	for _, tc := range []struct {
		description string
		steps       []stepType
	}{
		{
			description: "doesn't run any actions when no conditions true (FF)",
			steps: []stepType{
				{
					config: Config{
						SetIf:   false,
						OnSet:   SetFn,
						ResetIf: false,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   0,
					expectedResetActionCount: 0,
					expectedError:            nil,
				},
			},
		},
		{
			description: "runs just the set action when just SetIf is true (TF)",
			steps: []stepType{
				{
					config: Config{
						SetIf:   true,
						OnSet:   SetFn,
						ResetIf: false,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   1,
					expectedResetActionCount: 0,
					expectedError:            nil,
				},
			},
		},
		{
			description: "does not run the reset action if it hasn't been set (FT)",
			steps: []stepType{
				{
					config: Config{
						SetIf:   false,
						OnSet:   SetFn,
						ResetIf: true,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   0,
					expectedResetActionCount: 0,
					expectedError:            nil,
				},
			},
		},
		{
			description: "does run the reset action if it has been set",
			steps: []stepType{
				{
					config: Config{
						SetIf:   true,
						OnSet:   SetFn,
						ResetIf: false,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   1,
					expectedResetActionCount: 0,
					expectedError:            nil,
				},
				{
					config: Config{
						SetIf:   false,
						OnSet:   SetFn,
						ResetIf: true,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   1,
					expectedResetActionCount: 1,
					expectedError:            nil,
				},
			},
		},
		{
			description: "runs both actions when both Ifs are true simultaneously (TT)",
			steps: []stepType{
				{
					config: Config{
						SetIf:   true,
						OnSet:   SetFn,
						ResetIf: true,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   1,
					expectedResetActionCount: 1,
					expectedError:            nil,
				},
			},
		},
		{
			description: "remains reset when both Ifs are true (TT)",
			steps: []stepType{
				{
					config: Config{
						SetIf:   true,
						OnSet:   SetFn,
						ResetIf: true,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   1,
					expectedResetActionCount: 1,
					expectedError:            nil,
				},
				{
					config: Config{
						SetIf:   true,
						OnSet:   SetFn,
						ResetIf: false,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   2,
					expectedResetActionCount: 1,
					expectedError:            nil,
				},
			},
		},
		{
			description: "one-shots properly (doesn't fire twice)",
			steps: []stepType{
				{
					config: Config{
						SetIf:   true,
						OnSet:   SetFn,
						ResetIf: false,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   1,
					expectedResetActionCount: 0,
					expectedError:            nil,
				},
				{
					config: Config{
						SetIf:   true,
						OnSet:   SetFn,
						ResetIf: false,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   1,
					expectedResetActionCount: 0,
					expectedError:            nil,
				},
			},
		},
		{
			description: "doesn't set if action errors",
			steps: []stepType{
				{
					config: Config{
						SetIf:   true,
						OnSet:   SetFnWithError,
						ResetIf: false,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   1,
					expectedResetActionCount: 0,
					expectedError:            someError,
				},
				{
					config: Config{
						SetIf:   true,
						OnSet:   SetFn,
						ResetIf: false,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   2,
					expectedResetActionCount: 0,
					expectedError:            nil,
				},
			},
		},
		{
			description: "doesn't reset if action errors",
			steps: []stepType{
				{
					config: Config{
						SetIf:   true,
						OnSet:   SetFn,
						ResetIf: false,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   1,
					expectedResetActionCount: 0,
					expectedError:            nil,
				},
				{
					config: Config{
						SetIf:   false,
						OnSet:   SetFn,
						ResetIf: true,
						OnReset: ResetFnWithError,
					},
					expectedSetActionCount:   1,
					expectedResetActionCount: 1,
					expectedError:            someError,
				},
				{
					config: Config{
						SetIf:   true, // should not be able to set again
						OnSet:   SetFn,
						ResetIf: false,
						OnReset: ResetFn,
					},
					expectedSetActionCount:   1,
					expectedResetActionCount: 1,
					expectedError:            nil,
				},
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			for _, step := range tc.steps {
				err := uut.SetReset("testing", step.config)
				require.Equal(t, step.expectedError, err, "returned error")
				require.Equal(t, step.expectedSetActionCount, setActionCount, "set action count")
				require.Equal(t, step.expectedResetActionCount, resetActionCount, "reset action count")
			}
			setActionCount = 0
			resetActionCount = 0
			clear(uut.storage)
		})
	}
}

func TestEventsDontInterfere(t *testing.T) {
	// GIVEN a single oneshot instance
	uut := NewOneShot()

	var setActionCount int
	var resetActionCount int

	var SetFn = func() error {
		setActionCount++
		return nil
	}

	var ResetFn = func() error {
		resetActionCount++
		return nil
	}

	// WHEN we set event1
	err := uut.SetReset("event1", Config{
		SetIf:   true,
		OnSet:   SetFn,
		ResetIf: false,
		OnReset: ResetFn,
	},
	)

	// THEN we should see the normal behaviour
	require.Nil(t, err)
	require.Equal(t, 1, setActionCount)
	require.Equal(t, 0, resetActionCount)

	// AND WHEN we set event2
	err = uut.SetReset("event2", Config{
		SetIf:   true,
		OnSet:   SetFn,
		ResetIf: false,
		OnReset: ResetFn,
	},
	)

	// THEN we should see that it was independent - it also fired
	require.Nil(t, err)
	require.Equal(t, 2, setActionCount)
	require.Equal(t, 0, resetActionCount)

}
