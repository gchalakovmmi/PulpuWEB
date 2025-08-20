package whisper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// TranscribeService handles audio transcription
type TranscribeService struct {
	WhisperURL string
}

// NewTranscribeService creates a new transcription service by reading the URL from environment variables
func NewTranscribeService() (*TranscribeService, error) {
	whisperURL := os.Getenv("WHISPER_URL")
	if whisperURL == "" {
		return nil, fmt.Errorf("WHISPER_URL environment variable is not set")
	}
	
	return &TranscribeService{
		WhisperURL: whisperURL,
	}, nil
}

// GetWhisperURL returns the Whisper URL with default values if not set
func GetWhisperURL() (string, error) {
	whisperURL := os.Getenv("WHISPER_URL")
	if whisperURL == "" {
		return "", fmt.Errorf("WHISPER_URL environment variable is not set")
	}
	
	// You could add default values for other parameters here if needed
	return whisperURL, nil
}

// TranscribeRequest represents a transcription request
type TranscribeRequest struct {
	AudioData     []byte
	FileName      string
	Language      string
	Task          string
	OutputFormat  string
	ShouldEncode  bool
}

// TranscribeResponse represents a transcription response
type TranscribeResponse struct {
	Text      string      `json:"text"`
	Segments  []Segment   `json:"segments"`
	Language  string      `json:"language"`
	Error     string      `json:"error,omitempty"`
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

// SendToWhisper forwards audio data to the Whisper service
func (ts *TranscribeService) SendToWhisper(req *TranscribeRequest) (*TranscribeResponse, error) {
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
