package data

type DataFormat string

const (
	FormatRaw             DataFormat = "raw"
	FormatString          DataFormat = "string"
	FormatListString      DataFormat = "list-string"
	FormatListMapString   DataFormat = "list-map-string"
	FormatMapBytes        DataFormat = "map-bytes"
	FormatMapString       DataFormat = "map-string"
	FormatMapListString   DataFormat = "map-list-string"
	FormatMapNestedString DataFormat = "map-nested-string"
)

func AsBytes(data interface{}) []byte {
	if data == nil {
		return nil
	}
	return data.([]byte)
}

func AsString(data interface{}) string {
	if data == nil {
		return ""
	}
	return data.(string)
}

func AsListString(data interface{}) []string {
	if data == nil {
		return nil
	}
	return data.([]string)
}

func AsListMapString(data interface{}) []map[string]string {
	if data == nil {
		return nil
	}
	return data.([]map[string]string)
}

func AsMapBytes(data interface{}) map[string][]byte {
	if data == nil {
		return nil
	}
	return data.(map[string][]byte)
}

func AsMapString(data interface{}) map[string]string {
	if data == nil {
		return nil
	}

	ifcMap := data.(map[string]interface{})
	strMap := map[string]string{}
	for k, v := range ifcMap {
		strMap[k] = v.(string)
	}

	return strMap
}

func AsMapListString(data interface{}) map[string][]string {
	if data == nil {
		return nil
	}
	return data.(map[string][]string)
}

func AsMapNestedString(data interface{}) map[string]map[string]string {
	if data == nil {
		return nil
	}
	return data.(map[string]map[string]string)
}
