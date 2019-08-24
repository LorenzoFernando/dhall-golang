package eval

import . "github.com/philandstuff/dhall-golang/core"

type Env map[string][]Value

func Eval(t Term, e Env) Value {
	switch t := t.(type) {
	case Universe:
		return t
	case Builtin:
		return t
	case BoundVar:
		return e[t.Name][t.Index]
	case FreeVar:
		return t
	case LambdaTerm:
		return LambdaValue{
			Label:  t.Label,
			Domain: Eval(t.Type, e),
			Fn: func(x Value) Value {
				newEnv := Env{}
				for k, v := range e {
					newEnv[k] = v
				}
				newEnv[t.Label] = append([]Value{x}, newEnv[t.Label]...)
				return Eval(t.Body, newEnv)
			}}
	case PiTerm:
		return PiValue{
			Label:  t.Label,
			Domain: Eval(t.Type, e),
			Range: func(x Value) Value {
				newEnv := Env{}
				for k, v := range e {
					newEnv[k] = v
				}
				newEnv[t.Label] = append([]Value{x}, newEnv[t.Label]...)
				return Eval(t.Body, newEnv)
			}}
	case AppTerm:
		fn := Eval(t.Fn, e)
		arg := Eval(t.Arg, e)
		if f, ok := fn.(LambdaValue); ok {
			return f.Fn(arg)
		}
		if _, ok := fn.(PiValue); ok {
			panic("pi not implemented")
		}
		if f, ok := fn.(Neutral); ok {
			return AppNeutral{
				Fn:  f,
				Arg: arg,
			}
		}
		panic("appterm unimp")
	case NaturalLit:
		return t
	case EmptyList:
		return EmptyListVal{Type: Eval(t.Type, e)}
	default:
		panic("unknown term type")
	}
}
