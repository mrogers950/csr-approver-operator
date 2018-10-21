package csrapprover

import (
	v1beta12 "k8s.io/api/certificates/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func subset(first, second []string) bool {
	set := sets.NewString(second...)
	return set.HasAll(first...)
}

func isSubset(first, second []string) bool {
	set := make(map[string]bool)
	for _, v := range second {
		set[v] = true
	}

	for _, v := range first {
		if !set[v] {
			return false
		}
	}
	return true
}

func usageSubset(csrUsages []v1beta12.KeyUsage, allowedUsages []v1beta12.KeyUsage) bool {
	if len(allowedUsages) == 0 || len(csrUsages) == 0 {
		return true
	}

	allowed := make([]string, 0)
	for i := range allowedUsages {
		allowed = append(allowed, string(allowedUsages[i]))
	}
	has := make([]string, 0)
	for i := range csrUsages {
		has = append(has, string(csrUsages[i]))
	}

	return subset(has, allowed)
}
