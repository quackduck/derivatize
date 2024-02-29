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
