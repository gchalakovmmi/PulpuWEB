BINARY := exampleGoogle
.PHONY: clear run test

clear:
	@clear

run:
	@echo "Building binary ..."
	CGO_ENABLED=0 go build -o $(BINARY) .
	@echo "Runing $(BINARY)"
	env $$(grep -v '^#' vars.env | xargs) ./$(BINARY)

test:
	@echo "Building binary ..."
	CGO_ENABLED=0 go build -o $(BINARY) .
	@echo "Runing $(BINARY)"
	env $$(grep -v '^#' vars.env.real | xargs) ./$(BINARY)
