package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ansible-cortex-rules/ansible"
	"ansible-cortex-rules/models"
	"ansible-cortex-rules/services"
)

type Response struct {
	Message string `json:"msg"`
	Changed bool   `json:"changed"`
	Failed  bool   `json:"failed"`
}

func RenderResponse(r *Response) string {
	data, err := json.Marshal(*r)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func main() {
	response := Response{}
	localRuleGroup := &models.RuleGroup{
		Name: "my-rule-group",
		Rules: []models.Rule{
			{
				Record:     "node_xpto_file_free",
				Expression: "100 * node_filesystem_free_bytes{fstype=~\"xfs|ext[34]\"} / node_filesystem_size_bytes{fstype=~\"xfs|ext[34]\"}",
			},
		},
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	cr := services.NewCortexRuler(ctx, "http://cortex.mgl1.magalucloud.com.br:9009", "/prometheus")
	state := ansible.CompareRuleGroups(cr, localRuleGroup, "rules")
	if state.GroupFailed() {
		response.Failed = true
		response.Message = state.FailedReason
		fmt.Println(RenderResponse(&response))
		return
	}

	if state.GroupNeedToChange() || state.GroupNotFound() {
		err := state.ApplyChange(cr, localRuleGroup, "rules")
		if err != nil {
			response.Failed = true
			response.Message = err.Error()
		} else {
			response.Failed = false
			response.Changed = true
		}
	}
	fmt.Println(RenderResponse(&response))
}
