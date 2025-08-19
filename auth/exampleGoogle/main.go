package main

import (
	"os"
	"fmt"
	"log"
	"net/http"
	"io"
	"context"

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
	http.HandleFunc("/", googleAuth.WithOutGoogleAuth("/protected", func(w http.ResponseWriter, r *http.Request) {
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
	}))

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

	protectedHandler := func(w http.ResponseWriter, r *http.Request) {
			// Get session from context
			session, ok := r.Context().Value("user_session").(*auth.Session)
			if !ok {
					http.Error(w, "Session invalid", http.StatusUnauthorized)
					return
			}

			// Render all user details
			user := session.User
			comp := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
				_, err := io.WriteString(w, `
<html>
<head><title>User Info</title></head>
<body>
	<img src="`+user.AvatarURL+`" width="80">
	<pre>
Name:          `+user.Name+`
Email:         `+user.Email+`
NickName:      `+user.NickName+`
Location:      `+user.Location+`
Description:   `+user.Description+`
UserID:        `+user.UserID+`
Provider:      `+user.Provider+`
AccessToken:   `+user.AccessToken+`
RefreshToken:  `+user.RefreshToken+`
ExpiresAt:     `+user.ExpiresAt.Format("2006-01-02 15:04")+`
RawData:       `+fmt.Sprint(user.RawData)+`
	</pre>
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
