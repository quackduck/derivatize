package main

type ExprsSubtracted struct {
	expr1 Expression
	expr2 Expression
}

func (e *ExprsSubtracted) simplify() Expression {
	e.expr1 = e.expr1.simplify()
	e.expr2 = e.expr2.simplify()
	c1, okc1 := e.expr1.(*Constant)
	c2, okc2 := e.expr2.(*Constant)
	if okc1 || okc2 {
		if okc1 && okc2 {
			return num(c1.num - c2.num)
		}
		if okc1 {
			if c1.num == 0 {
				return mul(num(-1), e.expr2)
			}
		}
		if okc2 {
			if c2.num == 0 {
				return e.expr1
			}
		}
	}
	return e
}

func (e *ExprsSubtracted) Derivative() Expression {
	simp := e.simplify()
	if simp != e {
		return simp.Derivative()
	}
	return subtract(e.expr1.Derivative(), e.expr2.Derivative())
}

func (e *ExprsSubtracted) String() (str string) {
	simp := e.simplify()
	if simp != e {
		return simp.String()
	}
	if debug {
		defer func() {
			str = "sub[" + str + "]"
		}()
	}
	expr1str := e.expr1.String()
	switch e.expr1.(type) {
	case *ExprsAdded, *ExprsSubtracted, *Polynomial:
		expr1str = expr1str[1 : len(expr1str)-1]
	}
	return "(" + expr1str + " - " + e.expr2.String() + ")"
}

func subtract(e1 Expression, e2 Expression) Expression { return &ExprsSubtracted{e1, e2} }

func (e *ExprsSubtracted) structure() string {
	return "sub{" + e.expr1.structure() + " - " + e.expr2.structure() + "}"
}
