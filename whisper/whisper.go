package whisper

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "net/url"
    "os"
)

// TranscribeService handles audio transcription
type TranscribeService struct {
    WhisperURL string
    Provider   string
    APIKey     string // For external providers like Groq
    Model      string // For providers like Groq that require model specification
}

// NewTranscribeService creates a new transcription service
func NewTranscribeService() (*TranscribeService, error) {
    whisperURL := os.Getenv("WHISPER_URL")
    if whisperURL == "" {
        return nil, fmt.Errorf("WHISPER_URL environment variable is not set")
    }

    provider := os.Getenv("WHISPER_PROVIDER")
    if provider == "" {
        provider = "docker" // Default to docker provider
    }

    apiKey := os.Getenv("WHISPER_KEY")
    model := os.Getenv("WHISPER_MODEL")

    return &TranscribeService{
        WhisperURL: whisperURL,
        Provider:   provider,
        APIKey:     apiKey,
        Model:      model,
    }, nil
}

// TranscribeRequest represents a transcription request
type TranscribeRequest struct {
    AudioData     []byte
    FileName      string
    Language      string
    Task          string
    OutputFormat  string
    ShouldEncode  bool
    Model         string // Override the default model if needed
}

// TranscribeResponse represents a transcription response
type TranscribeResponse struct {
    Text     string    `json:"text"`
    Segments []Segment `json:"segments"`
    Language string    `json:"language"`
    Error    string    `json:"error,omitempty"`
}

// Segment represents a segment of the transcribed text
type Segment struct {
    ID               int     `json:"id"`
    Seek             int     `json:"seek"`
    Start            float64 `json:"start"`
    End              float64 `json:"end"`
    Text             string  `json:"text"`
    Tokens           []int   `json:"tokens"`
    Temperature      float64 `json:"temperature"`
    AvgLogprob       float64 `json:"avg_logprob"`
    CompressionRatio float64 `json:"compression_ratio"`
    NoSpeechProb     float64 `json:"no_speech_prob"`
}

// ParseAudioFromRequest extracts audio data from an HTTP request
func ParseAudioFromRequest(r *http.Request) ([]byte, string, error) {
    if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
        return nil, "", fmt.Errorf("failed to parse multipart form: %w", err)
    }

    file, header, err := r.FormFile("audio")
    if err != nil {
        return nil, "", fmt.Errorf("failed to get audio file from form: %w", err)
    }
    defer file.Close()

    var buf bytes.Buffer
    if _, err := io.Copy(&buf, file); err != nil {
        return nil, "", fmt.Errorf("failed to read audio file: %w", err)
    }

    return buf.Bytes(), header.Filename, nil
}

// SendToWhisper forwards audio data to the appropriate service
func (ts *TranscribeService) SendToWhisper(req *TranscribeRequest) (*TranscribeResponse, error) {
    switch ts.Provider {
    case "groq":
        return ts.sendToGroq(req)
    case "docker":
        fallthrough
    default:
        return ts.sendToDocker(req)
    }
}

// sendToDocker sends request to the Docker container
func (ts *TranscribeService) sendToDocker(req *TranscribeRequest) (*TranscribeResponse, error) {
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    part, err := writer.CreateFormFile("audio_file", req.FileName)
    if err != nil {
        return nil, fmt.Errorf("failed to create form file: %w", err)
    }

    if _, err := io.Copy(part, bytes.NewReader(req.AudioData)); err != nil {
        return nil, fmt.Errorf("failed to write audio data to form: %w", err)
    }

    if err := writer.Close(); err != nil {
        return nil, fmt.Errorf("failed to close multipart writer: %w", err)
    }

    // Build the URL with query parameters
    whisperURL := fmt.Sprintf("%s?encode=%t&task=%s&language=%s&output=%s",
        ts.WhisperURL, req.ShouldEncode, req.Task, req.Language, req.OutputFormat)

    httpReq, err := http.NewRequest("POST", whisperURL, body)
    if err != nil {
        return nil, fmt.Errorf("failed to create HTTP request: %w", err)
    }

    httpReq.Header.Set("Content-Type", writer.FormDataContentType())

    client := &http.Client{}
    resp, err := client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("failed to send request to Whisper: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("Whisper service returned non-OK status: %s", resp.Status)
    }

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    var result TranscribeResponse
    if err := json.Unmarshal(respBody, &result); err != nil {
        return nil, fmt.Errorf("failed to parse Whisper response: %w", err)
    }

    return &result, nil
}

// sendToGroq sends request to Groq API
func (ts *TranscribeService) sendToGroq(req *TranscribeRequest) (*TranscribeResponse, error) {
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // Use the model from the service (set via env) unless overridden in the request
    model := ts.Model
    if req.Model != "" {
        model = req.Model
    }
    
    if model == "" {
        return nil, fmt.Errorf("model is required for Groq provider")
    }
    
    // Add model parameter
    if err := writer.WriteField("model", model); err != nil {
        return nil, fmt.Errorf("failed to write model field: %w", err)
    }

    // Add response format
    responseFormat := "json"
    if req.OutputFormat == "verbose_json" {
        responseFormat = "verbose_json"
    }
    if err := writer.WriteField("response_format", responseFormat); err != nil {
        return nil, fmt.Errorf("failed to write response_format field: %w", err)
    }

    // Add language if specified
    if req.Language != "" && req.Language != "auto" {
        if err := writer.WriteField("language", req.Language); err != nil {
            return nil, fmt.Errorf("failed to write language field: %w", err)
        }
    }

    // Add audio file
    part, err := writer.CreateFormFile("file", req.FileName)
    if err != nil {
        return nil, fmt.Errorf("failed to create form file: %w", err)
    }

    if _, err := io.Copy(part, bytes.NewReader(req.AudioData)); err != nil {
        return nil, fmt.Errorf("failed to write audio data to form: %w", err)
    }

    if err := writer.Close(); err != nil {
        return nil, fmt.Errorf("failed to close multipart writer: %w", err)
    }

    httpReq, err := http.NewRequest("POST", ts.WhisperURL, body)
    if err != nil {
        return nil, fmt.Errorf("failed to create HTTP request: %w", err)
    }

    httpReq.Header.Set("Content-Type", writer.FormDataContentType())
    
    if ts.APIKey == "" {
        return nil, fmt.Errorf("API key is required for Groq provider")
    }
    httpReq.Header.Set("Authorization", "Bearer "+ts.APIKey)

    client := &http.Client{}
    resp, err := client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("failed to send request to Groq: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("Groq service returned non-OK status: %s - %s", resp.Status, string(body))
    }

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    // Parse Groq response
    var groqResponse struct {
        Text     string `json:"text"`
        Language string `json:"language"`
    }

    if err := json.Unmarshal(respBody, &groqResponse); err != nil {
        return nil, fmt.Errorf("failed to parse Groq response: %w", err)
    }

    // Convert to our standard response format
    result := &TranscribeResponse{
        Text:     groqResponse.Text,
        Language: groqResponse.Language,
    }

    return result, nil
}

// GetWhisperURL returns the Whisper URL with default values if not set
func GetWhisperURL() (string, error) {
    whisperURL := os.Getenv("WHISPER_URL")
    if whisperURL == "" {
        return "", fmt.Errorf("WHISPER_URL environment variable is not set")
    }

    // Parse URL to ensure it's valid
    _, err := url.Parse(whisperURL)
    if err != nil {
        return "", fmt.Errorf("invalid WHISPER_URL: %w", err)
    }

    return whisperURL, nil
}
