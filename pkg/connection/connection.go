package connection

var Connectioners = map[string]Connectioner{}

func GetConnection(name string) Connectioner {
	if c, ok := Connectioners[name]; !ok {
		return nil
	} else {
		return c
	}
}
