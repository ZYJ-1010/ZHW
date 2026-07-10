package appapi

type batchMutationResult struct {
	ID      int64  `json:"id"`
	Success bool   `json:"success"`
	Status  string `json:"status,omitempty"`
	Error   string `json:"error,omitempty"`
}

func uniquePositiveIDs(ids []int64) []int64 {
	seen := make(map[int64]bool, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}
