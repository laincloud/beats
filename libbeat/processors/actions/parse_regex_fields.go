package actions

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/processors"
)

type parseRegexFields struct {
	Regexp *regexp.Regexp
}

type parseRegexFieldsConfig struct {
	Regexp string `config:"regexp"`
}

func init() {
	processors.RegisterPlugin("parse_regex_fields",
		configChecked(newParseRegexFields, allowedFields("when", "regexp"), requireFields("regexp")))
}

func newParseRegexFields(c common.Config) (processors.Processor, error) {
	config := parseRegexFieldsConfig{}
	err := c.Unpack(&config)
	if err != nil {
		return nil, fmt.Errorf("fail to unpack the parse_regex_fields configuration: %s", err)
	}
	p := parseRegexFields{}
	if p.Regexp, err = regexp.Compile(config.Regexp); err != nil {
		err = fmt.Errorf("fail to compile the regexp of parse_regex_fields: %s", err)
		return nil, err
	}
	return &p, nil
}

func (p parseRegexFields) Run(event common.MapStr) (common.MapStr, error) {
	newEvent := event.Clone()
	messageObj, err := newEvent.GetValue("message")
	if err != nil {
		return newEvent, fmt.Errorf("process event failed: %s", err.Error())
	}
	message, ok := messageObj.(string)
	if !ok {
		return newEvent, fmt.Errorf("process event failed: %s", err.Error())
	}
	findResults := p.Regexp.FindStringSubmatch(message)
	if findResults == nil {
		return newEvent, nil
	}
	for index, name := range p.Regexp.SubexpNames() {
		if index == 0 {
			continue
		}
		if name == "" {
			name = fmt.Sprintf("%d", index)
		}
		newEvent.Put(name, findResults[index])
	}
	return newEvent, nil
}

func (p parseRegexFields) String() string {
	return "regex_fields=" + strings.Join(p.Regexp.SubexpNames(), ", ")
}
