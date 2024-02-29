package main

type Sin struct {
	expr Expression
}

func (s *Sin) simplify() Expression {
	s.expr = s.expr.simplify()
	return s
}

func (s *Sin) Derivative() Expression {
	simp := s.simplify()
	if simp != s {
		return simp.Derivative()
	}
	return mul(cos(s.expr), s.expr.Derivative())
}

func (s *Sin) String() string {
	simp := s.simplify()
	if simp != s {
		return simp.String()
	}
	insideStr := s.expr.String()
	//if _, ok := s.expr.(*Polynomial); !ok {
	insideStr = "[" + insideStr + "]"
	//}
	return "sin" + insideStr
}

func sin(expr Expression) *Sin { return &Sin{expr} }

func (s *Sin) structure() string {
	return "sin{" + s.expr.structure() + "}"
}

type Cos struct {
	expr Expression
}

func (c *Cos) simplify() Expression {
	c.expr = c.expr.simplify()
	return c
}

func (c *Cos) Derivative() Expression {
	simp := c.simplify()
	if simp != c {
		return simp.Derivative()
	}
	return mul(num(-1), sin(c.expr), c.expr.Derivative())
}

func (c *Cos) String() string {
	simp := c.simplify()
	if simp != c {
		return simp.String()
	}
	insideStr := c.expr.String()
	//if _, ok := c.expr.(*Polynomial); !ok {
	insideStr = "[" + insideStr + "]"
	//}
	return "cos" + insideStr
}

func cos(expr Expression) *Cos { return &Cos{expr} }

func (c *Cos) structure() string {
	return "cos{" + c.expr.structure() + "}"
}
