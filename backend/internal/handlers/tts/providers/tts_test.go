package providers

import (
	"testing"
)

func TestParseModelVoice(t *testing.T) {
	modelId, voiceId := ParseModelVoice("test", "default", "default-voice", nil)
	if modelId == "" {
		t.Fatal("ParseModelVoice returned empty modelId")
	}
	t.Logf("model=%s, voice=%s", modelId, voiceId)
}
