package oneshot

type Config struct {
	SetIf   bool
	OnSet   func() error
	ResetIf bool
	OnReset func() error
}

type OneShot map[string]bool

func NewOneShot() OneShot {
	return make(map[string]bool)
}

func (o OneShot) SetReset(event string, config Config) error {

	var err error

	if !o[event] && config.SetIf {
		if config.OnSet != nil {
			err = config.OnSet()
		}
		if err != nil {
			return err
		}

		o[event] = true
	}

	if o[event] && config.ResetIf {
		if config.OnReset != nil {
			err = config.OnReset()
		}
		if err != nil {
			return err
		}

		o[event] = false
	}

	return nil
}
