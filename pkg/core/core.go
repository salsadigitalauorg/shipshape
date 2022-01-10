package core

func (c *CheckBase) Init(pd string, ct CheckType) {
	c.ProjectDir = pd
	if c.Result.CheckType == "" {
		c.Result = Result{CheckType: ct}
	}
}

func (c *CheckBase) GetName() string {
	return c.Name
}

func (c *CheckBase) FetchData() error {
	return nil
}

func (c *CheckBase) RunCheck() error {
	return nil
}

func (c *CheckBase) GetResult() Result {
	return c.Result
}
