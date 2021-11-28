package ansible

import (
	"ansible-cortex-rules/models"
	"ansible-cortex-rules/services"
	"log"
)

type AnsibleCortexState struct {
	state      string
	RulerGroup *models.RuleGroup
	logger     *log.Logger
}

const (
	GroupStateNotFound     = "NOT_FOUND"
	GroupStateFound        = "FOUND"
	GroupStateFailed       = "FAILED"
	GroupStateNeedToChange = "NEED_CHANGE"
)

func (m *ModuleConfiguration) Run(logger *log.Logger) error {
	state, err := m.CompareState(logger)
	if state.GroupFailed() {
		return err
	}
	if state.GroupNeedToChange() || state.GroupNotFound() {
		err = m.ApplyChange(&state)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *ModuleConfiguration) CompareState(logger *log.Logger) (AnsibleCortexState, error) {
	ansibleState := AnsibleCortexState{logger: logger}
	ruleGroupFromServer, err := m.cortexRuler.GetRuleGroup(m.Namespace, m.RuleGroup.Name)
	ansibleState.RulerGroup = ruleGroupFromServer
	switch err {
	case services.ErrRuleGroupNotFound:
		ansibleState.state = GroupStateNotFound
		ansibleState.logger.Printf("CompareRuleGroups: %s", ansibleState.state)
	case nil:
		ansibleState.state = GroupStateFound
		if !ansibleState.RulerGroup.IsEqual(&m.RuleGroup) {
			ansibleState.state = GroupStateNeedToChange
		}
		ansibleState.logger.Printf("CompareRuleGroups: %s", ansibleState.state)
	default:
		ansibleState.state = GroupStateFailed
		ansibleState.logger.Printf("CompareRuleGroups: %s", ansibleState.state)
		m.response.Message = err.Error()
		m.response.Unreachable = true
		return ansibleState, err
	}

	return ansibleState, nil
}

func (m *ModuleConfiguration) ApplyChange(state *AnsibleCortexState) error {
	state.logger.Printf("ApplyChange: state %s", state.state)
	if !state.GroupNeedToChange() && !state.GroupNotFound() {
		return nil
	}

	if m.configuration.CheckMode {
		state.logger.Print("Check mode enabled, doing nothing")
		m.response.Changed = true
		return nil
	}

	if err := m.cortexRuler.SetRuleGroup(&m.RuleGroup, m.Namespace); err != nil {
		m.response.Failed = true
		m.response.Message = err.Error()
		return err
	}
	m.response.Changed = true
	return nil
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
