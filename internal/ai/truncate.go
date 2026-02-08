package ai

import "encoding/json"

func TruncateCV(cvData map[string]interface{}, maxTokens int) (map[string]interface{}, bool) {
	copied := deepCopy(cvData)
	if copied == nil {
		return cvData, false
	}

	if estimateTokens(copied) <= maxTokens {
		return copied, false
	}

	cv, ok := getInnerMap(copied, "curriculo-vitae")
	if !ok {
		return copied, true
	}

	removals := []string{"dados-complementares", "outra-producao", "producao-tecnica"}
	for _, key := range removals {
		delete(cv, key)
		if estimateTokens(copied) <= maxTokens {
			return copied, true
		}
	}

	truncateProdBibliografica(copied, cv, maxTokens)
	return copied, true
}

func deepCopy(src map[string]interface{}) map[string]interface{} {
	b, err := json.Marshal(src)
	if err != nil {
		return nil
	}
	var dst map[string]interface{}
	if err := json.Unmarshal(b, &dst); err != nil {
		return nil
	}
	return dst
}

func estimateTokens(data map[string]interface{}) int {
	b, err := json.Marshal(data)
	if err != nil {
		return 0
	}
	return len(b) / 4
}

func getInnerMap(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	val, exists := data[key]
	if !exists {
		return nil, false
	}
	m, ok := val.(map[string]interface{})
	return m, ok
}

func truncateProdBibliografica(root, cv map[string]interface{}, maxTokens int) {
	pb, ok := getInnerMap(cv, "producao-bibliografica")
	if !ok {
		return
	}

	arrays := collectArrays(pb)
	if len(arrays) == 0 {
		return
	}

	maxLen := maxArrayLen(arrays)
	for n := maxLen / 2; n >= 0 && estimateTokens(root) > maxTokens; n /= 2 {
		for _, entry := range arrays {
			arr := entry.slice
			if len(arr) > n {
				entry.parent[entry.key] = arr[:n]
			}
		}
		if n == 0 {
			break
		}
	}
}

type arrayRef struct {
	parent map[string]interface{}
	key    string
	slice  []interface{}
}

func collectArrays(pb map[string]interface{}) []arrayRef {
	var refs []arrayRef
	for _, v := range pb {
		inner, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		for k, iv := range inner {
			arr, ok := iv.([]interface{})
			if ok {
				refs = append(refs, arrayRef{parent: inner, key: k, slice: arr})
			}
		}
	}
	return refs
}

func maxArrayLen(refs []arrayRef) int {
	m := 0
	for _, r := range refs {
		if len(r.slice) > m {
			m = len(r.slice)
		}
	}
	return m
}
