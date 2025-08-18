module GoogleAuthExample

go 1.24.4

replace github.com/gchalakovmmi/PulpuWEB/auth => ..

replace github.com/gchalakovmmi/PulpuWEB/db => ../../db

require (
	github.com/a-h/templ v0.3.943
	github.com/gchalakovmmi/PulpuWEB/auth v0.0.0
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/go-chi/chi/v5 v5.2.2 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.2 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/sessions v1.1.1 // indirect
	github.com/markbates/goth v1.82.0 // indirect
	golang.org/x/oauth2 v0.27.0 // indirect
)
