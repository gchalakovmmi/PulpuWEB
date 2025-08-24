BINARY := exampleTTS

.PHONY: clear run

clear:
		@clear

run:
		@echo "Building binary ..."
		CGO_ENABLED=0 go build -o $(BINARY) .
		@echo "Running $(BINARY)"
		env $$(grep -v '^#' vars.env | xargs) ./$(BINARY)
