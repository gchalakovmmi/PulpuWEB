package main

import (
    "encoding/json"
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
    http.HandleFunc("/record", recordHandler)
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
        <title>Audio Recorder</title>
    </head>
    <body>
        <h1>Audio Recorder</h1>
        <button id="recordButton">Record</button>
        <button id="stopButton" disabled>Stop</button>
        <script>
            let mediaRecorder;
            let audioChunks = [];
            
            document.getElementById('recordButton').addEventListener('click', async () => {
                const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
                mediaRecorder = new MediaRecorder(stream);
                mediaRecorder.start();
                
                document.getElementById('recordButton').disabled = true;
                document.getElementById('stopButton').disabled = false;
                
                mediaRecorder.addEventListener('dataavailable', event => {
                    audioChunks.push(event.data);
                });
            });
            
            document.getElementById('stopButton').addEventListener('click', () => {
                mediaRecorder.stop();
                document.getElementById('recordButton').disabled = false;
                document.getElementById('stopButton').disabled = true;
                
                mediaRecorder.addEventListener('stop', () => {
                    const audioBlob = new Blob(audioChunks);
                    const formData = new FormData();
                    formData.append('audio', audioBlob, 'recording.wav');
                    
                    fetch('/transcribe', {
                        method: 'POST',
                        body: formData
                    }).then(response => response.json())
                    .then(data => {
                        window.location.href = '/result?text=' + encodeURIComponent(data.text);
                    });
                    
                    audioChunks = [];
                });
            });
        </script>
    </body>
    </html>
    `
    w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, tmpl)
}

func recordHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "record.html")
}

func transcribeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    audioData, filename, err := whisper.ParseAudioFromRequest(r)
    if err != nil {
        http.Error(w, "Failed to parse audio: "+err.Error(), http.StatusBadRequest)
        return
    }

    req := &whisper.TranscribeRequest{
        AudioData:    audioData,
        FileName:     filename,
        Language:     "en",
        Task:         "transcribe",
        OutputFormat: "json",
        ShouldEncode: true,
        Model:        "", // Use the model from environment variables
    }

    result, err := whisperService.SendToWhisper(req)
    if err != nil {
        http.Error(w, "Transcription failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
    tmpl := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Transcription Result</title>
    </head>
    <body>
        <h1>Transcription Result</h1>
        <p>{{.}}</p>
        <a href="/">Record Another</a>
    </body>
    </html>
    `
    text := r.URL.Query().Get("text")
    t := template.Must(template.New("result").Parse(tmpl))
    t.Execute(w, text)
}
