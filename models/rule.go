package models

type RuleGroup struct {
	Name     string `yaml:"name" json:"name"`
	Rules    []Rule `yaml:"rules" json:"rules"`
	Interval int    `yaml:"interval,omitempty" json:"interval,omitempty"`
}

type Rule struct {
	Record      string            `yaml:"record,omitempty" json:"record,omitempty"`
	Alert       string            `yaml:"alert,omitempty" json:"alert,omitempty"`
	Expression  string            `yaml:"expr" json:"expr"`
	For         int               `yaml:"for,omitempty" json:"for,omitempty"`
	Annotations map[string]string `yaml:"annotation,omitempty" json:"annotation,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

func areStringMapsEquals(m1, m2 map[string]string) bool {
	if len(m1) != len(m2) {
		return false
	}

	for k, v := range m1 {
		v2, ok := m2[k]
		if !ok || v2 != v {
			return false
		}
	}
	return true
}

func (r *Rule) IsEqual(r2 *Rule) bool {
	if r.Record == r2.Record && r.Expression == r2.Expression {
		// this is a recording rule
		return true
	}
	if r.Alert == r2.Alert && r.Expression == r2.Expression && r.For == r2.For {
		if !areStringMapsEquals(r.Annotations, r2.Annotations) {
			return false
		}
		if !areStringMapsEquals(r.Labels, r2.Labels) {
			return false
		}
		return true
	}
	return false
}

func (rg *RuleGroup) IsEqual(rg2 *RuleGroup) bool {
	if rg.Interval != rg2.Interval || rg.Name != rg2.Name {
		return false
	}
	if len(rg.Rules) != len(rg2.Rules) {
		return false
	}
	for i, rule := range rg.Rules {
		if !rule.IsEqual(&rg2.Rules[i]) {
			return false
		}
	}
	return true
}
