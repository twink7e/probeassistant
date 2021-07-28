package util

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func GetFastLabelSelector(ps *metav1.LabelSelector) (labels.Selector, error) {
	var selector labels.Selector
	if len(ps.MatchExpressions) == 0 && len(ps.MatchLabels) != 0 {
		selector = labels.SelectorFromValidatedSet(ps.MatchLabels)
		return selector, nil
	}

	return metav1.LabelSelectorAsSelector(ps)
}
