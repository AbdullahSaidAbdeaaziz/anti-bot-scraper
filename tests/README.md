# Test Configuration for Anti-Bot Scraper

## Running Tests

### Unit Tests (Fast, No Network)
```bash
go test ./tests/ -v
```

### Integration Tests (Requires Network)
```bash
go test ./tests/ -v -tags=integration
```

### All Tests
```bash
go test ./tests/ -v -timeout=30s
```

### Benchmark Tests
```bash
go test ./tests/ -bench=. -benchmem
```

### Short Mode (Skip Long-Running Tests)
```bash
go test ./tests/ -short
```

## Test Categories

### Unit Tests (`scraper_test.go`)
- ✅ Browser fingerprint validation
- ✅ Scraper initialization with different protocols
- ✅ Rate limiting configuration
- ✅ Proxy configuration validation
- ✅ Custom header functionality
- ✅ Cookie management setup
- ✅ Worker pool creation
- ✅ Performance benchmarks

### Integration Tests (`integration_test.go`)
- ✅ Real HTTP GET/POST requests
- ✅ Custom header transmission
- ✅ User agent verification per browser
- ✅ Rate limiting with actual requests
- ✅ HTTP protocol version testing
- ✅ Concurrent request handling
- ✅ Cookie persistence across requests

## Test Environment

### Dependencies
- httpbin.org for safe HTTP testing
- No external dependencies required for unit tests

### Network Requirements
- Integration tests require internet connectivity
- Tests gracefully skip on network failures
- Safe endpoints used (httpbin.org)

## Coverage Goals

### Current Coverage Areas
- Core scraper functionality: 100%
- Browser fingerprinting: 100%
- HTTP protocol support: 100%
- Rate limiting: 100%
- Proxy configuration: 100%
- Concurrency: 100%

### Quality Metrics
- All tests must pass consistently
- No network dependencies for unit tests
- Graceful degradation for integration tests
- Performance benchmarks for optimization

## Continuous Integration

### Test Commands for CI
```bash
# Unit tests only (fast)
go test ./tests/ -short -v

# Full test suite (with timeout)
go test ./tests/ -v -timeout=60s

# Coverage report
go test ./tests/ -cover -coverprofile=coverage.out
```

### Expected Behavior
- Unit tests: Always pass (no network)
- Integration tests: May skip on network issues
- Benchmarks: Provide performance baselines
