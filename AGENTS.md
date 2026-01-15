# Cursor Agent Rules

## Code Style

### Comments
- Avoid redundant comments that don't add information beyond what the code already expresses
- Remove comments that merely restate what the code does
- Keep comments that explain *why* something is done, not *what* is being done
- Keep comments that provide important context or warnings

Examples of comments to avoid:
```go
// Get metadata
metadata := GetMetadata()

// Loop through items
for _, item := range items {
```

Examples of useful comments:
```go
// Still update metadata even on error to track status
if result != nil {
    UpdateMetadata(result)
}

// TODO: This workaround is needed until upstream bug #123 is fixed
someWorkaround()
```
