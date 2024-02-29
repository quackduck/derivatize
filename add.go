package main

import (
	"strings"
)

type ExprsAdded struct {
	es []Expression
}

func (e *ExprsAdded) simplify() Expression {
	if len(e.es) == 1 {
		return e.es[0].simplify()
	}
	//fmt.Println("simplifying add", e.structure())
	e.es = mergeBasedOnReflect(e.es, false)
	//fmt.Println("simplifying add", e.structure())
	if len(e.es) == 1 {
		//fmt.Println("whew...")
		return e.es[0].simplify()
	}
	//if p, ok := checkIfCanBecomePolynomial(e); ok {
	//	return p.simplify()
	//}

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
