/*
Copyright 2022 Red Hat
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package condition

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateList returns a Conditions object from a list of Condition objects
func CreateList(conditions ...Condition) Conditions {
	cs := Conditions{}
	for _, c := range conditions {
		cs = append(cs, c)
	}
	return cs
}

// UnknownCondition returns a condition with Status=Unknown and the given type, reason and message.
func UnknownCondition(t Type, reason Reason, message string) Condition {
	return Condition{
		Type:               t,
		Status:             corev1.ConditionUnknown,
		Severity:           SeverityNone,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// TrueCondition returns a condition with Status=True and the given type.
func TrueCondition(t Type, message string) Condition {
	return Condition{
		Type:               t,
		Status:             corev1.ConditionTrue,
		Severity:           SeverityNone,
		LastTransitionTime: metav1.Now(),
		Reason:             ReadyReason,
		Message:            message,
	}
}

// FalseCondition returns a condition with Status=False and the given type, severity, reason, and message.
func FalseCondition(t Type, reason Reason, severity Severity, messageFormat string, messageArgs ...interface{}) Condition {
	return Condition{
		Type:               t,
		Status:             corev1.ConditionFalse,
		Severity:           severity,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            getMessage(messageFormat, messageArgs...),
	}
}

// getMessage formats the message
func getMessage(messageFormat string, messageArgs ...interface{}) string {
	if len(messageArgs) == 0 {
		return messageFormat
	}
	return messageFormat
}

// Init initializes the Conditions with the given list if not already set
func (c *Conditions) Init(cl *Conditions) {
	if len(*c) == 0 {
		*c = *cl
	}
}

// Set sets or updates a condition in the list
func (c *Conditions) Set(condition Condition) {
	for i, existing := range *c {
		if existing.Type == condition.Type {
			(*c)[i] = condition
			return
		}
	}
	*c = append(*c, condition)
}

// Get returns the condition with the given type, if it exists
func (c Conditions) Get(t Type) *Condition {
	for i := range c {
		if c[i].Type == t {
			return &c[i]
		}
	}
	return nil
}

// IsTrue returns true if the condition with the given type is True
func (c Conditions) IsTrue(t Type) bool {
	cond := c.Get(t)
	if cond == nil {
		return false
	}
	return cond.Status == corev1.ConditionTrue
}

// IsFalse returns true if the condition with the given type is False
func (c Conditions) IsFalse(t Type) bool {
	cond := c.Get(t)
	if cond == nil {
		return false
	}
	return cond.Status == corev1.ConditionFalse
}

// IsUnknown returns true if the condition with the given type is Unknown
func (c Conditions) IsUnknown(t Type) bool {
	cond := c.Get(t)
	if cond == nil {
		return true
	}
	return cond.Status == corev1.ConditionUnknown
}

// MarkTrue sets Status=True for the condition with the given type
func (c *Conditions) MarkTrue(t Type, message string) {
	c.Set(TrueCondition(t, message))
}

// MarkFalse sets Status=False for the condition with the given type
func (c *Conditions) MarkFalse(t Type, reason Reason, severity Severity, message string, messageArgs ...interface{}) {
	c.Set(FalseCondition(t, reason, severity, message, messageArgs...))
}

// MarkUnknown sets Status=Unknown for the condition with the given type
func (c *Conditions) MarkUnknown(t Type, reason Reason, message string) {
	c.Set(UnknownCondition(t, reason, message))
}

// AllSubConditionIsTrue returns true if all conditions except ReadyCondition are True
func (c Conditions) AllSubConditionIsTrue() bool {
	for _, cond := range c {
		if cond.Type == ReadyCondition {
			continue
		}
		if cond.Status != corev1.ConditionTrue {
			return false
		}
	}
	return true
}

// Mirror returns a condition based on the aggregated state of other conditions
func (c Conditions) Mirror(t Type) Condition {
	groups := c.groupConditions()

	// If there are any False conditions with Error severity, return a False condition
	for _, g := range groups {
		if g.status == corev1.ConditionFalse && g.severity == SeverityError {
			if len(g.conditions) > 0 {
				return Condition{
					Type:               t,
					Status:             corev1.ConditionFalse,
					Severity:           SeverityError,
					LastTransitionTime: metav1.Now(),
					Reason:             g.conditions[0].Reason,
					Message:            g.conditions[0].Message,
				}
			}
		}
	}

	// If there are any False conditions with Warning severity
	for _, g := range groups {
		if g.status == corev1.ConditionFalse && g.severity == SeverityWarning {
			if len(g.conditions) > 0 {
				return Condition{
					Type:               t,
					Status:             corev1.ConditionFalse,
					Severity:           SeverityWarning,
					LastTransitionTime: metav1.Now(),
					Reason:             g.conditions[0].Reason,
					Message:            g.conditions[0].Message,
				}
			}
		}
	}

	// If there are any Unknown conditions
	for _, g := range groups {
		if g.status == corev1.ConditionUnknown {
			if len(g.conditions) > 0 {
				return Condition{
					Type:               t,
					Status:             corev1.ConditionUnknown,
					Severity:           SeverityNone,
					LastTransitionTime: metav1.Now(),
					Reason:             g.conditions[0].Reason,
					Message:            g.conditions[0].Message,
				}
			}
		}
	}

	// All conditions are True
	return TrueCondition(t, ReadyMessage)
}

// groupConditions groups conditions by status and severity
func (c Conditions) groupConditions() []conditionGroup {
	groups := make(map[string]*conditionGroup)

	for _, cond := range c {
		if cond.Type == ReadyCondition {
			continue
		}
		key := string(cond.Status) + string(cond.Severity)
		if _, ok := groups[key]; !ok {
			groups[key] = &conditionGroup{
				status:   cond.Status,
				severity: cond.Severity,
			}
		}
		groups[key].conditions = append(groups[key].conditions, cond)
	}

	result := make([]conditionGroup, 0, len(groups))
	for _, g := range groups {
		result = append(result, *g)
	}

	// Sort by severity (Error > Warning > Info > None)
	sort.Slice(result, func(i, j int) bool {
		severityOrder := map[Severity]int{
			SeverityError:   0,
			SeverityWarning: 1,
			SeverityInfo:    2,
			SeverityNone:    3,
		}
		return severityOrder[result[i].severity] < severityOrder[result[j].severity]
	})

	return result
}

// RestoreLastTransitionTimes restores the LastTransitionTime for conditions that haven't changed
func RestoreLastTransitionTimes(conditions *Conditions, savedConditions *Conditions) {
	if savedConditions == nil {
		return
	}

	for i, c := range *conditions {
		for _, saved := range *savedConditions {
			if c.Type == saved.Type && c.Status == saved.Status {
				(*conditions)[i].LastTransitionTime = saved.LastTransitionTime
			}
		}
	}
}
