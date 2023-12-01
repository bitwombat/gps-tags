package oneshot

type Config struct {
	SetIf   func() bool
	OnSet   func() error
	ResetIf func() bool
	OnReset func() error
}

func SetOrReset(event string, storage map[string]bool, config Config) error {

	if storage == nil {
		panic("storage passed to oneshot is nil")
	}

	var err error

	if config.SetIf != nil {
		if !storage[event] && config.SetIf() {
			if config.OnSet != nil {
				err = config.OnSet()
			}
			if err == nil {
				storage[event] = true
			}
		}
	}

	err = nil

	if config.ResetIf != nil {
		if storage[event] && config.ResetIf() {
			if config.OnReset != nil {
				err = config.OnReset()
			}
			if err == nil {
				storage[event] = false
			}
		}
	}

	return nil
}
