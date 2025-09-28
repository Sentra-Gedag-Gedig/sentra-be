package audio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TTSService struct {
	apiKey  string
	voiceID string
}

func NewTTSService(apiKey, voiceID string) *TTSService {
	return &TTSService{
		apiKey:  apiKey,
		voiceID: voiceID,
	}
}

func (tts *TTSService) GenerateAudio(text string) ([]byte, error) {
	url := "https://api.elevenlabs.io/v1/text-to-speech/" + tts.voiceID

	requestBody := map[string]interface{}{
		"text": text,
		"model_id": "eleven_multilingual_v2",
		"voice_settings": map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.8,
			"style":           0.0,
			"use_speaker_boost": true,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", tts.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ElevenLabs API error: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}