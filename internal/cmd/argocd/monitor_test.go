package argocd

import (
	"testing"

	argocdhelper "github.com/doctolib/yak/internal/helper/argocd"
	"github.com/stretchr/testify/assert"
)

type isResourceOrphanIgnoredScenario struct {
	resource argocdhelper.AppResource
	ignored  []argocdhelper.IgnoredOrphanResource
	Result   bool
}

func TestIsResourceOrphanIgnored(t *testing.T) {
	var testScenarios = map[string]isResourceOrphanIgnoredScenario{
		"isIgnored": {
			resource: argocdhelper.AppResource{Kind: "foo", Name: "foo", Group: "foo", Namespace: "foo"},
			ignored:  []argocdhelper.IgnoredOrphanResource{{Kind: "foo", Name: "foo", Group: "foo"}},
			Result:   true,
		},
		"notIgnored": {
			resource: argocdhelper.AppResource{Kind: "foo", Name: "foo", Group: "foo", Namespace: "foo"},
			ignored:  []argocdhelper.IgnoredOrphanResource{{Kind: "foo", Name: "foo", Group: "bar"}},
			Result:   false,
		},
		"notIgnoredSameGroup": {
			resource: argocdhelper.AppResource{Kind: "foo", Name: "foo", Group: "foo", Namespace: "foo"},
			ignored:  []argocdhelper.IgnoredOrphanResource{{Kind: "bar", Name: "foo", Group: "foo"}},
			Result:   false,
		},
		"isIgnoredWirldcard": {
			resource: argocdhelper.AppResource{Kind: "foo", Name: "foo", Group: "foo", Namespace: "foo"},
			ignored:  []argocdhelper.IgnoredOrphanResource{{Kind: "foo", Name: "*", Group: "foo"}},
			Result:   true,
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, v.Result, argocdhelper.IsResourceOrphanIgnored(v.resource, v.ignored))
		})
	}
}
