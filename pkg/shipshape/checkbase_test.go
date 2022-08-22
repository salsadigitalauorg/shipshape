package shipshape_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

func TestCheckBase(t *testing.T) {
	c := shipshape.CheckBase{Name: "foo"}
	if c.GetName() != "foo" {
		t.Errorf("name should be 'foo', got '%s'", c.GetName())
	}

	c.Init("baz", "")
	if shipshape.ProjectDir != "baz" {
		t.Errorf("name should be 'baz', got '%s'", shipshape.ProjectDir)
	}
	if c.Severity != shipshape.NormalSeverity {
		t.Errorf("severity should be '%s', got '%s'", shipshape.NormalSeverity, c.Severity)
	}
	if c.Result.Name != "foo" {
		t.Errorf("result name should be 'foo', got '%s'", c.Result.Name)
	}
	if c.Result.Severity != shipshape.NormalSeverity {
		t.Errorf("result severity should be '%s', got '%s'", shipshape.NormalSeverity, c.Severity)
	}

	c.FetchData()
	c.RunCheck()
	if msg, ok := internal.EnsureFail(t, &c); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c, []string{"not implemented"}); !ok {
		t.Error(msg)
	}
}
