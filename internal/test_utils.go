package internal

type TestShellCommand struct {
	OutputterFunc func() ([]byte, error)
}

func (sc TestShellCommand) Output() ([]byte, error) {
	return sc.OutputterFunc()
}
