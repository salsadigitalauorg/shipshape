package merger

import (
	"os"

	"gopkg.in/yaml.v3"
	// "github.com/ghodss/yaml"
)

type Merger struct {
	data map[string]interface{}
}

func NewMerger() *Merger {
	merger := new(Merger)
	merger.data = map[string]interface{}{}
	return merger
}

func (m *Merger) AddData(data []byte) error {

	var s1 interface{}
	err := yaml.Unmarshal(data, &s1)
	if err != nil {
		return err
	}

	return m.merge(s1.(map[string]interface{}))
}

func (m *Merger) merge(f map[string]interface{}) error {
	for key, item := range f {
		if i, ok := item.(map[string]interface{}); ok {
			for subKey, subitem := range i {
				if _, ok := m.data[key]; !ok {
					m.data[key] = map[string]interface{}{}
				}

				m.data[key].(map[string]interface{})[subKey] = subitem
			}
		} else {
			m.data[key] = item
		}
	}

	return nil
}

func (m *Merger) Save(fileName string) error {
	res, _ := yaml.Marshal(m.data)

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(res)
	if err != nil {
		return err
	}

	return nil
}
