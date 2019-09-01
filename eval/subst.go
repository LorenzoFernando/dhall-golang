package eval

import (
	. "github.com/philandstuff/dhall-golang/core"
)

func subst(name string, replacement, t Term) Term {
	return substAtLevel(0, name, replacement, t)
}

func substAtLevel(i int, name string, replacement, t Term) Term {
	switch t := t.(type) {
	case Universe:
		return t
	case Builtin:
		return t
	case BoundVar:
		if t.Name == name && t.Index == i {
			return replacement
		}
		return t
	case FreeVar:
		return t
	case LambdaTerm:
		return LambdaTerm{
			Label: t.Label,
			Type:  substAtLevel(i, name, replacement, t.Type),
			Body:  substAtLevel(i+1, name, replacement, t.Body),
		}
	case PiTerm:
		return PiTerm{
			Label: t.Label,
			Type:  substAtLevel(i, name, replacement, t.Type),
			Body:  substAtLevel(i+1, name, replacement, t.Body),
		}
	case AppTerm:
		return AppTerm{
			Fn:  substAtLevel(i, name, replacement, t.Fn),
			Arg: substAtLevel(i, name, replacement, t.Arg),
		}
	case NaturalLit:
		return t
	case EmptyList:
		return EmptyList{
			Type: substAtLevel(i, name, replacement, t.Type),
		}
	default:
		panic("unknown term type")
	}
}
