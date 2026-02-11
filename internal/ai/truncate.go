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

// compactPublications replaces the full producao-bibliografica with a lightweight
// list of publication titles and years only, removing co-authors, DOIs, journal details, etc.
func compactPublications(cvs []interface{}) {
	// Known title field names in Lattes XML (lowercased)
	titleFields := map[string]string{
		"dados-basicos-do-artigo":                    "titulo-do-artigo",
		"dados-basicos-do-livro":                     "titulo-do-livro",
		"dados-basicos-do-capitulo":                  "titulo-do-capitulo-do-livro",
		"dados-basicos-do-trabalho":                  "titulo-do-trabalho",
		"dados-basicos-do-texto":                     "titulo-do-texto",
		"dados-basicos-de-outra-producao":            "titulo",
		"dados-basicos-da-traducao":                  "titulo",
		"dados-basicos-de-artigo-aceito-para-publicacao": "titulo-do-artigo-aceito-para-publicacao",
	}

	for _, cv := range cvs {
		cvMap, ok := cv.(map[string]interface{})
		if !ok {
			continue
		}
		inner, ok := cvMap["curriculo-vitae"]
		if !ok {
			continue
		}
		cvInner, ok := inner.(map[string]interface{})
		if !ok {
			continue
		}
		pb, ok := cvInner["producao-bibliografica"]
		if !ok {
			continue
		}
		pbMap, ok := pb.(map[string]interface{})
		if !ok {
			continue
		}

		// Build a compact list: [{tipo, titulo, ano}, ...]
		var publications []map[string]string

		for sectionName, sectionVal := range pbMap {
			sectionMap, ok := sectionVal.(map[string]interface{})
			if !ok {
				continue
			}
			for pubType, pubVal := range sectionMap {
				extractPubs(pubVal, pubType, titleFields, &publications)
			}
			_ = sectionName
		}

		// Replace heavy producao-bibliografica with compact list
		cvInner["producao-bibliografica"] = publications
	}
}

// extractPubs extracts title and year from publication items (single or array).
func extractPubs(val interface{}, pubType string, titleFields map[string]string, out *[]map[string]string) {
	items := toSlice(val)
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		pub := map[string]string{"tipo": pubType}
		// Search for dados-basicos-* fields to find title and year
		for dbKey, titleKey := range titleFields {
			db, ok := itemMap[dbKey]
			if !ok {
				continue
			}
			dbMap, ok := db.(map[string]interface{})
			if !ok {
				continue
			}
			if t, ok := dbMap[titleKey].(string); ok && t != "" {
				pub["titulo"] = t
			}
			// Year fields vary: ano-do-artigo, ano, ano-do-trabalho, etc.
			for k, v := range dbMap {
				if len(k) >= 3 && k[:3] == "ano" {
					if s, ok := v.(string); ok && s != "" {
						pub["ano"] = s
						break
					}
				}
			}
			break
		}
		if pub["titulo"] != "" {
			*out = append(*out, pub)
		}
	}
}

// toSlice converts a value to a slice of interfaces.
func toSlice(val interface{}) []interface{} {
	switch v := val.(type) {
	case []interface{}:
		return v
	case map[string]interface{}:
		return []interface{}{v}
	default:
		return nil
	}
}

// TruncateChatData compacts and truncates multiple CVs to fit within a token budget for chat.
// First it compacts publications to title+year only, then progressively removes fields.
func TruncateChatData(cvs []map[string]interface{}, maxTokens int) (string, bool) {
	copies := make([]interface{}, len(cvs))
	for i, cv := range cvs {
		copies[i] = deepCopyAny(cv)
	}

	// Step 0: compact producao-bibliografica to titles+years only (massive size reduction)
	compactPublications(copies)

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

	// Step 3: remove formacao-academica-titulacao
	removeFieldFromAll(copies, "formacao-academica-titulacao")
	wrapper["curriculos"] = copies
	if estimateTokensAny(wrapper) <= maxTokens {
		b, _ := json.Marshal(copies)
		return string(b), true
	}

	// Step 4: trim publication lists (keep recent ones)
	for _, cv := range copies {
		cvMap, ok := cv.(map[string]interface{})
		if !ok {
			continue
		}
		inner, ok := cvMap["curriculo-vitae"]
		if !ok {
			continue
		}
		cvInner, ok := inner.(map[string]interface{})
		if !ok {
			continue
		}
		pubs, ok := cvInner["producao-bibliografica"]
		if !ok {
			continue
		}
		pubList, ok := pubs.([]map[string]string)
		if !ok {
			continue
		}
		// Keep only first 30 publications per researcher
		if len(pubList) > 30 {
			cvInner["producao-bibliografica"] = pubList[:30]
		}
	}
	wrapper["curriculos"] = copies
	if estimateTokensAny(wrapper) <= maxTokens {
		b, _ := json.Marshal(copies)
		return string(b), true
	}

	// Step 5: remove producao-bibliografica entirely (last resort before removing CVs)
	removeFieldFromAll(copies, "producao-bibliografica")
	wrapper["curriculos"] = copies
	if estimateTokensAny(wrapper) <= maxTokens {
		b, _ := json.Marshal(copies)
		return string(b), true
	}

	// Step 6: remove CVs from the end
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
