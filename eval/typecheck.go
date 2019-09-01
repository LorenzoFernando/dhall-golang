package eval

import (
	"errors"
	"fmt"

	. "github.com/philandstuff/dhall-golang/core"
)

type Context map[string][]Value

func (ctx Context) extend(name string, t Value) Context {
	newctx := Context{}
	for k, v := range ctx {
		newctx[k] = v
	}
	newctx[name] = append(newctx[name], t)
	return newctx
}

func functionCheck(input Universe, output Universe) Universe {
	switch {
	case output == Type:
		return Type
	case input < output:
		return output
	default:
		return input
	}
}

func TypeOf(gamma Context, t Term) (Value, error) {
	switch t := t.(type) {
	case Universe:
		switch t {
		case Type:
			return Kind, nil
		case Kind:
			return Sort, nil
		case Sort:
			return nil, mkTypeError(untyped)
		default:
			return nil, mkTypeError(unhandledTypeCase)
		}
	case Builtin:
		switch t {
		case Natural:
			return Type, nil
		case List:
			return FnTypeVal(Type, Type), nil
		default:
			return nil, mkTypeError(unhandledTypeCase)
		}
	case BoundVar:
		return nil, mkTypeError(typeCheckBoundVar)
	case LocalVar:
		if vals, ok := gamma[t.Name]; ok {
			if t.Index < len(vals) {
				return vals[t.Index], nil
			}
			return nil, mkTypeError(unboundVariable(t))
		}
		return nil, fmt.Errorf("Unknown variable %s", t.Name)
	case FreeVar:
		return nil, errors.New("typecheck freevar unimp")
	case AppTerm:
		fnType, err := TypeOf(gamma, t.Fn)
		if err != nil {
			return nil, err
		}
		argType, err := TypeOf(gamma, t.Arg)
		if err != nil {
			return nil, err
		}
		piType, ok := fnType.(PiValue)
		if !ok {
			return nil, mkTypeError(notAFunction)
		}
		expectedType := Quote(piType.Domain)
		actualType := Quote(argType)
		if !judgmentallyEqual(expectedType, actualType) {
			return nil, mkTypeError(typeMismatch(expectedType, actualType))
		}
		bodyType := piType.Range(argType)
		return bodyType, nil
	case LambdaTerm:
		pi := PiTerm{Label: t.Label, Type: t.Type}
		freshLocal := LocalVar{Name: t.Label, Index: len(gamma[t.Label])}
		bt, err := TypeOf(
			gamma.extend(t.Label, Eval(t.Type, Env{})),
			subst(t.Label, freshLocal, t.Body))
		if err != nil {
			return nil, err
		}
		pi.Body = quoteAndRebindLocal(bt, freshLocal)
		_, err = TypeOf(gamma, pi)
		if err != nil {
			return nil, err
		}
		return Eval(pi, Env{}), nil
	case PiTerm:
		inUniv, err := TypeOf(gamma, t.Type)
		if err != nil {
			return nil, err
		}
		i, ok := inUniv.(Universe)
		if !ok {
			return nil, mkTypeError(invalidInputType)
		}
		freshLocal := LocalVar{Name: t.Label, Index: len(gamma[t.Label])}
		outUniv, err := TypeOf(
			gamma.extend(t.Label, Eval(t.Type, Env{})),
			subst(t.Label, freshLocal, t.Body))
		if err != nil {
			return nil, err
		}
		o, ok := outUniv.(Universe)
		if !ok {
			return nil, mkTypeError(invalidOutputType)
		}
		return functionCheck(i, o), nil
	case NaturalLit:
		return Natural, nil
	case EmptyList:
		_, err := TypeOf(gamma, t.Type)
		if err != nil {
			return nil, err
		}
		_, ok := Eval(t.Type, Env{}).(AppNeutral)
		if !ok {
			return nil, mkTypeError(invalidListType)
		}
		return Eval(t.Type, Env{}), nil
	}
	return nil, mkTypeError(unhandledTypeCase)
}

type typeError struct {
	ctx     Context
	message typeMessage
}

func mkTypeError(message typeMessage) typeError {
	return typeError{message: message}
}

func (t typeError) Error() string {
	return t.message.String()
}

type typeMessage interface {
	String() string
}

type staticTypeMessage struct{ text string }
type oneArgTypeMessage struct {
	format string
	expr   Term
}
type twoArgTypeMessage struct {
	format string
	expr0  Term
	expr1  Term
}

func (m staticTypeMessage) String() string { return m.text }
func (m oneArgTypeMessage) String() string {
	return fmt.Sprintf(m.format, m.expr)
}
func (m twoArgTypeMessage) String() string {
	return fmt.Sprintf(m.format, m.expr0, m.expr1)
}

func unboundVariable(e Term) typeMessage {
	return oneArgTypeMessage{
		format: "Unbound variable: %v",
		expr:   e,
	}
}

func typeMismatch(expectedType, actualType Term) typeMessage {
	return twoArgTypeMessage{
		format: "Wrong type of function argument\n" +
			"\n" +
			"expected %v but got %v",
		expr0: expectedType,
		expr1: actualType,
	}
}

var (
	invalidListType   = staticTypeMessage{"Invalid type for ❰List❱"}
	invalidInputType  = staticTypeMessage{"Invalid function input"}
	invalidOutputType = staticTypeMessage{"Invalid function output"}
	notAFunction      = staticTypeMessage{"Not a function"}
	untyped           = staticTypeMessage{"❰Sort❱ has no type, kind, or sort"}

	unhandledTypeCase = staticTypeMessage{"Internal error: unhandled case in TypeOf()"}
	typeCheckBoundVar = staticTypeMessage{"Internal error: shouldn't ever see BoundVar in TypeOf()"}
)
