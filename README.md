# ONEID - Distributable Unique numeric IDS


## Features Summary
* Thread-safe concurrent Numeric IDs.
* Support upto 1024 servers with upto 32 processes each by default.
* Partially-sortable time-based IDs.
* Trivially customizable to support even more.
* Uses only builtin Golang stdlib with no external dependencies.
* Fully testable.
* go.mod support.

## You will love it if:
* You believe numbers are IDs :)
* You believe IDs creation should be wired to the app with no dependencies on any external ID servers.


## Installation
* The usal way, get it:
```go get github.com/coderme/oneid````


## Usage
* Generation of int64 Id on a simple server with single process is as simple as:
```
id := oneid.Int64(1,1 &oneid.DefaultInt64Config)
```
* or better an uint64 if your database supports it
```
id := oneid.Uint64(1,1 &oneid.DefaultUint64Config)
```

## Advanced Usage 
* Creating a custom config for support of upto 16,384 servers with 32 processes each.
```

conf := oneid.NewUint64Config(14, 5, 20)
id, err := EnvUint64()
if err != nil {
   // deal with the error
}

````


