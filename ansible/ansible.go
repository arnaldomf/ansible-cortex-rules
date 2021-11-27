package ansible

import (
	"ansible-cortex-rules/models"
	"ansible-cortex-rules/services"
)

type AnsibleCortexState struct {
	state        string
	RulerGroup   *models.RuleGroup
	FailedReason string
}

const (
	GroupStateNotFound     = "NOT_FOUND"
	GroupStateFound        = "FOUND"
	GroupStateFailed       = "FAILED"
	GroupStateNeedToChange = "NEED_CHANGE"
)

func CompareRuleGroups(cr *services.CortexRuler, localRuleGroup *models.RuleGroup, namespace string) AnsibleCortexState {
	ansibleState := AnsibleCortexState{}
	ruleGroupFromServer, err := cr.GetRuleGroup(namespace, localRuleGroup.Name)
	ansibleState.RulerGroup = ruleGroupFromServer
	switch err {
	case services.ErrRuleGroupNotFound:
		ansibleState.state = GroupStateNotFound
	case nil:
		ansibleState.state = GroupStateFound
		if !ansibleState.RulerGroup.IsEqual(localRuleGroup) {
			ansibleState.state = GroupStateNeedToChange
		}
	default:
		ansibleState.state = GroupStateFailed
		ansibleState.FailedReason = err.Error()
	}

	return ansibleState
}

func (s *AnsibleCortexState) ApplyChange(cr *services.CortexRuler, localRuleGroup *models.RuleGroup, namespace string) error {
	if !s.GroupNeedToChange() && !s.GroupNotFound() {
		return nil
	}

	return cr.SetRuleGroup(localRuleGroup, namespace)
}

func (s *AnsibleCortexState) GroupFound() bool {
	return s.state == GroupStateFound
}

func (s *AnsibleCortexState) GroupNotFound() bool {
	return s.state == GroupStateNotFound
}

func (s *AnsibleCortexState) GroupFailed() bool {
	return s.state == GroupStateFailed
}

func (s *AnsibleCortexState) GroupNeedToChange() bool {
	return s.state == GroupStateNeedToChange
}
