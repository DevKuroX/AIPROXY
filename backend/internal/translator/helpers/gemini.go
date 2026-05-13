// Package helpers provides utility functions for translator operations.
// ref: open-sse/translator/helpers/geminiHelper.js
package helpers

var UnsupportedSchemaConstraints = []string{
	"minLength", "maxLength", "exclusiveMinimum", "exclusiveMaximum",
	"pattern", "minItems", "maxItems", "format",
	"default", "examples",
	"$schema", "$defs", "definitions", "const", "$ref", "$comment",
	"additionalProperties", "propertyNames", "patternProperties", "enumDescriptions",
	"anyOf", "oneOf", "allOf", "not",
	"dependencies", "dependentSchemas", "dependentRequired",
	"title", "if", "then", "else", "contentMediaType", "contentEncoding",
	"cornerRadius", "fillColor", "fontFamily", "fontSize", "fontWeight",
	"gap", "padding", "strokeColor", "strokeThickness", "textColor",
}

type GeminiSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

var DefaultGeminiSafetySettings = []GeminiSafetySetting{
	{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "OFF"},
	{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "OFF"},
	{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "OFF"},
	{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "OFF"},
	{Category: "HARM_CATEGORY_CIVIC_INTEGRITY", Threshold: "OFF"},
}

func Convert0penAIContentToGeminiParts(content interface{}) []map[string]interface{} {
	var parts []map[string]interface{}

	if str, ok := content.(string); ok {
		parts = append(parts, map[string]interface{}{"text": str})
		return parts
	}

	contentArr, ok := content.([]interface{})
	if !ok {
		return parts
	}

	for _, itemInterface := range contentArr {
		item, ok := itemInterface.(map[string]interface{})
		if !ok {
			continue
		}

		itemType, _ := item["type"].(string)

		if itemType == "text" {
			text, _ := item["text"].(string)
			parts = append(parts, map[string]interface{}{"text": text})
			continue
		}

		if itemType == "image_url" {
			imageURL, _ := item["image_url"].(map[string]interface{})
			url, _ := imageURL["url"].(string)

			if len(url) > 5 && url[:5] == "data:" {
				commaIdx := -1
				for i, c := range url {
					if c == ',' {
						commaIdx = i
						break
					}
				}

				if commaIdx != -1 {
					mimePart := url[5:commaIdx]
					data := url[commaIdx+1:]
					mimeType := mimePart
					for i, c := range mimePart {
						if c == ';' {
							mimeType = mimePart[:i]
							break
						}
					}

					parts = append(parts, map[string]interface{}{
						"inlineData": map[string]interface{}{
							"mime_type": mimeType,
							"data":      data,
						},
					})
				}
			} else if len(url) > 4 && (url[:4] == "http" || url[:5] == "https") {
				parts = append(parts, map[string]interface{}{
					"fileData": map[string]interface{}{
						"fileUri":  url,
						"mimeType": "image/*",
					},
				})
			}
			continue
		}

		if itemType == "input_audio" {
			inputAudio, _ := item["input_audio"].(map[string]interface{})
			data, _ := inputAudio["data"].(string)
			format, _ := inputAudio["format"].(string)
			if format == "" {
				format = "wav"
			}

			var mimeType string
			if format == "mp3" {
				mimeType = "audio/mpeg"
			} else {
				mimeType = "audio/" + format
			}

			parts = append(parts, map[string]interface{}{
				"inlineData": map[string]interface{}{
					"mime_type": mimeType,
					"data":      data,
				},
			})
			continue
		}

		if itemType == "audio_url" {
			audioURL, _ := item["audio_url"].(map[string]interface{})
			url, _ := audioURL["url"].(string)

			if len(url) > 5 && url[:5] == "data:" {
				commaIdx := -1
				for i, c := range url {
					if c == ',' {
						commaIdx = i
						break
					}
				}

				if commaIdx != -1 {
					mimePart := url[5:commaIdx]
					data := url[commaIdx+1:]
					mimeType := mimePart
					for i, c := range mimePart {
						if c == ';' {
							mimeType = mimePart[:i]
							break
						}
					}

					parts = append(parts, map[string]interface{}{
						"inlineData": map[string]interface{}{
							"mime_type": mimeType,
							"data":      data,
						},
					})
				}
			}
		}
	}

	return parts
}

func ExtractTextContent(content interface{}) string {
	if str, ok := content.(string); ok {
		return str
	}

	if arr, ok := content.([]interface{}); ok {
		var result string
		for _, item := range arr {
			if block, ok := item.(map[string]interface{}); ok {
				if blockType, _ := block["type"].(string); blockType == "text" {
					if text, _ := block["text"].(string); text != "" {
						result += text
					}
				}
			}
		}
		return result
	}

	return ""
}

func CleanJSONSchemaForAntigravity(schema map[string]interface{}) map[string]interface{} {
	if schema == nil {
		return schema
	}

	convertConstToEnum(schema)
	convertEnumValuesToStrings(schema)
	mergeAllOf(schema)
	flattenAnyOfOneOf(schema)
	flattenTypeArrays(schema)
	ensureObjectType(schema)
	removeUnsupportedKeywords(schema, UnsupportedSchemaConstraints)
	cleanupRequiredFields(schema)
	addPlaceholders(schema)

	return schema
}

func convertConstToEnum(obj map[string]interface{}) {
	if obj == nil {
		return
	}

	if _, hasConst := obj["const"]; hasConst {
		if _, hasEnum := obj["enum"]; !hasEnum {
			obj["enum"] = []interface{}{obj["const"]}
			delete(obj, "const")
		}
	}

	for _, v := range obj {
		if child, ok := v.(map[string]interface{}); ok {
			convertConstToEnum(child)
		}
	}
}

func convertEnumValuesToStrings(obj map[string]interface{}) {
	if obj == nil {
		return
	}

	if enum, ok := obj["enum"].([]interface{}); ok {
		strEnum := make([]interface{}, len(enum))
		for i, v := range enum {
			strEnum[i] = toString(v)
		}
		obj["enum"] = strEnum

		if _, hasType := obj["type"]; !hasType {
			obj["type"] = "string"
		}
	}

	for _, v := range obj {
		if child, ok := v.(map[string]interface{}); ok {
			convertEnumValuesToStrings(child)
		}
	}
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		return ""
	}
}

func mergeAllOf(obj map[string]interface{}) {
	if obj == nil {
		return
	}

	if allOf, ok := obj["allOf"].([]interface{}); ok {
		merged := map[string]interface{}{}

		for _, item := range allOf {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			if props, ok := itemMap["properties"].(map[string]interface{}); ok {
				if merged["properties"] == nil {
					merged["properties"] = map[string]interface{}{}
				}
				mergedProps := merged["properties"].(map[string]interface{})
				for k, v := range props {
					mergedProps[k] = v
				}
			}

			if req, ok := itemMap["required"].([]interface{}); ok {
				if merged["required"] == nil {
					merged["required"] = []interface{}{}
				}
				mergedReq := merged["required"].([]interface{})
				for _, r := range req {
					found := false
					for _, existing := range mergedReq {
						if existing == r {
							found = true
							break
						}
					}
					if !found {
						mergedReq = append(mergedReq, r)
					}
				}
				merged["required"] = mergedReq
			}
		}

		delete(obj, "allOf")

		if props, ok := merged["properties"].(map[string]interface{}); ok {
			if obj["properties"] == nil {
				obj["properties"] = map[string]interface{}{}
			}
			objProps := obj["properties"].(map[string]interface{})
			for k, v := range props {
				objProps[k] = v
			}
		}

		if req, ok := merged["required"].([]interface{}); ok {
			if obj["required"] == nil {
				obj["required"] = []interface{}{}
			}
			objReq := obj["required"].([]interface{})
			for _, r := range req {
				found := false
				for _, existing := range objReq {
					if existing == r {
						found = true
						break
					}
				}
				if !found {
					objReq = append(objReq, r)
				}
			}
			obj["required"] = objReq
		}
	}

	for _, v := range obj {
		if child, ok := v.(map[string]interface{}); ok {
			mergeAllOf(child)
		}
	}
}

func flattenAnyOfOneOf(obj map[string]interface{}) {
	if obj == nil {
		return
	}

	if anyOf, ok := obj["anyOf"].([]interface{}); ok && len(anyOf) > 0 {
		nonNull := make([]map[string]interface{}, 0)
		for _, s := range anyOf {
			if sm, ok := s.(map[string]interface{}); ok {
				if t, _ := sm["type"].(string); t != "null" {
					nonNull = append(nonNull, sm)
				}
			}
		}
		if len(nonNull) > 0 {
			best := selectBestSchema(nonNull)
			delete(obj, "anyOf")
			for k, v := range nonNull[best] {
				obj[k] = v
			}
		}
	}

	if oneOf, ok := obj["oneOf"].([]interface{}); ok && len(oneOf) > 0 {
		nonNull := make([]map[string]interface{}, 0)
		for _, s := range oneOf {
			if sm, ok := s.(map[string]interface{}); ok {
				if t, _ := sm["type"].(string); t != "null" {
					nonNull = append(nonNull, sm)
				}
			}
		}
		if len(nonNull) > 0 {
			best := selectBestSchema(nonNull)
			delete(obj, "oneOf")
			for k, v := range nonNull[best] {
				obj[k] = v
			}
		}
	}

	for _, v := range obj {
		if child, ok := v.(map[string]interface{}); ok {
			flattenAnyOfOneOf(child)
		}
	}
}

func selectBestSchema(items []map[string]interface{}) int {
	bestIdx := 0
	bestScore := -1

	for i, item := range items {
		score := 0
		t, _ := item["type"].(string)

		if t == "object" || item["properties"] != nil {
			score = 3
		} else if t == "array" || item["items"] != nil {
			score = 2
		} else if t != "" && t != "null" {
			score = 1
		}

		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}

	return bestIdx
}

func flattenTypeArrays(obj map[string]interface{}) {
	if obj == nil {
		return
	}

	if t, ok := obj["type"].([]interface{}); ok {
		nonNull := make([]interface{}, 0)
		for _, typ := range t {
			if s, ok := typ.(string); ok && s != "null" {
				nonNull = append(nonNull, s)
			}
		}
		if len(nonNull) > 0 {
			if s, ok := nonNull[0].(string); ok {
				obj["type"] = s
			}
		} else {
			obj["type"] = "string"
		}
	}

	for _, v := range obj {
		if child, ok := v.(map[string]interface{}); ok {
			flattenTypeArrays(child)
		}
	}
}

func ensureObjectType(obj map[string]interface{}) {
	if obj == nil {
		return
	}

	if obj["properties"] != nil && obj["type"] == nil {
		obj["type"] = "object"
	}

	for _, v := range obj {
		if child, ok := v.(map[string]interface{}); ok {
			ensureObjectType(child)
		}
	}
}

func removeUnsupportedKeywords(obj map[string]interface{}, keywords []string) {
	if obj == nil {
		return
	}

	for k := range obj {
		for _, kw := range keywords {
			if k == kw {
				delete(obj, k)
				break
			}
		}
		if len(k) > 2 && k[:2] == "x-" {
			delete(obj, k)
		}
	}

	for _, v := range obj {
		switch val := v.(type) {
		case map[string]interface{}:
			removeUnsupportedKeywords(val, keywords)
		case []interface{}:
			for _, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					removeUnsupportedKeywords(itemMap, keywords)
				}
			}
		}
	}
}

func cleanupRequiredFields(obj map[string]interface{}) {
	if obj == nil {
		return
	}

	if required, ok := obj["required"].([]interface{}); ok {
		props, _ := obj["properties"].(map[string]interface{})
		if props != nil {
			validRequired := make([]interface{}, 0)
			for _, r := range required {
				if _, exists := props[r.(string)]; exists {
					validRequired = append(validRequired, r)
				}
			}
			if len(validRequired) == 0 {
				delete(obj, "required")
			} else {
				obj["required"] = validRequired
			}
		}
	}

	for _, v := range obj {
		if child, ok := v.(map[string]interface{}); ok {
			cleanupRequiredFields(child)
		}
	}
}

func addPlaceholders(obj map[string]interface{}) {
	if obj == nil {
		return
	}

	if t, _ := obj["type"].(string); t == "object" {
		props, _ := obj["properties"].(map[string]interface{})
		if props == nil || len(props) == 0 {
			obj["properties"] = map[string]interface{}{
				"reason": map[string]interface{}{
					"type":        "string",
					"description": "Brief explanation of why you are calling this tool",
				},
			}
			obj["required"] = []interface{}{"reason"}
		}
	}

	for _, v := range obj {
		if child, ok := v.(map[string]interface{}); ok {
			addPlaceholders(child)
		}
	}
}
