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
	// Truncate the main CV (remove low-value fields first, preserve producao-bibliografica)
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

	// Step 1: remove low-value fields from others (NOT producao-bibliografica)
	lowValueFields := []string{"dados-complementares", "outra-producao", "producao-tecnica"}
	for _, field := range lowValueFields {
		removeFieldFromAll(othersCopy, field)
	}
	combined["outros_pesquisadores"] = othersCopy

	if estimateTokensAny(combined) <= maxTokens {
		b, _ := json.Marshal(combined)
		return string(b), true
	}

	// Step 2: remove atuacoes-profissionais from others
	removeFieldFromAll(othersCopy, "atuacoes-profissionais")
	combined["outros_pesquisadores"] = othersCopy

	if estimateTokensAny(combined) <= maxTokens {
		b, _ := json.Marshal(combined)
		return string(b), true
	}

	// Step 3: remove formacao-academica-titulacao from others
	removeFieldFromAll(othersCopy, "formacao-academica-titulacao")
	combined["outros_pesquisadores"] = othersCopy

	if estimateTokensAny(combined) <= maxTokens {
		b, _ := json.Marshal(combined)
		return string(b), true
	}

	// Step 4: trim producao-bibliografica arrays in others (keep fewer items)
	for fraction := 2; fraction <= 8; fraction *= 2 {
		for _, cv := range othersCopy {
			cvMap, ok := cv.(map[string]interface{})
			if !ok {
				continue
			}
			trimProdBibliograficaInCV(cvMap, fraction)
		}
		combined["outros_pesquisadores"] = othersCopy
		if estimateTokensAny(combined) <= maxTokens {
			b, _ := json.Marshal(combined)
			return string(b), true
		}
	}

	// Step 5: remove producao-bibliografica from others entirely (last resort for others)
	removeFieldFromAll(othersCopy, "producao-bibliografica")
	combined["outros_pesquisadores"] = othersCopy

	if estimateTokensAny(combined) <= maxTokens {
		b, _ := json.Marshal(combined)
		return string(b), true
	}

	// Step 6: remove others one by one
	for n := len(othersCopy); n >= 0; n-- {
		combined["outros_pesquisadores"] = othersCopy[:n]
		if estimateTokensAny(combined) <= maxTokens {
			b, _ := json.Marshal(combined)
			return string(b), true
		}
	}

	// Step 7: truncate main CV producao-bibliografica
	truncateMainCVProdBib(combined, maxTokens)

	b, _ := json.Marshal(combined)
	return string(b), true
}

// removeFieldFromAll removes a field from all CVs in the slice.
func removeFieldFromAll(cvs []interface{}, field string) {
	for _, cv := range cvs {
		cvMap, ok := cv.(map[string]interface{})
		if !ok {
			continue
		}
		removeFieldFromCV(cvMap, field)
	}
}

// trimProdBibliograficaInCV reduces producao-bibliografica arrays by the given fraction.
func trimProdBibliograficaInCV(cvMap map[string]interface{}, fraction int) {
	inner, ok := cvMap["curriculo-vitae"]
	if !ok {
		return
	}
	cv, ok := inner.(map[string]interface{})
	if !ok {
		return
	}
	pb, ok := getInnerMap(cv, "producao-bibliografica")
	if !ok {
		return
	}
	for _, v := range pb {
		section, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		for k, iv := range section {
			arr, ok := iv.([]interface{})
			if !ok || len(arr) == 0 {
				continue
			}
			keep := len(arr) / fraction
			if keep < 1 {
				keep = 1
			}
			section[k] = arr[:keep]
		}
	}
}

// TruncateChatData truncates multiple CVs to fit within a token budget for chat context.
// It progressively removes less important fields to fit within the limit.
func TruncateChatData(cvs []map[string]interface{}, maxTokens int) (string, bool) {
	copies := make([]interface{}, len(cvs))
	for i, cv := range cvs {
		copies[i] = deepCopyAny(cv)
	}

	wrapper := map[string]interface{}{"curriculos": copies}
	if estimateTokensAny(wrapper) <= maxTokens {
		b, _ := json.Marshal(copies)
		return string(b), false
	}

	// Step 1: remove low-value fields
	for _, field := range []string{"dados-complementares", "outra-producao", "producao-tecnica"} {
		removeFieldFromAll(copies, field)
	}
	wrapper["curriculos"] = copies
	if estimateTokensAny(wrapper) <= maxTokens {
		b, _ := json.Marshal(copies)
		return string(b), true
	}

	// Step 2: remove atuacoes-profissionais
	removeFieldFromAll(copies, "atuacoes-profissionais")
	wrapper["curriculos"] = copies
	if estimateTokensAny(wrapper) <= maxTokens {
		b, _ := json.Marshal(copies)
		return string(b), true
	}

	// Step 3: trim producao-bibliografica arrays progressively
	for fraction := 2; fraction <= 16; fraction *= 2 {
		for _, cv := range copies {
			cvMap, ok := cv.(map[string]interface{})
			if !ok {
				continue
			}
			trimProdBibliograficaInCV(cvMap, fraction)
		}
		wrapper["curriculos"] = copies
		if estimateTokensAny(wrapper) <= maxTokens {
			b, _ := json.Marshal(copies)
			return string(b), true
		}
	}

	// Step 4: remove producao-bibliografica entirely
	removeFieldFromAll(copies, "producao-bibliografica")
	wrapper["curriculos"] = copies
	if estimateTokensAny(wrapper) <= maxTokens {
		b, _ := json.Marshal(copies)
		return string(b), true
	}

	// Step 5: remove formacao-academica-titulacao
	removeFieldFromAll(copies, "formacao-academica-titulacao")
	wrapper["curriculos"] = copies
	if estimateTokensAny(wrapper) <= maxTokens {
		b, _ := json.Marshal(copies)
		return string(b), true
	}

	// Step 6: keep only names (last resort) — remove CVs from the end
	for n := len(copies); n >= 0; n-- {
		wrapper["curriculos"] = copies[:n]
		if estimateTokensAny(wrapper) <= maxTokens {
			b, _ := json.Marshal(copies[:n])
			return string(b), true
		}
	}

	b, _ := json.Marshal(copies)
	return string(b), true
}

// truncateMainCVProdBib progressively trims the main CV's producao-bibliografica.
func truncateMainCVProdBib(combined map[string]interface{}, maxTokens int) {
	main, ok := combined["pesquisador_alvo"]
	if !ok {
		return
	}
	mainMap, ok := main.(map[string]interface{})
	if !ok {
		return
	}
	cv, ok := getInnerMap(mainMap, "curriculo-vitae")
	if !ok {
		return
	}
	pb, ok := getInnerMap(cv, "producao-bibliografica")
	if !ok {
		return
	}
	arrays := collectArrays(pb)
	if len(arrays) == 0 {
		return
	}
	maxLen := maxArrayLen(arrays)
	for n := maxLen / 2; n >= 1 && estimateTokensAny(combined) > maxTokens; n /= 2 {
		for _, entry := range arrays {
			if len(entry.slice) > n {
				entry.parent[entry.key] = entry.slice[:n]
			}
		}
	}
}
