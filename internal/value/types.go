package value

import "github.com/marcin-radoszewski/viro/internal/core"

const (
	TypeNone core.ValueType = iota
	TypeLogic
	TypeInteger
	TypeString
	TypeWord
	TypeSetWord
	TypeGetWord
	TypeLitWord
	TypeBlock
	TypeParen
	TypeFunction
	TypeDecimal
	TypeObject
	TypePort
	TypePath
	TypeGetPath
	TypeSetPath
	TypeDatatype
	TypeBinary
	TypeWebUIWindow
)

func TypeToString(t core.ValueType) string {
	switch t {
	case TypeNone:
		return "none!"
	case TypeLogic:
		return "logic!"
	case TypeInteger:
		return "integer!"
	case TypeString:
		return "string!"
	case TypeWord:
		return "word!"
	case TypeSetWord:
		return "set-word!"
	case TypeGetWord:
		return "get-word!"
	case TypeLitWord:
		return "lit-word!"
	case TypeBlock:
		return "block!"
	case TypeParen:
		return "paren!"
	case TypeFunction:
		return "function!"
	case TypeDecimal:
		return "decimal!"
	case TypeObject:
		return "object!"
	case TypePort:
		return "port!"
	case TypePath:
		return "path!"
	case TypeGetPath:
		return "get-path!"
	case TypeSetPath:
		return "set-path!"
	case TypeDatatype:
		return "datatype!"
	case TypeBinary:
		return "binary!"
	case TypeWebUIWindow:
		return "webui-window!"
	default:
		return "unknown!"
	}
}

func IsWord(t core.ValueType) bool {
	return t == TypeWord || t == TypeSetWord || t == TypeGetWord || t == TypeLitWord
}

func IsSeries(t core.ValueType) bool {
	return t == TypeBlock || t == TypeParen || t == TypeString || t == TypeBinary
}
