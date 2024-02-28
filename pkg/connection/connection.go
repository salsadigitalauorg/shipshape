package connection

var Connectioners = map[string]Connectioner{}

func GetInstance(name string) Connectioner {
	if c, ok := Connectioners[name]; !ok {
		return nil
	} else {
		return c
	}
}
