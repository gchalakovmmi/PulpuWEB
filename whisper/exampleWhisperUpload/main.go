package main

import (
    "fmt"
    "html/template"
    "log"
    "net/http"

    "github.com/gchalakovmmi/PulpuWEB/whisper"
)

var (
    whisperService *whisper.TranscribeService
)

func main() {
    // Initialize whisper service
    var err error
    whisperService, err = whisper.NewTranscribeService()
    if err != nil {
        log.Fatal("Failed to initialize whisper service:", err)
    }

    // Setup routes
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/transcribe", transcribeHandler)
    http.HandleFunc("/result", resultHandler)

    log.Println("Server running on :8000")
    log.Fatal(http.ListenAndServe(":8000", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    tmpl := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Audio Transcription</title>
    </head>
    <body>
        <div class="container">
            <h1>Audio Transcription</h1>
            <form action="/transcribe" method="post" enctype="multipart/form-data">
                <div class="form-group">
                    <label for="audioFile">Select an audio file to transcribe:</label>
                    <input type="file" id="audioFile" name="audio" accept="audio/*" required>
                </div>
                <button type="submit">Transcribe Audio</button>
            </form>
            <p class="note">Supported formats: WAV, MP3, FLAC, and other common audio formats</p>
        </div>
    </body>
    </html>
    `
    w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, tmpl)
}

func transcribeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    // Parse the uploaded file
    audioData, filename, err := whisper.ParseAudioFromRequest(r)
    if err != nil {
        http.Error(w, "Failed to parse audio: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Prepare the transcription request
    req := &whisper.TranscribeRequest{
        AudioData:    audioData,
        FileName:     filename,
        Language:     "en",
        Task:         "transcribe",
        OutputFormat: "json",
        ShouldEncode: true,
    }

    // Send to Whisper service
    result, err := whisperService.SendToWhisper(req)
    if err != nil {
        http.Error(w, "Transcription failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Redirect to result page with the transcribed text
    http.Redirect(w, r, "/result?text="+result.Text, http.StatusSeeOther)
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
    tmpl := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Transcription Result</title>
    </head>
    <body>
        <div class="container">
            <h1>Transcription Result</h1>
            <div class="result">{{.}}</div>
            <a href="/" class="btn">Upload Another File</a>
        </div>
    </body>
    </html>
    `

    text := r.URL.Query().Get("text")
    t := template.Must(template.New("result").Parse(tmpl))
    t.Execute(w, text)
}
