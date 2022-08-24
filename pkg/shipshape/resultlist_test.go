package shipshape_test

import (
	"bytes"
	"testing"
	"text/tabwriter"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

func TestResultList(t *testing.T) {
	rl := shipshape.ResultList{
		Results: []shipshape.Result{
			{Status: shipshape.Pass},
			{Status: shipshape.Pass},
			{Status: shipshape.Pass},
		},
	}
	if rl.Status() != shipshape.Pass {
		t.Error("Status() should be Pass")
	}

	rl.Results[0].Status = shipshape.Fail
	if rl.Status() != shipshape.Fail {
		t.Error("Status() should be Fail")
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	rl = shipshape.ResultList{}
	rl.TableDisplay(w)
	if buf.String() != "No result available; ensure your shipshape.yml is configured correctly.\n" {
		t.Errorf("Buffer should indicate bad shipshape.yml, got '%s'", buf.String())
	}

	buf = bytes.Buffer{}
	rl = shipshape.ResultList{
		Results: []shipshape.Result{{Name: "a", Status: shipshape.Pass}},
	}
	rl.TableDisplay(w)
	if buf.String() != "NAME   STATUS   PASSES   FAILS\na      Pass              \n" {
		t.Errorf("buffer value not as expected, got '%#v'", buf.String())
	}

	buf = bytes.Buffer{}
	rl = shipshape.ResultList{
		Results: []shipshape.Result{
			{Name: "a", Status: shipshape.Pass},
			{Name: "b", Status: shipshape.Pass},
			{Name: "c", Status: shipshape.Pass},
		},
	}
	rl.TableDisplay(w)
	if buf.String() != "NAME   STATUS   PASSES   FAILS\na      Pass              \nb      Pass              \nc      Pass              \n" {
		t.Errorf("buffer value not as expected, got %#v", buf.String())
	}

	buf = bytes.Buffer{}
	rl = shipshape.ResultList{
		Results: []shipshape.Result{
			{
				Name:   "a",
				Status: shipshape.Pass,
				Passes: []string{"Pass a", "Pass ab"},
			},
			{
				Name:   "b",
				Status: shipshape.Pass,
				Passes: []string{"Pass b", "Pass bb", "Pass bc"},
			},
			{
				Name:     "c",
				Status:   shipshape.Fail,
				Failures: []string{"Fail c", "Fail cb"},
			},
			{
				Name:     "d",
				Status:   shipshape.Fail,
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
