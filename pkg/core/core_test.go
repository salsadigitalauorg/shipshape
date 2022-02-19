package core_test

import (
	"bytes"
	"salsadigitalauorg/shipshape/pkg/core"
	"testing"
	"text/tabwriter"
)

func TestCheckBase(t *testing.T) {
	c := core.CheckBase{Name: "foo"}
	if c.GetName() != "foo" {
		t.Errorf("name should be 'foo', got '%s'", c.GetName())
	}

	c.Init("baz", "")
	if c.ProjectDir != "baz" {
		t.Errorf("name should be 'baz', got '%s'", c.ProjectDir)
	}
	if c.Result.Name != "foo" {
		t.Errorf("result name should be 'foo', got '%s'", c.Result.Name)
	}

	c.FetchData()
	c.RunCheck()
	result := c.GetResult()
	if result.Status != core.Fail {
		t.Error("status should be fail")
	}
	if len(result.Failures) != 1 {
		t.Error("there should be exactly one failure")
	}
	if result.Failures[0] != "not implemented" {
		t.Errorf("failure should be 'not implemented', got '%s'", result.Failures[0])
	}
}

func TestResultList(t *testing.T) {
	rl := core.ResultList{
		Results: []core.Result{
			{Status: core.Pass},
			{Status: core.Pass},
			{Status: core.Pass},
		},
	}
	if rl.Status() != core.Pass {
		t.Error("Status() should be Pass")
	}

	rl.Results[0].Status = core.Fail
	if rl.Status() != core.Fail {
		t.Error("Status() should be Fail")
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	rl = core.ResultList{}
	rl.TableDisplay(w)
	if buf.String() != "" {
		t.Errorf("buffer should be empty, got '%s'", buf.String())
	}

	buf = bytes.Buffer{}
	rl = core.ResultList{
		Results: []core.Result{{Name: "a", Status: core.Pass}},
	}
	rl.TableDisplay(w)
	if buf.String() != "NAME   STATUS   PASSES   FAILS\na      Pass              \n" {
		t.Errorf("buffer value not as expected, got '%#v'", buf.String())
	}

	buf = bytes.Buffer{}
	rl = core.ResultList{
		Results: []core.Result{
			{Name: "a", Status: core.Pass},
			{Name: "b", Status: core.Pass},
			{Name: "c", Status: core.Pass},
		},
	}
	rl.TableDisplay(w)
	if buf.String() != "NAME   STATUS   PASSES   FAILS\na      Pass              \nb      Pass              \nc      Pass              \n" {
		t.Errorf("buffer value not as expected, got %#v", buf.String())
	}

	buf = bytes.Buffer{}
	rl = core.ResultList{
		Results: []core.Result{
			{
				Name:   "a",
				Status: core.Pass,
				Passes: []string{"Pass a", "Pass ab"},
			},
			{
				Name:   "b",
				Status: core.Pass,
				Passes: []string{"Pass b", "Pass bb", "Pass bc"},
			},
			{
				Name:     "c",
				Status:   core.Fail,
				Failures: []string{"Fail c", "Fail cb"},
			},
			{
				Name:     "d",
				Status:   core.Fail,
				Passes:   []string{"Pass d", "Pass db"},
				Failures: []string{"Fail c", "Fail cb"},
			},
		},
	}
	rl.TableDisplay(w)
	if buf.String() != "NAME   STATUS   PASSES    FAILS\na      Pass     Pass a    \n                Pass ab   \nb      Pass     Pass b    \n                Pass bb   \n                Pass bc   \nc      Fail               Fail c\n                          Fail cb\nd      Fail     Pass d    Fail c\n                Pass db   Fail cb\n" {
		t.Errorf("buffer value not as expected, got %#v", buf.String())
	}
}
