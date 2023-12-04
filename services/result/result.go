package result

type WorkerResult struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
}

func NewWorkerResult() *WorkerResult {
	return &WorkerResult{}
}
