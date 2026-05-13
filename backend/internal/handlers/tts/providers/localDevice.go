// ref: _ref/9router/open-sse/handlers/ttsProviders/localDevice.js
package providers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func init() {
	Register("local-device", &LocalDeviceProvider{})
}

// LocalDeviceProvider implements TTS using local device.
// ref: _ref/9router/open-sse/handlers/ttsProviders/localDevice.js
type LocalDeviceProvider struct{}

func (p *LocalDeviceProvider) Synthesize(ctx context.Context, text string, model string, creds *Credentials) (*TTSResult, error) {
	if runtime.GOOS == "darwin" {
		return p.synthesizeMac(ctx, text, model)
	} else if runtime.GOOS == "windows" {
		return p.synthesizeWindows(ctx, text, model)
	}
	return nil, fmt.Errorf("local device TTS not supported on %s", runtime.GOOS)
}

func (p *LocalDeviceProvider) synthesizeMac(ctx context.Context, text, voice string) (*TTSResult, error) {
	tmpDir, err := os.MkdirTemp("", "tts-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	aiffPath := filepath.Join(tmpDir, "out.aiff")
	mp3Path := filepath.Join(tmpDir, "out.mp3")

	args := []string{"-o", aiffPath, text}
	if voice != "" {
		args = []string{"-v", voice, "-o", aiffPath, text}
	}

	cmd := exec.CommandContext(ctx, "say", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("say command failed: %w, output: %s", err, string(output))
	}

	cmd = exec.CommandContext(ctx, "ffmpeg", "-y", "-i", aiffPath, "-codec:a", "libmp3lame", "-qscale:a", "4", mp3Path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg command failed: %w, output: %s", err, string(output))
	}

	audio, err := os.ReadFile(mp3Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	return &TTSResult{
		Base64: base64.StdEncoding.EncodeToString(audio),
		Format: "mp3",
	}, nil
}

func (p *LocalDeviceProvider) synthesizeWindows(ctx context.Context, text, voice string) (*TTSResult, error) {
	tmpDir, err := os.MkdirTemp("", "tts-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	wavPath := filepath.Join(tmpDir, "out.wav")
	mp3Path := filepath.Join(tmpDir, "out.mp3")

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Speech;
$s = New-Object System.Speech.Synthesis.SpeechSynthesizer;
%s
$s.SetOutputToWaveFile("%s");
$s.Speak("%s");
$s.Dispose();
`, func() string {
		if voice != "" {
			return fmt.Sprintf("$s.SelectVoice('%s');", voice)
		}
		return ""
	}(), wavPath, strings.ReplaceAll(text, "'", "''"))

	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", psScript)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("powershell command failed: %w, output: %s", err, string(output))
	}

	cmd = exec.CommandContext(ctx, "ffmpeg", "-y", "-i", wavPath, "-codec:a", "libmp3lame", "-qscale:a", "4", mp3Path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg command failed: %w, output: %s", err, string(output))
	}

	audio, err := os.ReadFile(mp3Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	return &TTSResult{
		Base64: base64.StdEncoding.EncodeToString(audio),
		Format: "mp3",
	}, nil
}

func (p *LocalDeviceProvider) FetchVoices(ctx context.Context, creds *Credentials) ([]Voice, error) {
	if runtime.GOOS == "darwin" {
		return p.fetchVoicesMac(ctx)
	} else if runtime.GOOS == "windows" {
		return p.fetchVoicesWindows(ctx)
	}
	return nil, nil
}

func (p *LocalDeviceProvider) fetchVoicesMac(ctx context.Context) ([]Voice, error) {
	cmd := exec.CommandContext(ctx, "say", "-v", "?")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get voices: %w", err)
	}

	var voices []Voice
	lines := strings.Split(string(output), "\n")
	re := regexp.MustCompile(`^([^\s].*?)\s{2,}([a-z]{2}_[A-Z]{2})`)

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		name := strings.TrimSpace(matches[1])
		locale := strings.TrimSpace(matches[2])
		parts := strings.Split(locale, "_")
		lang := ""
		if len(parts) >= 1 {
			lang = parts[0]
		}

		voices = append(voices, Voice{
			ID:     name,
			Name:   name,
			Lang:   lang,
			Locale: locale,
			Gender: "",
		})
	}

	return voices, nil
}

func (p *LocalDeviceProvider) fetchVoicesWindows(ctx context.Context) ([]Voice, error) {
	psScript := `
Add-Type -AssemblyName System.Speech;
$s = New-Object System.Speech.Synthesis.SpeechSynthesizer;
$s.GetInstalledVoices() | ForEach-Object { $v = $_.VoiceInfo;
[PSCustomObject]@{ Name=$v.Name; Culture=$v.Culture.Name; Gender=$v.Gender } }
| ConvertTo-Json -Compress
`

	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", psScript)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get voices: %w", err)
	}

	var rawVoices []struct {
		Name    string `json:"Name"`
		Culture string `json:"Culture"`
		Gender  int    `json:"Gender"`
	}
	if err := json.Unmarshal(output, &rawVoices); err != nil {
		if len(output) > 0 && output[0] != '[' {
			var single struct {
				Name    string `json:"Name"`
				Culture string `json:"Culture"`
				Gender  int    `json:"Gender"`
			}
			if err := json.Unmarshal(output, &single); err != nil {
				return nil, err
			}
			rawVoices = []struct {
				Name    string `json:"Name"`
				Culture string `json:"Culture"`
				Gender  int    `json:"Gender"`
			}{single}
		} else {
			return nil, err
		}
	}

	voices := make([]Voice, len(rawVoices))
	for i, v := range rawVoices {
		culture := v.Culture
		if culture == "" {
			culture = "en-US"
		}
		parts := strings.Split(culture, "-")
		lang := ""
		if len(parts) >= 1 {
			lang = parts[0]
		}

		gender := ""
		switch v.Gender {
		case 1:
			gender = "Male"
		case 2:
			gender = "Female"
		}

		voices[i] = Voice{
			ID:     v.Name,
			Name:   v.Name,
			Lang:   lang,
			Locale: strings.ReplaceAll(culture, "-", "_"),
			Gender: gender,
		}
	}

	return voices, nil
}
