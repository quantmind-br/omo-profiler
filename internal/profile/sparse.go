package profile

import (
	"encoding/json"
	"reflect"
	"strings"
	"unicode"

	"github.com/diogenes/omo-profiler/internal/config"
)

var jsonRawMessageType = reflect.TypeOf(json.RawMessage{})

func MarshalSparse(cfg *config.Config, selection *FieldSelection, preservedUnknown map[string]json.RawMessage) ([]byte, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}

	knownValues, err := buildSelectedStruct(reflect.ValueOf(*cfg), reflect.TypeOf(*cfg), selection, "")
	if err != nil {
		return nil, err
	}
	if knownValues == nil {
		knownValues = make(map[string]interface{})
	}

	mergedValues, err := mergePreservedUnknown(knownValues, preservedUnknown)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(mergedValues, "", "  ")
}

func buildSelectedStruct(value reflect.Value, valueType reflect.Type, selection *FieldSelection, prefix string) (map[string]interface{}, error) {
	if valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
		if value.IsValid() && value.Kind() == reflect.Pointer && !value.IsNil() {
			value = value.Elem()
		} else {
			value = reflect.Zero(valueType)
		}
	}

	if !value.IsValid() {
		value = reflect.Zero(valueType)
	}

	result := make(map[string]interface{})
	for i := 0; i < valueType.NumField(); i++ {
		field := valueType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		jsonName, ok := jsonFieldName(field)
		if !ok {
			continue
		}

		fieldPath := joinSelectionPath(prefix, canonicalPathSegment(jsonName))
		fieldValue, include, err := buildSelectedValue(value.Field(i), field.Type, selection, fieldPath)
		if err != nil {
			return nil, err
		}
		if include {
			result[jsonName] = fieldValue
		}
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

func buildSelectedValue(value reflect.Value, valueType reflect.Type, selection *FieldSelection, path string) (interface{}, bool, error) {
	if valueType == jsonRawMessageType {
		if !selection.IsSelected(path) {
			return nil, false, nil
		}
		leafValue, err := marshalLeafValue(value, valueType)
		return leafValue, true, err
	}

	if valueType.Kind() == reflect.Pointer {
		if valueType.Elem().Kind() == reflect.Struct {
			var nestedValue reflect.Value
			if value.IsValid() && value.Kind() == reflect.Pointer && !value.IsNil() {
				nestedValue = value.Elem()
			} else {
				nestedValue = reflect.Zero(valueType.Elem())
			}

			nested, err := buildSelectedStruct(nestedValue, valueType.Elem(), selection, path)
			if err != nil {
				return nil, false, err
			}
			if len(nested) == 0 {
				return nil, false, nil
			}
			return nested, true, nil
		}

		if !selection.IsSelected(path) {
			return nil, false, nil
		}

		leafValue, err := marshalLeafValue(value, valueType)
		return leafValue, true, err
	}

	switch valueType.Kind() {
	case reflect.Struct:
		nested, err := buildSelectedStruct(value, valueType, selection, path)
		if err != nil {
			return nil, false, err
		}
		if len(nested) == 0 {
			return nil, false, nil
		}
		return nested, true, nil

	case reflect.Map:
		elementType := valueType.Elem()
		concreteElementType := elementType
		if concreteElementType.Kind() == reflect.Pointer {
			concreteElementType = concreteElementType.Elem()
		}

		if concreteElementType.Kind() == reflect.Struct {
			if !value.IsValid() || value.IsNil() || value.Len() == 0 {
				return nil, false, nil
			}

			nested := make(map[string]interface{})
			for _, mapKey := range value.MapKeys() {
				entryPath := joinSelectionPath(path, mapKey.String())
				entryValue, include, err := buildSelectedValue(value.MapIndex(mapKey), elementType, selection, entryPath)
				if err != nil {
					return nil, false, err
				}
				if include {
					nested[mapKey.String()] = entryValue
				}
			}

			if len(nested) == 0 {
				return nil, false, nil
			}

			return nested, true, nil
		}

		if !selection.IsSelected(path) {
			return nil, false, nil
		}

		leafValue, err := marshalLeafValue(value, valueType)
		return leafValue, true, err

	default:
		if !selection.IsSelected(path) {
			return nil, false, nil
		}

		leafValue, err := marshalLeafValue(value, valueType)
		return leafValue, true, err
	}
}

func marshalLeafValue(value reflect.Value, valueType reflect.Type) (interface{}, error) {
	if valueType == jsonRawMessageType {
		if !value.IsValid() || value.IsNil() || value.Len() == 0 {
			return nil, nil
		}

		var decoded interface{}
		if err := json.Unmarshal(value.Interface().(json.RawMessage), &decoded); err != nil {
			return nil, err
		}
		return decoded, nil
	}

	if !value.IsValid() {
		return zeroJSONValue(valueType), nil
	}

	if valueType.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil, nil
		}
		return marshalLeafValue(value.Elem(), value.Elem().Type())
	}

	if valueType.Kind() == reflect.Pointer {
		if value.IsNil() {
			return zeroJSONValue(valueType.Elem()), nil
		}
		return marshalLeafValue(value.Elem(), valueType.Elem())
	}

	if valueType.Kind() == reflect.Map && value.IsNil() {
		return map[string]interface{}{}, nil
	}

	if valueType.Kind() == reflect.Slice && value.IsNil() {
		return []interface{}{}, nil
	}

	switch valueType.Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		return value.Interface(), nil
	}

	data, err := json.Marshal(value.Interface())
	if err != nil {
		return nil, err
	}

	var decoded interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}

	return decoded, nil
}

func zeroJSONValue(valueType reflect.Type) interface{} {
	switch valueType.Kind() {
	case reflect.Pointer:
		return zeroJSONValue(valueType.Elem())
	case reflect.Bool:
		return false
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64:
		return 0
	case reflect.String:
		return ""
	case reflect.Map:
		return map[string]interface{}{}
	case reflect.Slice:
		if valueType == jsonRawMessageType {
			return nil
		}
		return []interface{}{}
	case reflect.Interface:
		return nil
	}

	return nil
}

func mergePreservedUnknown(known map[string]interface{}, preservedUnknown map[string]json.RawMessage) (map[string]interface{}, error) {
	merged := make(map[string]interface{}, len(known)+len(preservedUnknown))
	for key, value := range known {
		merged[key] = value
	}

	for key, raw := range preservedUnknown {
		preservedValue, err := decodePreservedUnknown(raw)
		if err != nil {
			return nil, err
		}

		if knownValue, exists := merged[key]; exists {
			merged[key] = mergeKnownValue(knownValue, preservedValue)
			continue
		}

		merged[key] = preservedValue
	}

	return merged, nil
}

func decodePreservedUnknown(raw json.RawMessage) (interface{}, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var decoded interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}

	return decoded, nil
}

func mergeKnownValue(knownValue, preservedValue interface{}) interface{} {
	knownMap, knownIsMap := knownValue.(map[string]interface{})
	preservedMap, preservedIsMap := preservedValue.(map[string]interface{})
	if !knownIsMap || !preservedIsMap {
		return knownValue
	}

	merged := make(map[string]interface{}, len(preservedMap)+len(knownMap))
	for key, value := range preservedMap {
		merged[key] = value
	}

	for key, value := range knownMap {
		if preservedChild, exists := merged[key]; exists {
			merged[key] = mergeKnownValue(value, preservedChild)
			continue
		}
		merged[key] = value
	}

	return merged
}

func jsonFieldName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")
	if tag == "" || tag == "-" {
		return "", false
	}

	name, _, _ := strings.Cut(tag, ",")
	if name == "" || name == "-" {
		return "", false
	}

	return name, true
}

func joinSelectionPath(prefix, segment string) string {
	if prefix == "" {
		return segment
	}
	return prefix + "." + segment
}

func canonicalPathSegment(segment string) string {
	var builder strings.Builder
	runes := []rune(segment)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				if prev != '_' && ((unicode.IsLower(prev) || unicode.IsDigit(prev)) || (unicode.IsUpper(prev) && nextIsLower)) {
					builder.WriteRune('_')
				}
			}
			builder.WriteRune(unicode.ToLower(r))
			continue
		}

		builder.WriteRune(r)
	}

	return builder.String()
}
