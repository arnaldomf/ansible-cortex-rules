package services

import (
	"ansible-cortex-rules/models"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	ListRulesURL  = "/api/v1/rules"
	ListAlertsURL = "/api/v1/alerts"
)

var (
	ErrRuleGroupNotFound = errors.New("rule group not found")
)

type CortexRuler struct {
	RootURL              string
	PrometheusHTTPPrefix string
	Context              context.Context
}

func NewCortexRuler(ctx context.Context, rootURL, prometheusHTTPPrefix string) *CortexRuler {
	if len(prometheusHTTPPrefix) == 0 {
		prometheusHTTPPrefix = "/prometheus"
	}

	prometheusHTTPPrefix = strings.TrimSuffix(prometheusHTTPPrefix, "/")

	return &CortexRuler{
		RootURL:              rootURL,
		PrometheusHTTPPrefix: prometheusHTTPPrefix,
		Context:              ctx,
	}
}

func (cr *CortexRuler) GetRuleGroup(namespace, name string) (*models.RuleGroup, error) {
	ruleGroup := new(models.RuleGroup)
	requestURL := fmt.Sprintf("%s%s/%s/%s",
		cr.RootURL,
		ListRulesURL,
		namespace,
		name)
	req, err := http.NewRequestWithContext(cr.Context, "GET", requestURL, nil)
	if err != nil {
		return ruleGroup, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ruleGroup, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return ruleGroup, ErrRuleGroupNotFound
	}
	if resp.StatusCode > 299 {
		return ruleGroup, fmt.Errorf("GetRuleGroup: http error: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ruleGroup, err
	}
	if err = yaml.Unmarshal(data, ruleGroup); err != nil {
		return ruleGroup, err
	}
	return ruleGroup, nil
}

func (cr *CortexRuler) SetRuleGroup(ruleGroup *models.RuleGroup, namespace string) error {
	requestURL := fmt.Sprintf("%s%s/%s",
		cr.RootURL,
		ListRulesURL,
		namespace)
	data, err := yaml.Marshal(*ruleGroup)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(data)
	req, err := http.NewRequestWithContext(cr.Context, "POST", requestURL, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content", "application/yaml")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 202 {
		return fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}
	return nil
}
