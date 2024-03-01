package main

import "strings"

type Constant struct{ num float64 }

func (n *Constant) Derivative() Expression { return &Constant{0} }
func (n *Constant) String() string         { return ftoa(n.num) }
func (n *Constant) simplify() Expression   { return n }
func num(num float64) *Constant            { return &Constant{num} }
func (n *Constant) structure() string      { return "const{" + n.String() + "}" }

type Var struct {
	name     string
	wrt      bool
	implicit bool
	derivnum int
}

func (v *Var) Derivative() Expression {
	if v.implicit {
		return &Var{name: v.name, wrt: v.wrt, implicit: true, derivnum: v.derivnum + 1}
	}
	if v.wrt {
		return &Constant{1}
	} else {
		return &Constant{0}
	}
}

func (v *Var) String() string {
	if v.implicit {
		return v.name + strings.Repeat("'", v.derivnum)
	}
	return v.name
}

func (v *Var) simplify() Expression { return v }
func (v *Var) structure() string    { return "var{" + v.String() + "}" }

//type X struct{}
//
//func (x *X) Derivative() Expression { return &Constant{1} }
//func (x *X) String() string         { return "x" }
//func (x *X) simplify() Expression   { return x }
//func x() Expression                 { return &X{} }
//func (x *X) structure() string      { return "x{}" }
//
//type Y struct{ derivnum int }
//
//func (y *Y) Derivative() Expression { return &Y{y.derivnum + 1} }
//func (y *Y) String() string         { return "y" + strings.Repeat("'", y.derivnum) }
//func (y *Y) simplify() Expression   { return y }
//func y() Expression                 { return &Y{} }
//func (y *Y) structure() string      { return "y{}" }
//
//type B struct{}
//
//func (b *B) Derivative() Expression { return &Constant{1} }
//func (b *B) String() string         { return "b" }
//func (b *B) simplify() Expression   { return b }
//func b() Expression                 { return &B{} }
//func (b *B) structure() string      { return "b{}" }
//
//type C struct{}
//
//func (c *C) Derivative() Expression { return &Constant{0} }
//func (c *C) String() string         { return "c" }
//func (c *C) simplify() Expression   { return c }
//func c() Expression                 { return &C{} }
//func (c *C) structure() string      { return "c{}" }
