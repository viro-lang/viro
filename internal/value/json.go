package value

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/core"
)

func FromJSON(jsonStr string) (core.Value, error) {
	var raw interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, err
	}
	return fromJSONValue(raw)
}

func fromJSONValue(raw interface{}) (core.Value, error) {
	switch v := raw.(type) {
	case nil:
		return NewNoneVal(), nil
	case bool:
		return NewLogicVal(v), nil
	case float64:
		if v == float64(int64(v)) {
			return NewIntVal(int64(v)), nil
		}
		big := &decimal.Big{}
		big.SetFloat64(v)
		return DecimalVal(big, 0), nil
	case string:
		return NewStrVal(v), nil
	case []interface{}:
		elements := make([]core.Value, len(v))
		for i, item := range v {
			val, err := fromJSONValue(item)
			if err != nil {
				return nil, err
			}
			elements[i] = val
		}
		return NewBlockValue(elements), nil
	case map[string]interface{}:
		obj := NewObject(nil)
		for key, value := range v {
			val, err := fromJSONValue(value)
			if err != nil {
				return nil, err
			}
			obj.Frame.Bind(key, val)
		}
		return ObjectVal(obj), nil
	default:
		return nil, fmt.Errorf("unsupported JSON type: %T", v)
	}
}

func ToJSON(val core.Value) (string, error) {
	raw, err := toJSONValue(val)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func toJSONValue(val core.Value) (interface{}, error) {
	switch val.GetType() {
	case TypeNone:
		return nil, nil
	case TypeLogic:
		if logic, ok := AsLogicValue(val); ok {
			return logic, nil
		}
	case TypeInteger:
		if intVal, ok := AsIntValue(val); ok {
			return intVal, nil
		}
	case TypeDecimal:
		if decVal, ok := AsDecimal(val); ok {
			if f, ok := decVal.Magnitude.Float64(); ok {
				return f, nil
			}
		}
	case TypeString:
		if strVal, ok := AsStringValue(val); ok {
			return strVal.String(), nil
		}
	case TypeBinary:
		if binVal, ok := AsBinaryValue(val); ok {
			return base64.StdEncoding.EncodeToString(binVal.Bytes()), nil
		}
	case TypeBlock:
		if block, ok := AsBlockValue(val); ok {
			elements := make([]interface{}, len(block.Elements))
			for i, elem := range block.Elements {
				converted, err := toJSONValue(elem)
				if err != nil {
					return nil, err
				}
				elements[i] = converted
			}
			return elements, nil
		}
	case TypeObject:
		if obj, ok := AsObject(val); ok {
			result := make(map[string]interface{})
			for _, binding := range obj.Frame.GetAll() {
				converted, err := toJSONValue(binding.Value)
				if err != nil {
					return nil, err
				}
				result[binding.Symbol] = converted
			}
			return result, nil
		}
	}
	return nil, fmt.Errorf("cannot convert %v to JSON", val.GetType())
}
