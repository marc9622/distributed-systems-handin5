# distributed-systems-handin5

To run the program, make an exe file

```go build -o program.exe src\program.go```

Then run nodes with added ports. The first number is the node's port, and the following numbers are all of the other ports. Example:

```program.exe -node -port 1111 1111 2222```

```program.exe -node -port 2222 1111 2222```


And then run a client

```program.exe 1111 2222```

From there on the client can run 
```
  - bid int

    Bid a number

  - result

    Gets the highest bid

  - end

    Ends the program

```
The logs can be found in the log folder
