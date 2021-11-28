package ansible

import (
	"ansible-cortex-rules/models"
	"ansible-cortex-rules/services"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type Configuration struct {
	CheckMode      bool   `json:"_ansible_check_mode"`
	NoLog          bool   `json:"_ansible_no_log"`
	Debug          bool   `json:"_ansible_debug"`
	Diff           bool   `json:"_ansible_diff"`
	Verbosity      int    `json:"_ansible_verbosity"`
	Version        string `json:"_ansible_version"`
	ModuleName     string `json:"_ansible_module_name"`
	SyslogFacility string `json:"_ansible_syslog_facility"`
}

type ModuleConfiguration struct {
	models.RuleGroup
	RootURL       string `json:"root_url"`
	Namespace     string `json:"namespace"`
	LogFilePath   string `json:"log_file_path,omitempty"`
	cortexRuler   *services.CortexRuler
	configuration *Configuration
	response      *AnsibleResponse
	logger        *log.Logger
}

type AnsibleResponse struct {
	Message     string `json:"msg"`
	Changed     bool   `json:"changed"`
	Failed      bool   `json:"failed"`
	Unreachable bool   `json:"unreachable"`
}

func (m *ModuleConfiguration) logSetup() *log.Logger {
	if m.logger != nil {
		return m.logger
	}
	var writer io.Writer
	if len(m.LogFilePath) == 0 {
		writer = io.Discard
	} else {
		f, err := os.OpenFile(m.LogFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		writer = f
	}
	return log.New(writer, "", log.LstdFlags)
}

func (m *ModuleConfiguration) Logger() *log.Logger {
	m.logger = m.logSetup()
	return m.logger
}

func (m *ModuleConfiguration) ResponseMessage() string {
	return m.response.Message
}

func (m *ModuleConfiguration) RenderResponse() string {
	data, err := json.Marshal(*m.response)
	if err != nil {
		return ""
	}
	return string(data)
}

func ModuleSetup(ctx context.Context, ansibleInputFile io.Reader) (*ModuleConfiguration, error) {
	moduleConfiguration := &ModuleConfiguration{response: new(AnsibleResponse)}
	input, err := ioutil.ReadAll(ansibleInputFile)
	if err != nil {
		return moduleConfiguration, err
	}

	configuration := &Configuration{}
	err = json.Unmarshal(input, configuration)
	if err != nil {
		return moduleConfiguration, err
	}

	err = json.Unmarshal(input, moduleConfiguration)
	if err != nil {
		return moduleConfiguration, err
	}
	moduleConfiguration.cortexRuler = services.NewCortexRuler(ctx, moduleConfiguration.RootURL, "/prometheus")
	moduleConfiguration.configuration = configuration
	return moduleConfiguration, nil
}
