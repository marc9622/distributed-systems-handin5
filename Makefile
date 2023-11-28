
bin/program: src/program.go genProtoc
	@echo "Building program.go..."
	@go build -o bin/program src/program.go
	@echo "program.go built."

build: bin/program

run: bin/program
	@echo "Running program..."
	@./bin/program

genProtoc: src/proto/program.proto
	@echo "Generating proto files..."
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative src/proto/program.proto
	@echo "Proto files generated."

.PHONY: build run build genProtoc

