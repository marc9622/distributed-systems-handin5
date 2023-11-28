# distributed-systems-handin5

To run the program, make an exe file
```go build -o program.exe src\program.go```

Then run nodes with added ports. Example:
```program.exe -node -port 1111 1111 2222```
```program.exe -node -port 2222 1111 2222```

And then run a client
```program.exe 1111 2222```
