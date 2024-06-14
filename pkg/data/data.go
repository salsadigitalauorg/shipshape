package data

type DataFormat string

const (
	FormatRaw             DataFormat = "raw"
	FormatList            DataFormat = "list"
	FormatMapBytes        DataFormat = "map-bytes"
	FormatMapString       DataFormat = "map-string"
	FormatMapNestedString DataFormat = "map-nested-string"
)

func AsBytes(data interface{}) []byte {
	if data == nil {
		return nil
	}
	return data.([]byte)
}

func AsMapStringBytes(data interface{}) map[string][]byte {
	if data == nil {
		return nil
	}
	return data.(map[string][]byte)
}

func AsNestedStringMap(data interface{}) map[string]map[string]string {
	if data == nil {
		return nil
	}
	return data.(map[string]map[string]string)
}
