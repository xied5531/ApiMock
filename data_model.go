package main

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type Server struct {
	Name         string `yaml:"name"`
	Address      string `yaml:"address"`
	ReadTimeout  int    `yaml:"read_timeout_s,omitempty"`
	WriteTimeout int    `yaml:"write_timeout_s,omitempty"`
	CertFile     string `yaml:"cert_file,omitempty"`
	KeyFile      string `yaml:"key_file,omitempty"`
	Apis         []Api  `yaml:"apis"`
}

type Api struct {
	Request struct {
		Url      string `yaml:"url"`
		Method   string `yaml:"method"`
		Metadata struct {
			PathVars     []string `yaml:"path_vars,omitempty"`
			HeaderKeys   []string `yaml:"header_keys,omitempty"`
			QueryParams  []string `yaml:"query_params,omitempty"`
			FormVars     []string `yaml:"form_vars,omitempty"`
			JsonBodyKeys []string `yaml:"json_body_keys,omitempty"`
		} `yaml:"metadata,omitempty"`
	} `yaml:"request"`
	Response struct {
		Status      int    `yaml:"status"`
		ContentType string `yaml:"content_type,omitempty"`
		Body        string `yaml:"body,omitempty"`
		Headers     []struct {
			Key   string `yaml:"key,omitempty"`
			Value string `yaml:"value,omitempty"`
		} `yaml:"headers,omitempty"`
	} `yaml:"response"`
}

func initMockData(mockDataFile string) {
	var filePath = mockDataFile
	if len(filePath) == 0 {
		filePath = "mock_data.yaml"
	}
	d, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Can not find data file[%s].", filePath)
	} else {
		err = yaml.Unmarshal(d, &ApiMockData)
		if err != nil {
			log.Fatalf("Data file[%s] unmarshal error: %v", filePath, err)
		}
	}
}
