/*
 * Copyright (c) 2020, 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package v1

import (
	"encoding/json"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclock "k8s.io/utils/clock"
)

// clock is used to set status condition timestamps.
// This variable makes it easier to test conditions.
var clock kubeclock.Clock = &kubeclock.RealClock{}

// ConditionType is the type of the condition and is typically a CamelCased
// word or short phrase.
//
// Condition types should indicate state in the "abnormal-true" polarity. For
// example, if the condition indicates when a policy is invalid, the "is valid"
// case is probably the norm, so the condition should be called "Invalid".
type ConditionType string

// ConditionReason is intended to be a one-word, CamelCase representation of
// the category of cause of the current status. It is intended to be used in
// concise output, such as one-line kubectl get output, and in summarizing
// occurrences of causes.
type ConditionReason string

// Condition represents an observation of an object's state. Conditions are an
// extension mechanism intended to be used when the details of an observation
// are not a priori known or would not apply to all instances of a given Kind.
//
// Conditions should be added to explicitly convey properties that users and
// components care about rather than requiring those properties to be inferred
// from other observations. Once defined, the meaning of a Condition can not be
// changed arbitrarily - it becomes part of the API, and has the same
// backwards- and forwards-compatibility concerns of any other part of the API.
type Condition struct {
	Type               ConditionType          `json:"type"`
	Status             corev1.ConditionStatus `json:"status"`
	Reason             ConditionReason        `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
}

// IsTrue Condition whether the condition status is "True".
func (c Condition) IsTrue() bool {
	return c.Status == corev1.ConditionTrue
}

// IsFalse returns whether the condition status is "False".
func (c Condition) IsFalse() bool {
	return c.Status == corev1.ConditionFalse
}

// IsUnknown returns whether the condition status is "Unknown".
func (c Condition) IsUnknown() bool {
	return c.Status == corev1.ConditionUnknown
}

// DeepCopyInto copies in into out.
func (c *Condition) DeepCopyInto(cpy *Condition) {
	*cpy = *c
}

// Conditions is a set of Condition instances.
type Conditions []Condition

// NewConditions initializes a set of conditions with the given list of
// conditions.
func NewConditions(conds ...Condition) Conditions {
	conditions := Conditions{}
	for _, c := range conds {
		conditions.SetCondition(c)
	}
	return conditions
}

// NormalizeConditions removes invalid condition entries and deduplicates by
// condition type, keeping the condition with the newest transition time.
//
// Bug39366679/coherence.yaml showed that empty type/status entries can grow to
// tens of thousands of list items. Normalizing before status writes makes the
// next status patch replace that bloated list with the small API condition set.
func NormalizeConditions(in Conditions) Conditions {
	if len(in) == 0 {
		return in
	}

	byType := make(map[ConditionType]Condition, len(in))
	for _, condition := range in {
		if condition.Type == "" || condition.Status == "" {
			// Empty type or status cannot describe a useful API condition; dropping
			// them is what lets Bug39366679 repairs shrink the stored status.
			continue
		}

		existing, found := byType[condition.Type]
		if !found || shouldReplaceCondition(existing, condition) {
			byType[condition.Type] = condition
		}
	}

	conditions := make(Conditions, 0, len(byType))
	for _, condition := range byType {
		conditions = append(conditions, condition)
	}
	sort.Slice(conditions, func(a, b int) bool {
		return conditions[a].Type < conditions[b].Type
	})
	return conditions
}

func shouldReplaceCondition(existing, candidate Condition) bool {
	switch {
	case candidate.LastTransitionTime.After(existing.LastTransitionTime.Time):
		return true
	case existing.LastTransitionTime.After(candidate.LastTransitionTime.Time):
		return false
	default:
		// When Kubernetes timestamps tie, prefer the entry with user-visible context
		// so the normalized condition remains useful after deduplication.
		return conditionContextScore(candidate) > conditionContextScore(existing)
	}
}

func conditionContextScore(condition Condition) int {
	score := 0
	if condition.Reason != "" {
		score++
	}
	if condition.Message != "" {
		score++
	}
	return score
}

// IsTrueFor searches the set of conditions for a condition with the given
// ConditionType. If found, it returns `condition.IsTrue()`. If not found,
// it returns false.
func (in Conditions) IsTrueFor(t ConditionType) bool {
	for _, condition := range in {
		if condition.Type == t {
			return condition.IsTrue()
		}
	}
	return false
}

// IsFalseFor searches the set of conditions for a condition with the given
// ConditionType. If found, it returns `condition.IsFalse()`. If not found,
// it returns false.
func (in Conditions) IsFalseFor(t ConditionType) bool {
	for _, condition := range in {
		if condition.Type == t {
			return condition.IsFalse()
		}
	}
	return false
}

// IsUnknownFor searches the set of conditions for a condition with the given
// ConditionType. If found, it returns `condition.IsUnknown()`. If not found,
// it returns true.
func (in Conditions) IsUnknownFor(t ConditionType) bool {
	for _, condition := range in {
		if condition.Type == t {
			return condition.IsUnknown()
		}
	}
	return true
}

// SetCondition adds (or updates) the set of conditions with the given
// condition. It returns a boolean value indicating whether the set condition
// is new or was a change to the existing condition with the same type.
func (in *Conditions) SetCondition(newCond Condition) bool {
	newCond.LastTransitionTime = metav1.Time{Time: clock.Now()}

	for i, condition := range *in {
		if condition.Type == newCond.Type {
			if condition.Status == newCond.Status {
				newCond.LastTransitionTime = condition.LastTransitionTime
			}
			changed := condition.Status != newCond.Status ||
				condition.Reason != newCond.Reason ||
				condition.Message != newCond.Message
			(*in)[i] = newCond
			return changed
		}
	}
	*in = append(*in, newCond)
	return true
}

// GetCondition searches the set of conditions for the condition with the given
// ConditionType and returns it. If the matching condition is not found,
// GetCondition returns nil.
func (in Conditions) GetCondition(t ConditionType) *Condition {
	for _, condition := range in {
		if condition.Type == t {
			return &condition
		}
	}
	return nil
}

// RemoveCondition removes the condition with the given ConditionType from
// the conditions set. If no condition with that type is found, RemoveCondition
// returns without performing any action. If the passed condition type is not
// found in the set of conditions, RemoveCondition returns false.
func (in *Conditions) RemoveCondition(t ConditionType) bool {
	if in == nil {
		return false
	}
	for i, condition := range *in {
		if condition.Type == t {
			*in = append((*in)[:i], (*in)[i+1:]...)
			return true
		}
	}
	return false
}

// MarshalJSON marshals the set of conditions as a JSON array, sorted by
// condition type.
func (in Conditions) MarshalJSON() ([]byte, error) {
	conds := []Condition(in)
	sort.Slice(conds, func(a, b int) bool {
		return conds[a].Type < conds[b].Type
	})
	return json.Marshal(conds)
}
