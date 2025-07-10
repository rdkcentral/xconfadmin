package shared

import (
	"testing"

	"gotest.tools/assert"
)

func TestNormalizeContext(t *testing.T) {
	contextMap := map[string]string{}

	// normalize empty contextMap, no change
	NormalizeCommonContext(contextMap, "estMacAddress", "ecmMacAddress")
	assert.Equal(t, len(contextMap), 0)

	// normalize invalid mac addresses, no change
	contextMap["estbMacAddress"] = "estbMacAddress"
	contextMap["ecmMacAddress"] = "ecmMacAddress"

	NormalizeCommonContext(contextMap, "estMacAddress", "ecmMacAddress")
	assert.Equal(t, len(contextMap), 2)
	assert.Equal(t, contextMap["estbMacAddress"], "estbMacAddress")
	assert.Equal(t, contextMap["ecmMacAddress"], "ecmMacAddress")

	// normalize expected values
	contextMap["estbMacAddress"] = "00:0a:95:9d:68:16"
	contextMap["ecmMacAddress"] = "00:0a:95:9d:68:17"
	contextMap["model"] = "model"
	contextMap["env"] = "env"
	contextMap["partnerId"] = "partnerId"
	NormalizeCommonContext(contextMap, "estbMacAddress", "ecmMacAddress")
	assert.Equal(t, len(contextMap), 5)
	assert.Equal(t, contextMap["estbMacAddress"], "00:0A:95:9D:68:16")
	assert.Equal(t, contextMap["ecmMacAddress"], "00:0A:95:9D:68:17")
	assert.Equal(t, contextMap["model"], "MODEL")
	assert.Equal(t, contextMap["env"], "ENV")
	assert.Equal(t, contextMap["partnerId"], "PARTNERID")

	// normalize where keys don't match what's passed in, no change in those values
	contextMap = map[string]string{}
	contextMap["estbMacAddress"] = "00:0a:95:9d:68:16"
	contextMap["ecmMacAddress"] = "00:0a:95:9d:68:17"
	contextMap["model"] = "model"
	contextMap["env"] = "env"
	contextMap["partnerId"] = "partnerId"
	NormalizeCommonContext(contextMap, "eStbMac", "eCMMac")
	assert.Equal(t, len(contextMap), 5)
	assert.Equal(t, contextMap["estbMacAddress"], "00:0a:95:9d:68:16")
	assert.Equal(t, contextMap["ecmMacAddress"], "00:0a:95:9d:68:17")
	assert.Equal(t, contextMap["model"], "MODEL")
	assert.Equal(t, contextMap["env"], "ENV")
	assert.Equal(t, contextMap["partnerId"], "PARTNERID")
}
