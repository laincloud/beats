package actions

import (
	"testing"

	"github.com/elastic/beats/libbeat/common"
	"github.com/stretchr/testify/assert"
)

func TestInvalidRegex(t *testing.T) {
	testConfig, _ := common.NewConfigFrom(map[string]interface{}{
		"regexp": "^(?P<failure>",
	})
	_, err := newParseRegexFields(*testConfig)
	assert.Error(t, err)
}

func TestNotMatch(t *testing.T) {
	testConfig, _ := common.NewConfigFrom(map[string]interface{}{
		"regexp": `^(?P<remote_addr>\S+)@(?P<remote_user>\S+)@\[(?P<time_local>[^\]]+)\]@(?P<request_host>\S+)`,
	})
	regexProc, err := newParseRegexFields(*testConfig)
	assert.NoError(t, err)
	srcEvent := common.MapStr{
		"message": "1234567@abcde",
	}
	expectedEvent, err := regexProc.Run(srcEvent)
	remoteAddrExist, _ := expectedEvent.HasKey("remote_addr")
	remoteUserExist, _ := expectedEvent.HasKey("remote_user")
	timeLocalExist, _ := expectedEvent.HasKey("time_local")
	requestHostExist, _ := expectedEvent.HasKey("request_host")
	messageExist, _ := expectedEvent.HasKey("message")
	assert.NoError(t, err)
	assert.False(t, remoteAddrExist)
	assert.False(t, remoteUserExist)
	assert.False(t, timeLocalExist)
	assert.False(t, requestHostExist)
	assert.True(t, messageExist)
}

func TestEmptySubmatch(t *testing.T) {
	testConfig, _ := common.NewConfigFrom(map[string]interface{}{
		"regexp": `^(\S+)@(\S+)@\[([^\]]+)\]@(\S+)`,
	})
	regexProc, err := newParseRegexFields(*testConfig)
	assert.NoError(t, err)
	srcEvent := common.MapStr{
		"message": "127.0.0.1@-@[15/May/2017:17:27:01 +0800]@filebeat.lain.test",
	}
	expectedEvent, err := regexProc.Run(srcEvent)
	messageExist, _ := expectedEvent.HasKey("message")
	assert.NoError(t, err)
	assert.True(t, messageExist)
}

func TestMultiSubmatch(t *testing.T) {
	testConfig, _ := common.NewConfigFrom(map[string]interface{}{
		"regexp":       `^(?P<remote_addr>\S+)@(?P<remote_user>\S+)@\[(?P<time_local>[^\]]+)\]@(?P<request_host>\S+)`,
		"source_field": "test_message",
	})
	regexProc, err := newParseRegexFields(*testConfig)
	assert.NoError(t, err)
	srcEvent := common.MapStr{
		"test_message": "127.0.0.1@-@[15/May/2017:17:27:01 +0800]@filebeat.lain.test",
	}
	expectedEvent, err := regexProc.Run(srcEvent)
	remoteAddr, _ := expectedEvent.GetValue("remote_addr")
	remoteUser, _ := expectedEvent.GetValue("remote_user")
	timeLocal, _ := expectedEvent.GetValue("time_local")
	requestHost, _ := expectedEvent.GetValue("request_host")
	message, _ := expectedEvent.GetValue("test_message")
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1", remoteAddr)
	assert.Equal(t, "-", remoteUser)
	assert.Equal(t, "15/May/2017:17:27:01 +0800", timeLocal)
	assert.Equal(t, "filebeat.lain.test", requestHost)
	assert.Equal(t, "127.0.0.1@-@[15/May/2017:17:27:01 +0800]@filebeat.lain.test", message)
}
