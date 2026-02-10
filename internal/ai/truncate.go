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
	return len(b) / 3
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

func estimateTokensAny(data interface{}) int {
	b, err := json.Marshal(data)
	if err != nil {
		return 0
	}
	return len(b) / 3
}

func deepCopyAny(src interface{}) interface{} {
	b, err := json.Marshal(src)
	if err != nil {
		return src
	}
	var dst interface{}
	if err := json.Unmarshal(b, &dst); err != nil {
		return src
	}
	return dst
}

func removeFieldFromCV(cv map[string]interface{}, field string) {
	inner, ok := cv["curriculo-vitae"]
	if !ok {
		return
	}
	cvMap, ok := inner.(map[string]interface{})
	if !ok {
		return
	}
	delete(cvMap, field)

	dg, ok := cvMap["dados-gerais"]
	if !ok {
		return
	}
	dgMap, ok := dg.(map[string]interface{})
	if !ok {
		return
	}
	delete(dgMap, field)
}

func TruncateAnalysisData(currentCV map[string]interface{}, otherCVs []map[string]interface{}, maxTokens int) (string, bool) {
	// Truncate the main CV first (use half the budget for it)
	currentCopy, mainTruncated := TruncateCV(currentCV, maxTokens/2)

	othersCopy := make([]interface{}, len(otherCVs))
	for i, cv := range otherCVs {
		othersCopy[i] = deepCopyAny(cv)
	}

	combined := map[string]interface{}{
		"pesquisador_alvo":     currentCopy,
		"outros_pesquisadores": othersCopy,
	}

	if estimateTokensAny(combined) <= maxTokens {
		b, _ := json.Marshal(combined)
		return string(b), mainTruncated
	}

	// Step 1: remove producao-bibliografica from others
	for i, cv := range othersCopy {
		cvMap, ok := cv.(map[string]interface{})
		if !ok {
			continue
		}
		removeFieldFromCV(cvMap, "producao-bibliografica")
		othersCopy[i] = cvMap
	}
	combined["outros_pesquisadores"] = othersCopy

	if estimateTokensAny(combined) <= maxTokens {
		b, _ := json.Marshal(combined)
		return string(b), true
	}

	// Step 2: remove atuacoes-profissionais from others
	for i, cv := range othersCopy {
		cvMap, ok := cv.(map[string]interface{})
		if !ok {
			continue
		}
		removeFieldFromCV(cvMap, "atuacoes-profissionais")
		othersCopy[i] = cvMap
	}
	combined["outros_pesquisadores"] = othersCopy

	if estimateTokensAny(combined) <= maxTokens {
		b, _ := json.Marshal(combined)
		return string(b), true
	}

	// Step 3: remove formacao-academica-titulacao from others
	for i, cv := range othersCopy {
		cvMap, ok := cv.(map[string]interface{})
		if !ok {
			continue
		}
		removeFieldFromCV(cvMap, "formacao-academica-titulacao")
		othersCopy[i] = cvMap
	}
	combined["outros_pesquisadores"] = othersCopy

	if estimateTokensAny(combined) <= maxTokens {
		b, _ := json.Marshal(combined)
		return string(b), true
	}

	// Step 4: remove others one by one
	for n := len(othersCopy); n >= 0; n-- {
		combined["outros_pesquisadores"] = othersCopy[:n]
		if estimateTokensAny(combined) <= maxTokens {
			b, _ := json.Marshal(combined)
			return string(b), true
		}
	}

	// Step 5: aggressively truncate the main CV itself
	aggressiveCopy := deepCopy(currentCopy)
	if aggressiveCopy != nil {
		cv, ok := getInnerMap(aggressiveCopy, "curriculo-vitae")
		if ok {
			delete(cv, "dados-complementares")
			delete(cv, "outra-producao")
			delete(cv, "producao-tecnica")
		}
		combined["pesquisador_alvo"] = aggressiveCopy
		combined["outros_pesquisadores"] = othersCopy[:0]
		if estimateTokensAny(combined) <= maxTokens {
			b, _ := json.Marshal(combined)
			return string(b), true
		}

		// Remove producao-bibliografica from main CV as last resort
		if cv != nil {
			delete(cv, "producao-bibliografica")
		}
		if estimateTokensAny(combined) <= maxTokens {
			b, _ := json.Marshal(combined)
			return string(b), true
		}
	}

	b, _ := json.Marshal(combined)
	return string(b), true
}
