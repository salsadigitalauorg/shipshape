package lagoon_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/fact/testdata"
	"github.com/salsadigitalauorg/shipshape/pkg/lagoon"
)

func TestFactsFromFacters(t *testing.T) {
	tt := []struct {
		name    string
		facters map[string]fact.Facter
		source  string
		want    []lagoon.Fact
	}{
		{
			name: "FormatString",
			facters: map[string]fact.Facter{
				"webform-count": testdata.New("webform-count", data.FormatString, "12"),
			},
			source: "Shipshape",
			want: []lagoon.Fact{
				{Name: "webform-count", Value: "12", Source: "Shipshape", Category: "webform-count"},
			},
		},
		{
			name: "FormatMapString",
			facters: map[string]fact.Facter{
				"db-permissions": testdata.New("db-permissions", data.FormatMapString,
					map[string]string{"govcms_content_author": "import configuration"}),
			},
			source: "Shipshape",
			want: []lagoon.Fact{
				{Name: "govcms_content_author", Value: "import configuration", Source: "Shipshape", Category: "db-permissions"},
			},
		},
		{
			name: "FormatMapListString joins values with comma-space",
			facters: map[string]fact.Facter{
				"roles": testdata.New("roles", data.FormatMapListString,
					map[string][]string{"editor": {"create", "update"}}),
			},
			source: "Shipshape",
			want: []lagoon.Fact{
				{Name: "editor", Value: "create, update", Source: "Shipshape", Category: "roles"},
			},
		},
		{
			name: "FormatMapNestedString uses outer: inner as name",
			facters: map[string]fact.Facter{
				"webforms": testdata.New("webforms", data.FormatMapNestedString,
					map[string]map[string]string{
						"contact": {"to_mail": "[current-user]"},
					}),
			},
			source: "Shipshape",
			want: []lagoon.Fact{
				{Name: "contact: to_mail", Value: "[current-user]", Source: "Shipshape", Category: "webforms"},
			},
		},
		{
			name: "FormatListString emits one fact per item under the fact id",
			facters: map[string]fact.Facter{
				"disallowed-scripts": testdata.New("disallowed-scripts", data.FormatListString,
					[]string{"a.php", "b.php"}),
			},
			source: "Shipshape",
			want: []lagoon.Fact{
				{Name: "disallowed-scripts", Value: "a.php", Source: "Shipshape", Category: "disallowed-scripts"},
				{Name: "disallowed-scripts", Value: "b.php", Source: "Shipshape", Category: "disallowed-scripts"},
			},
		},
		{
			name: "FormatNil is skipped",
			facters: map[string]fact.Facter{
				"empty": testdata.New("empty", data.FormatNil, nil),
			},
			source: "Shipshape",
			want:   []lagoon.Fact{},
		},
		{
			name: "empty format is skipped",
			facters: map[string]fact.Facter{
				"uncollected": testdata.New("uncollected", "", nil),
			},
			source: "Shipshape",
			want:   []lagoon.Fact{},
		},
		{
			name: "FormatRaw is skipped with a warning",
			facters: map[string]fact.Facter{
				"raw": testdata.New("raw", data.FormatRaw, []byte("binary")),
			},
			source: "Shipshape",
			want:   []lagoon.Fact{},
		},
		{
			name: "FormatMapBytes is skipped with a warning",
			facters: map[string]fact.Facter{
				"filedata": testdata.New("filedata", data.FormatMapBytes,
					map[string][]byte{"a.yml": []byte("data")}),
			},
			source: "Shipshape",
			want:   []lagoon.Fact{},
		},
		{
			name: "multiple facts sorted deterministically by id",
			facters: map[string]fact.Facter{
				"zzz": testdata.New("zzz", data.FormatString, "last"),
				"aaa": testdata.New("aaa", data.FormatString, "first"),
			},
			source: "Shipshape",
			want: []lagoon.Fact{
				{Name: "aaa", Value: "first", Source: "Shipshape", Category: "aaa"},
				{Name: "zzz", Value: "last", Source: "Shipshape", Category: "zzz"},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			for id, f := range tc.facters {
				f.(*testdata.TestFacter).Collect()
				tc.facters[id] = f
			}
			got := lagoon.FactsFromFacters(tc.facters, tc.source)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFactsFromFactersCategoryAndDescription(t *testing.T) {
	f := testdata.New("db-permissions", data.FormatString, "import configuration")
	f.Collect()
	f.Category = "custom-category"
	f.Description = "Disallowed permissions on active site"

	got := lagoon.FactsFromFacters(map[string]fact.Facter{"db-permissions": f}, "Shipshape")
	assert.Equal(t, []lagoon.Fact{
		{
			Name:        "db-permissions",
			Value:       "import configuration",
			Source:      "Shipshape",
			Category:    "custom-category",
			Description: "Disallowed permissions on active site",
		},
	}, got)
}

func TestFactsFromFactersCategoryDefaultsToFactId(t *testing.T) {
	f := testdata.New("db-permissions", data.FormatString, "import configuration")
	f.Collect()

	got := lagoon.FactsFromFacters(map[string]fact.Facter{"db-permissions": f}, "Shipshape")
	assert.Len(t, got, 1)
	assert.Equal(t, "db-permissions", got[0].Category)
	assert.Equal(t, "", got[0].Description)
}

func TestFactsFromFactersJSONShape(t *testing.T) {
	f := testdata.New("db-permissions", data.FormatMapString,
		map[string]string{"govcms_content_author": "import configuration"})
	f.Collect()
	f.Category = "db-permissions"
	f.Description = "[DATABASE] Disallowed permissions on active site"

	facts := lagoon.FactsFromFacters(map[string]fact.Facter{"db-permissions": f}, "Shipshape")

	type payload struct {
		Project     string        `json:"project"`
		Environment string        `json:"environment"`
		Facts       []lagoon.Fact `json:"facts"`
	}
	p := payload{Project: "test-shipshape", Environment: "master", Facts: facts}

	b, err := json.Marshal(p)
	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"project": "test-shipshape",
		"environment": "master",
		"facts": [
			{
				"category": "db-permissions",
				"source": "Shipshape",
				"name": "govcms_content_author",
				"description": "[DATABASE] Disallowed permissions on active site",
				"value": "import configuration"
			}
		]
	}`, string(b))
}
