package argocd

import (
	"testing"

	argocdhelper "github.com/santi1s/yak/internal/helper/argocd"
	"github.com/stretchr/testify/assert"
)

type substractAppResourcesScenario struct {
	a, b, Result []argocdhelper.AppResource
}

func TestSubstractAppResources(t *testing.T) {
	var testScenarios = map[string]substractAppResourcesScenario{
		"diff": {
			a: []argocdhelper.AppResource{
				{Kind: "foo", Name: "foo", Group: "foo", Namespace: "foo"},
				{Kind: "bar", Name: "bar", Group: "bar", Namespace: "bar"},
			},
			b: []argocdhelper.AppResource{
				{Kind: "foo", Name: "foo", Group: "foo", Namespace: "foo"},
				{Kind: "baz", Name: "baz", Group: "baz", Namespace: "baz"},
			},
			Result: []argocdhelper.AppResource{
				{Kind: "bar", Name: "bar", Group: "bar", Namespace: "bar"},
			},
		},
		"no_diff": {
			a: []argocdhelper.AppResource{
				{Kind: "foo", Name: "foo", Group: "foo", Namespace: "foo"},
				{Kind: "bar", Name: "bar", Group: "bar", Namespace: "bar"},
			},
			b: []argocdhelper.AppResource{
				{Kind: "foo", Name: "foo", Group: "foo", Namespace: "foo"},
				{Kind: "bar", Name: "bar", Group: "bar", Namespace: "bar"},
			},
			Result: []argocdhelper.AppResource(nil),
		},
		"a_empty": {
			a: []argocdhelper.AppResource(nil),
			b: []argocdhelper.AppResource{
				{Kind: "foo", Name: "foo", Group: "foo", Namespace: "foo"},
				{Kind: "bar", Name: "bar", Group: "bar", Namespace: "bar"},
			},
			Result: []argocdhelper.AppResource(nil),
		},
		"b_empty": {
			a: []argocdhelper.AppResource{
				{Kind: "foo", Name: "foo", Group: "foo", Namespace: "foo"},
				{Kind: "bar", Name: "bar", Group: "bar", Namespace: "bar"},
			},
			b: []argocdhelper.AppResource(nil),
			Result: []argocdhelper.AppResource{
				{Kind: "foo", Name: "foo", Group: "foo", Namespace: "foo"},
				{Kind: "bar", Name: "bar", Group: "bar", Namespace: "bar"},
			},
		},
		"all_empty": {
			a:      []argocdhelper.AppResource(nil),
			b:      []argocdhelper.AppResource(nil),
			Result: []argocdhelper.AppResource(nil),
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, v.Result, argocdhelper.SubstractAppResources(v.a, v.b))
		})
	}
}
