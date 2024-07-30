package data

type DataFormat string

const (
	// FormatRaw is used to represent []byte.
	FormatRaw DataFormat = "raw"
	// FormatString is used to represent string.
	FormatString DataFormat = "string"
	// FormatListString is used to represent []string.
	FormatListString DataFormat = "list-string"
	// FormatListMapString is used to represent []map[string]string.
	FormatListMapString DataFormat = "list-map-string"
	// FormatMapBytes is used to represent map[string][]byte.
	FormatMapBytes DataFormat = "map-bytes"
	// FormatMapString is used to represent map[string]string.
	FormatMapString DataFormat = "map-string"
	// FormatMapListString is used to represent map[string][]string.
	FormatMapListString DataFormat = "map-list-string"
	// FormatMapNestedString is used to represent map[string]map[string]string.
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

	if parsedData, ok := data.(map[string]map[string]string); ok {
		return parsedData
	} else if parsedData, ok := data.(map[string]interface{}); ok {
		strStrMap := map[string]map[string]string{}
		for k, v := range parsedData {
			strStrMap[k] = v.(map[string]string)
		}
		return strStrMap
	}
	panic("unexpected data type when converting to MapNestedString")
}
