package parser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

type Summary struct {
	LattesID   string
	Name       string
	LastUpdate string
	Counts     ProductionCounts
}

type ProductionCounts struct {
	BibliographicProduction int `json:"bibliographicProduction"`
	TechnicalProduction     int `json:"technicalProduction"`
	OtherProduction         int `json:"otherProduction"`
}

type ParseResult struct {
	Document map[string]interface{}
	Summary  Summary
}

func Parse(data []byte) (*ParseResult, error) {
	reader := charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(data))

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil
	}

	var root xml.StartElement
	for {
		tok, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("arquivo XML inválido: %w", err)
		}
		if se, ok := tok.(xml.StartElement); ok {
			root = se
			break
		}
	}

	if root.Name.Local != "CURRICULO-VITAE" {
		return nil, fmt.Errorf("arquivo inválido: elemento raiz CURRICULO-VITAE não encontrado")
	}

	lattesID := attrValue(root.Attr, "NUMERO-IDENTIFICADOR")
	if lattesID == "" {
		return nil, fmt.Errorf("arquivo inválido: atributo NUMERO-IDENTIFICADOR não encontrado")
	}

	doc := make(map[string]interface{})
	for _, attr := range root.Attr {
		doc[strings.ToLower(attr.Name.Local)] = attr.Value
	}

	if err := parseChildren(decoder, doc); err != nil {
		return nil, fmt.Errorf("erro ao processar XML: %w", err)
	}

	filterDadosGerais(doc)

	fullDoc := map[string]interface{}{
		"curriculo-vitae": doc,
	}

	summary := extractSummary(doc, lattesID)

	return &ParseResult{
		Document: fullDoc,
		Summary:  summary,
	}, nil
}

func parseChildren(decoder *xml.Decoder, parent map[string]interface{}) error {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			child := make(map[string]interface{})
			for _, attr := range t.Attr {
				child[strings.ToLower(attr.Name.Local)] = attr.Value
			}

			if err := parseChildren(decoder, child); err != nil {
				return err
			}

			name := strings.ToLower(t.Name.Local)
			existing, exists := parent[name]
			if !exists {
				parent[name] = child
			} else {
				switch v := existing.(type) {
				case []interface{}:
					parent[name] = append(v, child)
				default:
					parent[name] = []interface{}{v, child}
				}
			}

		case xml.EndElement:
			return nil

		case xml.CharData:
			text := string(bytes.TrimSpace([]byte(t)))
			if text != "" {
				parent["#text"] = text
			}
		}
	}
}

var dadosGeraisAllowed = map[string]bool{
	"nome-completo":                   true,
	"orcid-id":                        true,
	"nome-em-citacoes-bibliograficas": true,
	"formacao-academica-titulacao":    true,
	"atuacoes-profissionais":          true,
	"areas-de-atuacao":                true,
}

func filterDadosGerais(cv map[string]interface{}) {
	dg, ok := cv["dados-gerais"].(map[string]interface{})
	if !ok {
		return
	}
	for key := range dg {
		if !dadosGeraisAllowed[key] {
			delete(dg, key)
		}
	}
}

func extractSummary(cv map[string]interface{}, lattesID string) Summary {
	s := Summary{
		LattesID:   lattesID,
		LastUpdate: stringField(cv, "data-atualizacao"),
	}

	if dg, ok := cv["dados-gerais"].(map[string]interface{}); ok {
		s.Name = stringField(dg, "nome-completo")
	}

	s.Counts.BibliographicProduction = countProduction(cv, "producao-bibliografica")
	s.Counts.TechnicalProduction = countProduction(cv, "producao-tecnica")
	s.Counts.OtherProduction = countProduction(cv, "outra-producao")

	return s
}

func countProduction(cv map[string]interface{}, sectionKey string) int {
	section, ok := cv[sectionKey].(map[string]interface{})
	if !ok {
		return 0
	}

	count := 0
	for _, val := range section {
		count += countItems(val)
	}
	return count
}

func countItems(v interface{}) int {
	switch val := v.(type) {
	case []interface{}:
		return len(val)
	case map[string]interface{}:
		count := 0
		hasChildElements := false
		for key, child := range val {
			if key == "#text" {
				continue
			}
			hasChildElements = true
			count += countItems(child)
		}
		if !hasChildElements {
			return 1
		}
		return count
	default:
		return 0
	}
}

func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func stringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
