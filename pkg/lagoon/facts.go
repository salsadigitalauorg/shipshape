package lagoon

import (
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
)

// factConfig is the optional interface a fact plugin implements to carry
// Lagoon-specific metadata. Matched via assertion, not added to the Facter
// interface, so existing fact plugins need no change.
type factConfig interface {
	GetCategory() string
	GetDescription() string
}

// FactsFromFacters maps collected facts to the addFactsByName mutation's
// input shape (see lagoon.Fact). source is the Lagoon Problem/Fact source to
// stamp on every emitted fact. Facts with an unsupported or empty format are
// skipped (with a warning logged for unsupported, non-nil formats).
func FactsFromFacters(facters map[string]fact.Facter, source string) []Fact {
	// Sort fact IDs for deterministic output.
	ids := make([]string, 0, len(facters))
	for id := range facters {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	facts := []Fact{}
	for _, id := range ids {
		f := facters[id]

		category := id
		description := ""
		if fc, ok := f.(factConfig); ok {
			category = fc.GetCategory()
			description = fc.GetDescription()
		}

		facts = append(facts, factsFromOne(id, f, source, category, description)...)
	}
	return facts
}

func factsFromOne(id string, f fact.Facter, source, category, description string) []Fact {
	newFact := func(name, value string) Fact {
		return Fact{
			Name:        name,
			Value:       value,
			Source:      source,
			Category:    category,
			Description: description,
		}
	}

	switch f.GetFormat() {
	case data.FormatString:
		return []Fact{newFact(id, data.AsString(f.GetData()))}

	case data.FormatMapString:
		m := data.AsMapString(f.GetData())
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		facts := make([]Fact, 0, len(keys))
		for _, k := range keys {
			facts = append(facts, newFact(k, m[k]))
		}
		return facts

	case data.FormatMapListString:
		m := data.AsMapListString(f.GetData())
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		facts := make([]Fact, 0, len(keys))
		for _, k := range keys {
			facts = append(facts, newFact(k, strings.Join(m[k], ", ")))
		}
		return facts

	case data.FormatMapNestedString:
		m := data.AsMapNestedString(f.GetData())
		outerKeys := make([]string, 0, len(m))
		for k := range m {
			outerKeys = append(outerKeys, k)
		}
		sort.Strings(outerKeys)
		facts := []Fact{}
		for _, ok := range outerKeys {
			innerMap := m[ok]
			innerKeys := make([]string, 0, len(innerMap))
			for ik := range innerMap {
				innerKeys = append(innerKeys, ik)
			}
			sort.Strings(innerKeys)
			for _, ik := range innerKeys {
				facts = append(facts, newFact(ok+": "+ik, innerMap[ik]))
			}
		}
		return facts

	case data.FormatListString:
		items := data.AsListString(f.GetData())
		facts := make([]Fact, 0, len(items))
		for _, v := range items {
			facts = append(facts, newFact(id, v))
		}
		return facts

	case data.FormatNil, "":
		return nil

	default:
		log.WithFields(log.Fields{
			"fact":   id,
			"format": f.GetFormat(),
		}).Warn("lagoon-facts output not supported for this format, skipping")
		return nil
	}
}
