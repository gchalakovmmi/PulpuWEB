package main

import (
    "encoding/base64"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/gchalakovmmi/PulpuWEB/tts"
)

func main() {
    // Initialize TTS service
    ttsService, err := tts.NewTTSService()
    if err != nil {
        log.Fatal("Failed to initialize TTS service:", err)
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>TTS Example</title>
        </head>
        <body>
            <h1>TTS Example</h1>
            <form action="/tts" method="post">
                <label for="text">Text to convert:</label><br>
                <textarea id="text" name="text" rows="4" cols="50">Hello, this is a test message.</textarea><br>
                <input type="submit" value="Convert to Speech">
            </form>
        </body>
        </html>
        `
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprint(w, html)
    })

    http.HandleFunc("/tts", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        text := r.FormValue("text")
        if text == "" {
            http.Error(w, "Text is required", http.StatusBadRequest)
            return
        }

        // Create TTS request
        req := tts.TTSRequest{
            Text:           text,
            Voice:          os.Getenv("TTS_VOICE"),
            ResponseFormat: os.Getenv("TTS_RESPONSE_FORMAT"),
            Model:          os.Getenv("TTS_MODEL"),
        }

        if speed := os.Getenv("TTS_SPEED"); speed != "" {
            fmt.Sscanf(speed, "%f", &req.Speed)
        }

        // Convert text to speech
        resp, err := ttsService.ConvertTextToSpeech(req)
        if err != nil {
            http.Error(w, "TTS conversion failed: "+err.Error(), http.StatusInternalServerError)
            return
        }

        if resp.Error != "" {
            http.Error(w, "TTS error: "+resp.Error, http.StatusInternalServerError)
            return
        }

        // Encode audio to base64 for HTML response
        audioBase64 := base64.StdEncoding.EncodeToString(resp.AudioData)

        html := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>TTS Result</title>
        </head>
        <body>
            <h1>TTS Result</h1>
            <p>Converted text: %s</p>
            <audio controls>
                <source src="data:audio/mp3;base64,%s" type="audio/mp3">
                Your browser does not support the audio element.
            </audio>
            <br>
            <a href="/">Convert another text</a>
        </body>
        </html>
        `, text, audioBase64)

        w.Header().Set("Content-Type", "text/html")
        fmt.Fprint(w, html)
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server running on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
