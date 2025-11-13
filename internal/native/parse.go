package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/tokenize"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func NativeTokenize(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("tokenize", 1, len(args))
	}

	inputVal, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("tokenize", "string!", args[0])
	}
	input := inputVal.String()

	tokenizer := tokenize.NewTokenizer(input)
	tokenizer.SetSource("(native)")
	tokens, err := tokenizer.Tokenize()
	if err != nil {
		return value.NewNoneVal(), err
	}

	result := make([]core.Value, 0, len(tokens))
	for _, tok := range tokens {
		if tok.Type == tokenize.TokenEOF {
			continue
		}
		tokenObj := createTokenObject(tok)
		result = append(result, tokenObj)
	}

	return value.NewBlockVal(result), nil
}

func createTokenObject(tok tokenize.Token) core.Value {
	objFrame := frame.NewFrame(frame.FrameObject, -1)

	tokenType := getTokenTypeName(tok.Type)
	objFrame.Bind("type", value.NewWordVal(tokenType))
	objFrame.Bind("value", value.NewStrVal(tok.Value))
	objFrame.Bind("line", value.NewIntVal(int64(tok.Line)))
	objFrame.Bind("column", value.NewIntVal(int64(tok.Column)))

	obj := value.NewObject(objFrame)
	return value.ObjectVal(obj)
}

func getTokenTypeName(t tokenize.TokenType) string {
	switch t {
	case tokenize.TokenLiteral:
		return "literal"
	case tokenize.TokenString:
		return "string"
	case tokenize.TokenLParen:
		return "lparen"
	case tokenize.TokenRParen:
		return "rparen"
	case tokenize.TokenLBracket:
		return "lbracket"
	case tokenize.TokenRBracket:
		return "rbracket"
	case tokenize.TokenEOF:
		return "eof"
	default:
		return "unknown"
	}
}

func NativeParseValues(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("parse-values", 1, len(args))
	}

	tokensBlockVal, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("parse-values", "block!", args[0])
	}
	tokensBlock := tokensBlockVal.Elements

	tokens, err := convertToTokens(tokensBlock)
	if err != nil {
		return value.NewNoneVal(), err
	}

	parser := parse.NewParser(tokens, "")
	values, locations, err := parser.Parse()
	if err != nil {
		if vErr, ok := err.(*verror.Error); ok {
			return value.NewNoneVal(), vErr
		}
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDInvalidToken, [3]string{"parse-values", err.Error(), ""})
	}

	block := value.NewBlockVal(values)
	if blockVal, ok := value.AsBlockValue(block); ok {
		blockVal.SetLocations(locations)
	}
	return block, nil
}

// NativeParse is kept as an alias for backward compatibility during transition
// Will be replaced with the parse dialect implementation
func NativeParse(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return NativeParseValues(args, refValues, eval)
}

func convertToTokens(tokensBlock []core.Value) ([]tokenize.Token, error) {
	tokens := make([]tokenize.Token, 0, len(tokensBlock)+1)

	for _, tokenVal := range tokensBlock {
		obj, ok := value.AsObject(tokenVal)
		if !ok {
			return nil, verror.NewScriptError("type-mismatch", [3]string{"parse-values", "token object", value.TypeToString(tokenVal.GetType())})
		}

		typeVal, typeOk := obj.GetField("type")
		valueVal, valueOk := obj.GetField("value")
		lineVal, lineOk := obj.GetField("line")
		columnVal, columnOk := obj.GetField("column")

		if !typeOk || !valueOk || !lineOk || !columnOk {
			return nil, verror.NewScriptError("invalid-arg", [3]string{"parse-values", "token object must have type, value, line, and column fields", ""})
		}

		tokenTypeStrVal, ok := value.AsStringValue(typeVal)
		var tokenTypeStr string
		if ok {
			tokenTypeStr = tokenTypeStrVal.String()
		} else {
			tokenTypeStr = typeVal.Mold()
		}

		tokenValueStrVal, ok := value.AsStringValue(valueVal)
		var tokenValueStr string
		if ok {
			tokenValueStr = tokenValueStrVal.String()
		} else {
			tokenValueStr = valueVal.Mold()
		}

		lineInt, ok := value.AsIntValue(lineVal)
		if !ok {
			return nil, verror.NewScriptError("type-mismatch", [3]string{"parse-values", "token line must be integer", value.TypeToString(lineVal.GetType())})
		}

		columnInt, ok := value.AsIntValue(columnVal)
		if !ok {
			return nil, verror.NewScriptError("type-mismatch", [3]string{"parse-values", "token column must be integer", value.TypeToString(columnVal.GetType())})
		}

		tokenType := getTokenTypeFromName(tokenTypeStr)
		tokens = append(tokens, tokenize.Token{
			Type:   tokenType,
			Value:  tokenValueStr,
			Line:   int(lineInt),
			Column: int(columnInt),
		})
	}

	tokens = append(tokens, tokenize.Token{Type: tokenize.TokenEOF, Line: 0, Column: 0})
	return tokens, nil
}

func getTokenTypeFromName(name string) tokenize.TokenType {
	switch name {
	case "literal":
		return tokenize.TokenLiteral
	case "string":
		return tokenize.TokenString
	case "lparen":
		return tokenize.TokenLParen
	case "rparen":
		return tokenize.TokenRParen
	case "lbracket":
		return tokenize.TokenLBracket
	case "rbracket":
		return tokenize.TokenRBracket
	case "eof":
		return tokenize.TokenEOF
	default:
		return tokenize.TokenLiteral
	}
}

func NativeLoadString(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("load-string", 1, len(args))
	}

	inputVal, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("load-string", "string!", args[0])
	}
	input := inputVal.String()

	values, locations, err := parse.ParseWithSource(input, "(native)")
	if err != nil {
		return value.NewNoneVal(), err
	}

	block := value.NewBlockVal(values)
	if blockVal, ok := value.AsBlockValue(block); ok {
		blockVal.SetLocations(locations)
	}
	return block, nil
}

func NativeClassify(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("classify", 1, len(args))
	}

	inputVal, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("classify", "string!", args[0])
	}
	input := inputVal.String()

	parser := parse.NewParser([]tokenize.Token{}, "(native)")
	token := tokenize.Token{Type: tokenize.TokenLiteral, Value: input, Line: 1, Column: 1, Source: "(native)"}
	val, err := parser.ClassifyLiteral(token)
	if err != nil {
		if vErr, ok := err.(*verror.Error); ok {
			return value.NewNoneVal(), vErr
		}
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDInvalidToken, [3]string{"classify", err.Error(), ""})
	}

	return val, nil
}
