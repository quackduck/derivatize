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
	e.es = mergeBasedOnReflect(e.es, false)
	if len(e.es) == 1 {
		return e.es[0].simplify()
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

	if len(noConsts) == 0 {
		return num(constant)
	}
	if constant != 0.0 {
		noConsts = append([]Expression{&Constant{constant}}, noConsts...)
	}
	e.es = noConsts
	return e
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
