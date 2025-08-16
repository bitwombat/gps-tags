package oneshot

type Config struct {
	SetIf   bool
	OnSet   func() error
	ResetIf bool
	OnReset func() error
}

type OneShot struct {
	storage map[string]bool
}

func NewOneShot() OneShot {
	return OneShot{
		storage: make(map[string]bool),
	}
}

func (o OneShot) SetReset(event string, config Config) error {

	var err error

	if !o.storage[event] && config.SetIf {
		if config.OnSet != nil {
			err = config.OnSet()
		}
		if err != nil {
			return err
		}

		o.storage[event] = true
	}

	if o.storage[event] && config.ResetIf {
		if config.OnReset != nil {
			err = config.OnReset()
		}
		if err != nil {
			return err
		}

		o.storage[event] = false
	}

	return nil
}
