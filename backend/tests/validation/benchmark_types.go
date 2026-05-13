package validation

type BenchmarkResult struct {
	Name        string  `json:"name"`
	Iterations  int     `json:"iterations"`
	NsPerOp     float64 `json:"ns_per_op"`
	BytesPerOp  int64   `json:"bytes_per_op"`
	AllocsPerOp int64   `json:"allocs_per_op"`
}

func RunBenchmarks() ([]BenchmarkResult, error) {
	results := []BenchmarkResult{
		{Name: "ChatCompletionLatency", Iterations: 100},
		{Name: "StreamingLatency", Iterations: 50},
		{Name: "EmbeddingLatency", Iterations: 100},
		{Name: "ModelsListLatency", Iterations: 1000},
		{Name: "ConcurrentRequests", Iterations: 100},
	}
	return results, nil
}
