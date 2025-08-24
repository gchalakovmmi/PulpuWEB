package tts

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
)

// TTSProvider represents the available TTS providers
type TTSProvider string

const (
    ProviderKittenTTS TTSProvider = "kittentts"
    ProviderGroq      TTSProvider = "groq"
)

// TTSRequest represents a TTS request
type TTSRequest struct {
    Text           string
    Voice          string
    ResponseFormat string
    Speed          float64
    Model          string // Only for Groq
}

// TTSResponse represents a TTS response
type TTSResponse struct {
    AudioData []byte
    Error     string
}

// TTSService handles text-to-speech conversion
type TTSService struct {
    Provider TTSProvider
    BaseURL  string
    APIKey   string
}

// NewTTSService creates a new TTS service
func NewTTSService() (*TTSService, error) {
    provider := os.Getenv("TTS_PROVIDER")
    if provider == "" {
        return nil, fmt.Errorf("TTS_PROVIDER environment variable is not set")
    }

    baseURL := os.Getenv("TTS_BASE_URL")
    if baseURL == "" {
        return nil, fmt.Errorf("TTS_BASE_URL environment variable is not set")
    }

    apiKey := os.Getenv("TTS_API_KEY")

    return &TTSService{
        Provider: TTSProvider(provider),
        BaseURL:  baseURL,
        APIKey:   apiKey,
    }, nil
}

// ConvertTextToSpeech converts text to speech using the configured provider
func (ts *TTSService) ConvertTextToSpeech(req TTSRequest) (*TTSResponse, error) {
    switch ts.Provider {
    case ProviderKittenTTS:
        return ts.convertWithKittenTTS(req)
    case ProviderGroq:
        return ts.convertWithGroq(req)
    default:
        return nil, fmt.Errorf("unsupported TTS provider: %s", ts.Provider)
    }
}

// convertWithKittenTTS converts text to speech using KittenTTS
func (ts *TTSService) convertWithKittenTTS(req TTSRequest) (*TTSResponse, error) {
    url := fmt.Sprintf("%s/v1/audio/speech", ts.BaseURL)

    // Create TTS request
    ttsRequest := map[string]interface{}{
        "model":           req.Model,
        "input":           req.Text,
        "voice":           req.Voice,
        "response_format": req.ResponseFormat,
        "speed":           req.Speed,
    }

    jsonData, err := json.Marshal(ttsRequest)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal TTS request: %w", err)
    }

    // Create HTTP request
    httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to create TTS request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")
    if ts.APIKey != "" {
        httpReq.Header.Set("Authorization", "Bearer "+ts.APIKey)
    }

    // Send request
    client := &http.Client{}
    resp, err := client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("TTS request failed: %w", err)
    }
    defer resp.Body.Close()

    // Check if the response is successful
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return &TTSResponse{
            Error: fmt.Sprintf("TTS server returned error: Status %d, Body: %s", resp.StatusCode, string(body)),
        }, nil
    }

    // Read the audio data
    audioBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read TTS audio: %w", err)
    }

    return &TTSResponse{
        AudioData: audioBytes,
    }, nil
}

// convertWithGroq converts text to speech using Groq API
func (ts *TTSService) convertWithGroq(req TTSRequest) (*TTSResponse, error) {
    url := fmt.Sprintf("%s/openai/v1/audio/speech", ts.BaseURL)

    // Create TTS request
    ttsRequest := map[string]interface{}{
        "model":           req.Model,
        "input":           req.Text,
        "voice":           req.Voice,
        "response_format": req.ResponseFormat,
    }

    jsonData, err := json.Marshal(ttsRequest)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal TTS request: %w", err)
    }

    // Create HTTP request
    httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to create TTS request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")
    if ts.APIKey != "" {
        httpReq.Header.Set("Authorization", "Bearer "+ts.APIKey)
    }

    // Send request
    client := &http.Client{}
    resp, err := client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("TTS request failed: %w", err)
    }
    defer resp.Body.Close()

    // Check if the response is successful
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return &TTSResponse{
            Error: fmt.Sprintf("TTS server returned error: Status %d, Body: %s", resp.StatusCode, string(body)),
        }, nil
    }

    // Read the audio data
    audioBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read TTS audio: %w", err)
    }

    return &TTSResponse{
        AudioData: audioBytes,
    }, nil
}
