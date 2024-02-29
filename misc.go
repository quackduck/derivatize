package main

import (
	"math"
	"reflect"
	"slices"
)

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
	case *Constant, *X:
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

// mul or add
// caution: edits elements within es. Need to change that.
func mergeBasedOnReflect(es []Expression, ismul bool) []Expression {

	//for i := 0; i < len(es); i++ {
	//	fmt.Println("i", i, es[i].structure())
	//}

	// get all polys to the front
	polys, rest := splitByType[*Polynomial](es)
	for _, p := range polys {
		rest = append(rest, p)
	}
	slices.Reverse(rest)

	es = rest

	exprMap := make(map[Expression]int)
	for i := 0; i < len(es); i++ {
		exprMap[es[i]] += 1
		for j := i + 1; j < len(es); j++ {

			_, oki := es[i].(*Polynomial) // TODO: handle case where j is a poly. // TODO: do all this after the tally
			_, okj := es[j].(*Polynomial)
			if oki && okj {
				//es, j = mergePolys(es, ismul, i, j)
				p1 := es[i].(*Polynomial)
				p2 := es[j].(*Polynomial)

				changed := mergePolys(p1, p2, ismul)
				if changed { // es[i] has been modified
					es = append(es[:j], es[j+1:]...)
					j--
					continue
				}
			}

			if oki {
				var poly *Polynomial
				var other Expression
				poly = es[i].(*Polynomial)
				other = es[j]

				if reflect.DeepEqual(poly.inside, other) {
					if ismul {
						newPowerToCoeff := make(map[float64]float64, len(poly.powerToCoeff))
						for power, coeff := range poly.powerToCoeff {
							newPowerToCoeff[power+1] = coeff
						}
						poly.powerToCoeff = newPowerToCoeff
					} else { // addition
						poly.powerToCoeff[1]++
					}
					es = append(es[:j], es[j+1:]...)
					j--
					continue
				}
			}

			if reflect.DeepEqual(es[i], es[j]) {
				exprMap[es[i]]++
				es = append(es[:j], es[j+1:]...)
				j--
				continue
			}
		}
	}
	newEs := make([]Expression, 0, len(exprMap))
	for expr, count := range exprMap {
		if count == 1 {
			newEs = append(newEs, expr)
		} else if count > 1 {
			if ismul {
				newEs = append(newEs, poly(map[float64]float64{float64(count): 1}, expr))
			} else { // addition
				newEs = append(newEs, mul(num(float64(count)), expr))
			}
		}
	}
	return newEs
}

// caution: edits p1 based on p2
func mergePolys(p1 *Polynomial, p2 *Polynomial, ismul bool) (changed bool) {
	if ismul {
		return false // don't multiply the polys together
	}
	if reflect.DeepEqual(p1.inside, p2.inside) {
		// addition
		newPowerToCoeff := make(map[float64]float64, len(p1.powerToCoeff)+len(p2.powerToCoeff))
		for power, coeff := range p1.powerToCoeff {
			newPowerToCoeff[power] = coeff
		}
		for power, coeff := range p2.powerToCoeff {
			newPowerToCoeff[power] += coeff
		}
		p1.powerToCoeff = newPowerToCoeff
		return true
	}
	return false
}
