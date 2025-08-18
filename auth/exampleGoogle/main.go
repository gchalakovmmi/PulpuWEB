package main

import (
	"os"
	"fmt"
	"log"
	"net/http"
	"io"
	"context"
	"time"

	"github.com/a-h/templ"
	"github.com/gchalakovmmi/PulpuWEB/auth"
)

func main() {
	// Initialize authentication
	authConfig, err := auth.GetGoogleAuthConfig()
	if err != nil {
		log.Fatalf("Error getting Google auth config: %v", err)
	}
	
	googleAuth := auth.NewGoogleAuth(authConfig)

	// Root handler - shows login link
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, `
		<html>
		<head><title>Login</title></head>
		<body>
			<h1>Welcome!</h1>
			<a href="/auth/google">Login with Google</a>
		</body>
		</html>`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Start Google authentication flow
	http.HandleFunc("/auth/google", func(w http.ResponseWriter, r *http.Request) {
		if _, err := googleAuth.GetSession(r); err == nil {
			http.Redirect(w, r, "/protected", http.StatusSeeOther)
			return
		}
		googleAuth.BeginAuthHandler(w, r)
	})

	// Google callback handler
	http.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		user, err := googleAuth.CompleteUserAuth(w, r)
		if err != nil {
			http.Error(w, "Authentication failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := googleAuth.StoreSession(w, user); err != nil {
			http.Error(w, "Session creation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/protected", http.StatusSeeOther)
	})

	// Logout handler
	http.HandleFunc("/logout/google", func(w http.ResponseWriter, r *http.Request) {
		googleAuth.LogoutHandler(w, r)
		googleAuth.ClearSession(w)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})

	// Protected route handler
	protectedHandler := func(w http.ResponseWriter, r *http.Request) {
		session, err := googleAuth.GetSession(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Render protected content
		comp := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			user := session.User
			_, err := io.WriteString(w, `
        <html>
        <head><title>Protected Page</title></head>
        <body>
            <h1>User Information</h1>
            <p>Name: `+user.Name+`</p>
            <p>Email: `+user.Email+`</p>
            <p>NickName: `+user.NickName+`</p>
            <p>Location: `+user.Location+`</p>
            <p>AvatarURL: `+user.AvatarURL+`</p>
            <p>Description: `+user.Description+`</p>
            <p>UserID: `+user.UserID+`</p>
            <p>AccessToken: `+user.AccessToken+`</p>
            <p>AccessTokenSecret: `+user.AccessTokenSecret+`</p>
            <p>RefreshToken: `+user.RefreshToken+`</p>
            <p>ExpiresAt: `+user.ExpiresAt.Format(time.RFC3339)+`</p>
            <p><img src="`+user.AvatarURL+`" width="100" referrerpolicy="no-referrer"></p>
            <a href="/logout/google">Logout</a>
        </body>
        </html>`)
			return err
		})

		templ.Handler(comp).ServeHTTP(w, r)
	}

	// Apply auth middleware to protected route
	http.HandleFunc("/protected", googleAuth.WithGoogleAuth(protectedHandler))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	fmt.Printf("Serving on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
