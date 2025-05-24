# slicealias

⚠️ A static analyzer for Go that detects dangerous slice aliasing with unknown capacity.

## What It Does

This analyzer finds cases where a slice is aliased (e.g., `b := a`) and both are independently appended to without ensuring enough capacity. This can lead to bugs due to unexpected memory divergence.

### Example

```go
func broken() {
    a := make([]int, 0) // unknown capacity
    b := a              // alias

    b = append(b, 1)    // Warning: alias of unknown-cap slice
    a = append(a, 2)    // fine
}
```


## Usage
```bash
go install github.com/nitinmewar/alias-analyser@latest

# or clone and run manually
go run . ./slicetest
```

## Development

Run tests with : 
```bash
go test ./...
```

## Status
This is a WIP and not yet part of golang.org/x/tools/go/analysis/passes.
