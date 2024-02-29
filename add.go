package main

import (
	"reflect"
	"slices"
	"strings"
)

type ExprsAdded struct {
	es []Expression
}

func (e *ExprsAdded) simplify() Expression {
	if len(e.es) == 1 {
		return e.es[0].simplify()
	}
	e.es = mergeBasedOnReflect(e.es, false)
	if p, ok := checkIfCanBecomePolynomial(e); ok {
		return p.simplify()
	}

	merged := make([]Expression, 0, len(e.es))
	// combine constants
	constant := 0.0
	constCount := 0
	newTerms := 0
	for _, expr := range e.es {
		if add2, ok := expr.(*ExprsAdded); ok {
			// simplify the merged expressions
			for i, expr2 := range add2.es {
				add2.es[i] = expr2.simplify()
			}
			merged = append(merged, add2.es...)
			newTerms += len(add2.es) - 1
		} else {
			merged = append(merged, expr.simplify())
		}
	}

	noConsts := make([]Expression, 0, len(merged))
	for _, expr := range merged {
		if c, ok := expr.(*Constant); ok {
			constant += c.num
			constCount++
		} else {
			noConsts = append(noConsts, expr)
		}
	}

	noConsts = addPolysAndVars(noConsts)

	if len(noConsts) == 0 {
		return num(constant)
	}
	if constant != 0.0 {
		noConsts = append([]Expression{&Constant{constant}}, noConsts...)
	}
	e.es = noConsts
	return e
}

// mul or add
func mergeBasedOnReflect(es []Expression, ismul bool) []Expression {

	// get all polys to the front
	polys, rest := splitByType[*Polynomial](es)
	for _, p := range polys {
		rest = append(rest, p)
	}
	slices.Reverse(rest)

	exprMap := make(map[Expression]int)
	for i := 0; i < len(es); i++ {
		exprMap[es[i]] = 1
		for j := i + 1; j < len(es); j++ {
			if reflect.DeepEqual(es[i], es[j]) {
				exprMap[es[i]]++
				es = append(es[:j], es[j+1:]...)
				j--
			}
			_, oki := es[i].(*Polynomial) // TODO: handle case where j is a poly. // TODO: do all this after the tally
			_, okj := es[j].(*Polynomial)
			if oki || okj {
				var poly *Polynomial
				var other Expression

				if oki {
				} else {
					continue
					// we'll swap i and j so that i is the polynomial and we don't have to think about whether i or j needs to decrement later
					es[i], es[j] = es[j], es[i]
				}
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
				}
				if !oki {
					es[i], es[j] = es[j], es[i] // swap back
				}
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

func checkIfCanBecomePolynomial(a *ExprsAdded) (*Polynomial, bool) {
	// check if the expressions are made of just nums and either x's or y's (with same derivnum)

	powerToCoeff := make(map[float64]float64)

	isX := false
	var x *X
	isY := false
	var y *Y
	for _, expr := range a.es {
		c, okc := expr.(*Constant)
		x1, okx := expr.(*X)
		y1, oky := expr.(*Y)
		if okc {
			powerToCoeff[0] += c.num
			continue
		}
		if okx {
			if isY {
				return nil, false
			}
			isX = true
			x = x1
			powerToCoeff[1] += 1
			continue
		}
		if oky {
			if isX {
				return nil, false
			}
			if y != nil && y.derivnum != y1.derivnum {
				return nil, false
			}
			isY = true
			y = y1
			powerToCoeff[1] += 1
			continue
		}
		// TODO: merge polynomials?
		return nil, false
	}
	if isX {
		return poly(powerToCoeff, x), true
	}
	if isY {
		return poly(powerToCoeff, y), true
	}
	return nil, false
}

func addPolysAndVars(es []Expression) []Expression {
	// split into polynomials, x, y and the rest
	polys, rest := splitByType[*Polynomial](es)
	var xs []*X
	xs, rest = splitByType[*X](rest)

	var ys []*Y
	ys, rest = splitByType[*Y](rest)

	for i := 0; i < len(ys); i++ {
		if ys[i].derivnum != 0 {
			rest = append(rest, ys[i])
			ys = append(ys[:i], ys[i+1:]...)
			i--
		}
	}

	xPowerCoeffs := make(map[float64]float64)
	yPowerCoeffs := make(map[float64]float64)

	for _, p := range polys {
		_, isX := p.inside.(*X)
		yin, isY := p.inside.(*Y)
		if isY && yin.derivnum != 0 {
			isY = false // ignore y' y'' etc for now at least
		}
		if !isX && !isY {
			rest = append(rest, p)
			continue
		}
		for power, coeff := range p.powerToCoeff {
			if isX {
				xPowerCoeffs[power] += coeff
				//fmt.Println("added to x", power, coeff)
			} else if isY {
				yPowerCoeffs[power] += coeff
				//fmt.Println("added to y", power, coeff)
			}
		}
	}

	if len(xs) > 0 {
		if len(xPowerCoeffs) == 0 {
			rest = append(rest, mul(num(float64(len(xs))), x()))
		} else {
			xPowerCoeffs[1] += float64(len(xs))
		}
	}
	if len(ys) > 0 {
		if len(yPowerCoeffs) == 0 {
			rest = append(rest, mul(num(float64(len(ys))), y()))
		} else {
			yPowerCoeffs[1] += float64(len(ys))
		}
	}

	if len(xPowerCoeffs) > 0 {
		rest = append(rest, poly(xPowerCoeffs, x()).simplify())
	}
	if len(yPowerCoeffs) > 0 {
		rest = append(rest, poly(yPowerCoeffs, y()).simplify())
	}
	return rest
}

func (e *ExprsAdded) Derivative() Expression {
	simp := e.simplify()
	if simp != e {
		return simp.Derivative()
	}
	newEs := make([]Expression, 0, len(e.es))
	for _, expr := range e.es {
		if _, ok := expr.(*Constant); ok {
			continue
		}
		newEs = append(newEs, expr.Derivative())
	}
	return add(newEs...)
}

func (e *ExprsAdded) String() (str string) {
	//simp := e.simplify()
	//if simp != e {
	//	return simp.String()
	//}
	//if debug {
	//	defer func() {
	//		str = "add[" + str + "]"
	//	}()
	//}
	//
	//expr1str := e.expr1.String()
	//expr2str := e.expr2.String()
	//
	//switch e.expr1.(type) {
	//case *ExprsSubtracted, *ExprsAdded:
	//	expr1str = expr1str[1 : len(expr1str)-1]
	//}
	//switch e.expr2.(type) {
	//case *ExprsSubtracted, *ExprsAdded:
	//	expr2str = expr2str[1 : len(expr2str)-1]
	//}
	//
	////return "[" + e.expr1.String() + " + " + e.expr2.String() + "]"
	//return "(" + expr1str + " + " + expr2str + ")"

	simp := e.simplify()
	if simp != e {
		return simp.String()
	}
	if debug {
		defer func() {
			str = "add[" + str + "]"
		}()
	}
	result := strings.Builder{}
	result.WriteRune('(')
	for i, expr := range e.es {
		if i == 0 {
			result.WriteString(expr.String())
		} else {
			result.WriteString(" + ")
			result.WriteString(expr.String())
		}
	}
	result.WriteRune(')')
	return result.String()
}

func add(es ...Expression) Expression {
	return &ExprsAdded{es}
}

func (e *ExprsAdded) structure() string {
	result := ""
	for _, expr := range e.es {
		result += expr.structure() + " + "
	}
	return "add{" + result[:len(result)-3] + "}"
}
