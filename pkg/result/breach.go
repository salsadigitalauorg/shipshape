package result

import (
	"fmt"
	"strings"
)

// Breach provides a representation for different breach types.
type Breach interface {
	GetCheckName() string
	GetCheckType() string
	GetSeverity() string
	GetType() BreachType
	SetCommonValues(checkType string, checkName string, severity string)
	String() string
}

type BreachType string

const (
	// BreachTypeValue is a breach with a value.
	BreachTypeValue BreachType = "value"
	// BreachTypeKeyValue is a breach with a key and a value.
	BreachTypeKeyValue BreachType = "key-value"
	// BreachTypeKeyValues is a breach with a key and a list of values.
	BreachTypeKeyValues BreachType = "key-values"
)

// Simple breach with no key.
// Example:
//
//	"file foo.ext not found": file is the ValueLabel, foo.ext is the Value
type ValueBreach struct {
	BreachType
	CheckType     string
	CheckName     string
	Severity      string
	ValueLabel    string
	Value         string
	ExpectedValue string
}

func (b *ValueBreach) GetCheckName() string {
	return b.CheckName
}

func (b *ValueBreach) GetCheckType() string {
	return b.CheckType
}

func (b *ValueBreach) GetSeverity() string {
	return b.Severity
}

func (b *ValueBreach) GetType() BreachType {
	return BreachTypeValue
}

func (b *ValueBreach) SetCommonValues(checkType string, checkName string, severity string) {
	b.BreachType = b.GetType()
	b.CheckType = checkType
	b.CheckName = checkName
	b.Severity = severity
}

func (b ValueBreach) String() string {
	if b.ValueLabel != "" {
		return fmt.Sprintf("[%s] %s", b.ValueLabel, b.Value)
	}
	return b.Value
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
	BreachType
	CheckType     string
	CheckName     string
	Severity      string
	KeyLabel      string
	Key           string
	ValueLabel    string
	Value         string
	ExpectedValue string
}

func (b *KeyValueBreach) GetCheckName() string {
	return b.CheckName
}

func (b *KeyValueBreach) GetCheckType() string {
	return b.CheckType
}

func (b *KeyValueBreach) GetType() BreachType {
	return BreachTypeKeyValue
}

func (b *KeyValueBreach) GetSeverity() string {
	return b.Severity
}

func (b *KeyValueBreach) SetCommonValues(checkType string, checkName string, severity string) {
	b.BreachType = b.GetType()
	b.CheckType = checkType
	b.CheckName = checkName
	b.Severity = severity
}

func (b KeyValueBreach) String() string {
	if b.ExpectedValue != "" {
		return fmt.Sprintf("[%s] '%s' equals '%s', expected '%s'", b.KeyLabel, b.Key, b.Value, b.ExpectedValue)
	}
	return fmt.Sprintf("[%s:%s] %s: %s", b.KeyLabel, b.Key, b.ValueLabel, b.Value)
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
	BreachType
	CheckType  string
	CheckName  string
	Severity   string
	KeyLabel   string
	Key        string
	ValueLabel string
	Values     []string
}

func (b *KeyValuesBreach) GetCheckName() string {
	return b.CheckName
}

func (b *KeyValuesBreach) GetCheckType() string {
	return b.CheckType
}

func (b *KeyValuesBreach) GetSeverity() string {
	return b.Severity
}

func (b *KeyValuesBreach) GetType() BreachType {
	return BreachTypeKeyValues
}

func (b *KeyValuesBreach) SetCommonValues(checkType string, checkName string, severity string) {
	b.BreachType = b.GetType()
	b.CheckType = checkType
	b.CheckName = checkName
	b.Severity = severity
}

func (b KeyValuesBreach) String() string {
	if b.KeyLabel != "" && b.ValueLabel != "" {
		return fmt.Sprintf("[%s:%s] %s: %s", b.KeyLabel, b.Key, b.ValueLabel,
			"["+strings.Join(b.Values, ", ")+"]")
	}
	return fmt.Sprintf("%s: %s", b.Key, "["+strings.Join(b.Values, ", ")+"]")
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
	} else if b, ok := bIfc.(*KeyValuesBreach); ok {
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
	return []string(nil)
}

func BreachGetExpectedValue(bIfc Breach) string {
	if b, ok := bIfc.(*ValueBreach); ok {
		return b.ExpectedValue
	} else if b, ok := bIfc.(*KeyValueBreach); ok {
		return b.ExpectedValue
	}
	return ""
}
