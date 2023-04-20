package config

// Breach provides a representation for different breach types.
type Breach interface{}

// Simple breach with no key.
// Example:
//
//	"file foo.ext not found": file is the ValueLabel, foo.ext is the Value
type ValueBreach struct {
	CheckType
	CheckName string
	Severity
	ValueLabel string
	Value      string
}

// Breach with key and value.
// Example:
//
//	"wordpress detected at /path/to/app":
//	  - file could be the KeyLabel
//	  - /path/to/app is the Key
//	  - app could be the ValueLabel
//	  - wordpress is the Value
type KeyValueBreach struct {
	CheckType
	CheckName string
	Severity
	KeyLabel   string
	Key        string
	ValueLabel string
	Value      string
}

// Breach with key and list of values.
// Example:
//
//	"[site_administrator] disallowed permissions: [administer site configuration, import configuration]":
//	  - role could be the KeyLabel
//	  - site_administrator is the Key
//	  - permissions could be the ValueLabel
//	  - [administer site configuration, import configuration] are the Values
type KeyValuesBreach struct {
	CheckType
	CheckName string
	Severity
	KeyLabel   string
	Key        string
	ValueLabel string
	Values     []string
}

func BreachGetCheckName(bIfc Breach) string {
	if b, ok := bIfc.(*ValueBreach); ok {
		return b.CheckName
	} else if b, ok := bIfc.(*KeyValueBreach); ok {
		return b.CheckName
	} else if b, ok := bIfc.(*KeyValuesBreach); ok {
		return b.CheckName
	}
	return ""
}

func BreachGetCheckType(bIfc Breach) CheckType {
	if b, ok := bIfc.(*ValueBreach); ok {
		return b.CheckType
	} else if b, ok := bIfc.(*KeyValueBreach); ok {
		return b.CheckType
	} else if b, ok := bIfc.(*KeyValuesBreach); ok {
		return b.CheckType
	}
	return ""
}

func BreachGetSeverity(bIfc Breach) Severity {
	if b, ok := bIfc.(*ValueBreach); ok {
		return b.Severity
	} else if b, ok := bIfc.(*KeyValueBreach); ok {
		return b.Severity
	} else if b, ok := bIfc.(*KeyValuesBreach); ok {
		return b.Severity
	}
	return ""
}

func BreachGetKeyLabel(bIfc Breach) string {
	if b, ok := bIfc.(*KeyValueBreach); ok {
		return b.KeyLabel
	} else if b, ok := bIfc.(*KeyValuesBreach); ok {
		return b.KeyLabel
	}
	return ""
}

func BreachGetKey(bIfc Breach) string {
	if b, ok := bIfc.(*KeyValueBreach); ok {
		return b.Key
	} else if b, ok := bIfc.(*KeyValuesBreach); ok {
		return b.Key
	}
	return ""
}

func BreachGetValueLabel(bIfc Breach) string {
	if b, ok := bIfc.(*ValueBreach); ok {
		return b.ValueLabel
	} else if b, ok := bIfc.(*KeyValueBreach); ok {
		return b.ValueLabel
	}
	return ""
}

func BreachGetValue(bIfc Breach) string {
	if b, ok := bIfc.(*ValueBreach); ok {
		return b.Value
	} else if b, ok := bIfc.(*KeyValueBreach); ok {
		return b.Value
	}
	return ""
}

func BreachGetValues(bIfc Breach) []string {
	if b, ok := bIfc.(*KeyValuesBreach); ok {
		return b.Values
	}
	return []string{}
}
