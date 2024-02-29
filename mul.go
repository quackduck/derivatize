package main

import "strings"

type ExprsMultiplied struct {
	es []Expression
}

func (e *ExprsMultiplied) simplify() (ret Expression) {
	if len(e.es) == 1 {
		return e.es[0].simplify()
	}
	if len(e.es) == 0 {
		return e
	}

	merged := make([]Expression, 0, len(e.es))
	// combine constants
	constant := 1.0
	constCount := 0
	newTerms := 0
	//fmt.Println("es", e.es)
	for _, expr := range e.es {
		if mult2, ok := expr.(*ExprsMultiplied); ok {
			for i, expr2 := range mult2.es {
				mult2.es[i] = expr2.simplify()
			}
			merged = append(merged, mult2.es...)
			newTerms += len(mult2.es) - 1
		} else {
			merged = append(merged, expr.simplify())
		}
	}

	merged = mergeBasedOnReflect(merged, true)
	if len(e.es) == 1 {
		return e.es[0].simplify()
	}

	noConsts := make([]Expression, 0, len(merged))
	for _, expr := range merged {
		if c, ok := expr.(*Constant); ok {
			if c.num == 0 {
				return &Constant{0}
			}
			constant *= c.num
			// TODO: if constants are too big, don't multiply them
			constCount++
		} else {
			noConsts = append(noConsts, expr)
		}
	}
	if len(noConsts) == 0 {
		return num(constant)
	}
	//noConsts = mulPolysAndVars(noConsts, &Constant{constant})
	noConsts = mulMergeDivides(noConsts)
	if constant != 1.0 {
		noConsts = append([]Expression{&Constant{constant}}, noConsts...)
	}
	e.es = noConsts
	return e
}

//func mulPolysAndVars(es []Expression, c *Constant) (rest []Expression) {
//	es = mulMergeDivides(es)
//
//	// split into polynomials, x, y and the rest
//	var polys []*Polynomial
//	polys, rest = splitByType[*Polynomial](es)
//
//	if len(polys) != 0 && c.num != 1.0 {
//		for power, coeff := range polys[0].powerToCoeff {
//			polys[0].powerToCoeff[power] = coeff * c.num
//		}
//	} else if len(polys) == 0 && c.num != 1.0 {
//		defer func() {
//			rest = append([]Expression{c}, rest...)
//		}()
//	}
//
//	var xs []*X
//	xs, rest = splitByType[*X](rest)
//	var ys []*Y
//	ys, rest = splitByType[*Y](rest)
//
//	for i := 0; i < len(ys); i++ {
//		if ys[i].derivnum != 0 {
//			rest = append(rest, ys[i])
//			ys = append(ys[:i], ys[i+1:]...)
//			i--
//		}
//	}
//
//	if len(xs) == 0 && len(ys) == 0 {
//		for _, p := range polys {
//			rest = append(rest, p)
//		}
//		return rest
//	}
//
//	xdone := false
//	ydone := false
//
//	if len(xs) == 0 {
//		xdone = true
//	}
//	if len(ys) == 0 {
//		ydone = true
//	}
//
//	for _, p := range polys {
//		_, isX := p.inside.(*X)
//		yin, isY := p.inside.(*Y)
//		if isY && yin.derivnum != 0 {
//			isY = false // ignore y' y'' etc for now at least
//		}
//		if !isX && !isY {
//			//rest = append(rest, p)
//			continue
//		}
//		newPowerToCoeff := make(map[float64]float64, len(p.powerToCoeff))
//		for power := range p.powerToCoeff {
//			if isX {
//				//p.powerToCoeff[power] += float64(len(xs))
//				newPowerToCoeff[power+float64(len(xs))] = p.powerToCoeff[power]
//				//fmt.Println("added to x", power, coeff)
//			} else if isY {
//				//p.powerToCoeff[power] += float64(len(xs))
//				newPowerToCoeff[power+float64(len(ys))] = p.powerToCoeff[power]
//				//fmt.Println("added to y", power, coeff)
//			}
//		}
//		p.powerToCoeff = newPowerToCoeff
//		xdone = xdone || isX
//		ydone = ydone || isY
//		if xdone && ydone {
//			break
//		}
//	}
//	//rest = append(rest, polys...)
//	if !xdone {
//		if len(xs) == 1 {
//			rest = append(rest, x())
//		} else {
//			rest = append(rest, poly(map[float64]float64{float64(len(xs)): 1}, x()))
//		}
//	}
//	if !ydone {
//		if len(ys) == 1 {
//			rest = append(rest, y())
//		} else {
//			rest = append(rest, poly(map[float64]float64{float64(len(ys)): 1}, y()))
//		}
//	}
//	for i := range polys {
//		rest = append(rest, polys[i])
//	}
//	return rest
//}

func mulMergeDivides(es []Expression) []Expression {
	divs, rest := splitByType[*ExprsDivided](es)
	if len(divs) == 0 {
		return es
	}
	allhighs := make([]Expression, 0, len(divs))
	alllows := make([]Expression, 0, len(divs))
	for i := range divs {
		allhighs = append(allhighs, divs[i].high)
		alllows = append(alllows, divs[i].low)
	}
	return append(rest, div(mul(allhighs...).simplify(), mul(alllows...).simplify()).simplify())
}

// splitByType splits a list of expressions based on an input type. this is a generic function
func splitByType[T Expression](es []Expression) (split []T, rest []Expression) {
	for _, e := range es {
		if e2, ok := any(e).(T); ok {
			split = append(split, e2)
		} else {
			rest = append(rest, e)
		}
	}
	return
}

func (e *ExprsMultiplied) Derivative() Expression {
	simp := e.simplify()
	if simp != e {
		return simp.Derivative()
	}
	// for f(x)g(x)h(x) the derivative is
	// f′(x)g(x)h(x)+f(x)g′(x)h(x)+f(x)g(x)h′(x)
	// now generalize to n functions
	result := make([]Expression, 0, len(e.es))
	for i := range e.es {
		if i == 0 {
			result = append(result, mul(e.es[i].Derivative(), mul(e.es[i+1:]...)))
			continue
		}
		if i == len(e.es)-1 {
			result = append(result, mul(mul(e.es[:i]...), e.es[i].Derivative()))
			continue
		}
		result = append(result,
			mul(
				mul(e.es[:i]...),
				e.es[i].Derivative(),
				mul(e.es[i+1:]...),
			),
		)
	}
	if len(result) == 1 {
		return result[0]
	}
	return add(result...).simplify()
}

func (e *ExprsMultiplied) String() (str string) {
	simp := e.simplify().simplify() // TODO: fix this
	if simp != e {
		return simp.String()
	}
	if debug {
		defer func() {
			str = "mult[" + str + "]"
		}()
	}
	result := strings.Builder{}
	for i, expr := range e.es {
		if i == 0 {
			result.WriteString(expr.String())
		} else {
			result.WriteString(" * ")
			result.WriteString(expr.String())
		}
	}
	return result.String()
}

func mul(es ...Expression) Expression {
	return &ExprsMultiplied{es}
}

func (e *ExprsMultiplied) structure() string {
	result := ""
	for i, expr := range e.es {
		if i == 0 {
			result += expr.structure()
			continue
		}
		result += " * " + expr.structure()
	}
	return "mul{" + result + "}"
}
