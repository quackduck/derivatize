package main

import "math"

type Abs struct {
	inside Expression
}

func (a *Abs) Derivative() Expression {
	return mul(div(a.inside, abs(a.inside)), a.inside.Derivative())
}

func (a *Abs) String() string {
	return "|" + a.inside.String() + "|"
}

func (a *Abs) simplify() Expression {
	a.inside = a.inside.simplify()
	return a
}

func abs(e Expression) Expression {
	return &Abs{e}
}

func (a *Abs) structure() string {
	return "abs{" + a.inside.structure() + "}"
}

type Log struct {
	base float64
	expr Expression
}

func (l *Log) simplify() Expression {
	l.expr = l.expr.simplify()

	if e, ok := l.expr.(*Exponential); ok {
		if e.base == l.base {
			return e.power
		}
	}
	return l
}

var E = math.E

func (l *Log) Derivative() Expression {
	simp := l.simplify()
	if simp != l {
		return simp.Derivative()
	}

	if l.base == E {
		return div(l.expr.Derivative(), l.expr).simplify()
	}

	return div(
		l.expr.Derivative(),
		mul(
			l.expr,
			&Log{E, &Constant{l.base}},
		),
	).simplify()
}

func (l *Log) String() string {
	simp := l.simplify()
	if simp != l {
		return simp.String()
	}
	insideStr := l.expr.String()
	switch l.expr.(type) {
	case *ExprsAdded, *ExprsSubtracted:

	case *Polynomial:
		if insideStr[0] != '(' {
			insideStr = "(" + insideStr + ")"
		}
	case *Constant:
		insideStr = "(" + insideStr + ")"
	default:
		insideStr = "[ " + insideStr + " ]"
	}
	if l.base == E {
		return "ln" + insideStr
	}
	return "log_" + ftoa(l.base) + insideStr
}

func log(base float64, expr Expression) *Log { return &Log{base, expr} }

func (l *Log) structure() string {
	return "log_" + ftoa(l.base) + "{ " + l.expr.structure() + "}"
}

type Exponential struct {
	base  float64
	power Expression
}

func (e *Exponential) simplify() Expression {
	e.power = e.power.simplify()
	if c, ok := e.power.(*Constant); ok {
		if e.base != E {
			return &Constant{math.Pow(e.base, c.num)}
		}
	}
	if l, ok := e.power.(*Log); ok {
		if l.base == e.base {
			return l.expr
		}
	}
	return e
}

func (e *Exponential) Derivative() Expression {
	simp := e.simplify()
	if simp != e {
		return simp.Derivative()
	}
	if e.base == E {
		return mul(
			exp(E, e.power),
			e.power.Derivative(),
		)
	}
	return mul(
		log(E, num(e.base)),
		exp(e.base, e.power),
		e.power.Derivative(),
	)
}

func (e *Exponential) String() string {
	simp := e.simplify()
	if simp != e {
		return simp.String()
	}
	insideStr := e.power.String()
	//if _, ok := e.power.(*Polynomial); !ok {
	//	insideStr = "[ " + insideStr + " ]"
	//}
	switch e.power.(type) {
	case *ExprsAdded, *ExprsSubtracted, *X:
	case *Polynomial:
		insideStr = "[ " + insideStr + " ]"
	}
	if e.base == E {
		return "e^" + insideStr
	}
	return ftoa(e.base) + "^" + insideStr
}

func exp(base float64, power Expression) *Exponential { return &Exponential{base, power} }

func (e *Exponential) structure() string {
	return "exp_" + ftoa(e.base) + "{" + e.power.structure() + "}"
}
