package main

import (
	"fmt"
	"math"
	"slices"
	"strconv"
)

var debug = false

type Expression interface {
	Derivative() Expression
	simplify() Expression
	String() string
}

type Constant struct{ num float64 }

func (n *Constant) Derivative() Expression { return &Constant{0} }
func (n *Constant) String() string         { return ftoa(n.num) }
func (n *Constant) simplify() Expression   { return n }

type X struct{}

func (x *X) Derivative() Expression { return &Constant{1} }
func (x *X) String() string         { return "x" }
func (x *X) simplify() Expression   { return x }

type ExprsMultiplied struct {
	expr1 Expression
	expr2 Expression
}

func (e *ExprsMultiplied) simplify() Expression {
	e.expr1 = e.expr1.simplify()
	e.expr2 = e.expr2.simplify()
	c1, okc1 := e.expr1.(*Constant)
	//fmt.Println("c1", c1, "okc1", okc1, "high", e.high)
	c2, okc2 := e.expr2.(*Constant)
	//fmt.Println("c2", c2, "okc2", okc2, "low", e.low)
	if okc1 || okc2 {
		if okc1 && okc2 {
			return &Constant{c1.num * c2.num}
		}
		var constant *Constant
		var expr Expression
		if okc1 {
			constant = c1
			expr = e.expr2
		} else {
			constant = c2
			expr = e.expr1
		}
		if mul, ok := expr.(*ExprsMultiplied); ok {
			if c2, ok := mul.expr1.(*Constant); ok {
				return &ExprsMultiplied{expr1: &Constant{constant.num * c2.num}, expr2: mul.expr2}
			}
		}
		if constant.num == 0 {
			return constant
		}
		if constant.num == 1 {
			return expr
		}
	}
	return e
}

func (e *ExprsMultiplied) Derivative() Expression {
	simp := e.simplify()
	//fmt.Println("simplified", simp, "original", e)
	if simp != e {
		return simp.Derivative()
	}
	if _, ok := e.expr1.(*Constant); ok {
		return &ExprsMultiplied{expr1: e.expr1, expr2: e.expr2.Derivative()}
	}
	if _, ok := e.expr2.(*Constant); ok {
		return &ExprsMultiplied{expr1: e.expr2, expr2: e.expr1.Derivative()}
	}
	return &ExprsAdded{
		expr1: &ExprsMultiplied{expr1: e.expr1.Derivative(), expr2: e.expr2},
		expr2: &ExprsMultiplied{expr1: e.expr1, expr2: e.expr2.Derivative()},
	}
}

func (e *ExprsMultiplied) String() (str string) {
	simp := e.simplify()
	if simp != e {
		return simp.String()
	}
	if debug {
		defer func() {
			str = "mult[" + str + "]"
		}()
	}
	expr1str := e.expr1.String()
	expr2str := e.expr2.String()
	switch e.expr1.(type) {
	case *ExprsAdded, *ExprsSubtracted:
		expr1str = "(" + expr1str + ")"
	}
	switch e.expr2.(type) {
	case *ExprsAdded, *ExprsSubtracted:
		expr2str = "(" + expr2str + ")"
	}
	return expr1str + " * " + expr2str
}

type ExprsAdded struct {
	expr1 Expression
	expr2 Expression
}

func (e *ExprsAdded) simplify() Expression {
	e.expr1 = e.expr1.simplify()
	e.expr2 = e.expr2.simplify()
	c1, okc1 := e.expr1.(*Constant)
	c2, okc2 := e.expr2.(*Constant)
	if okc1 || okc2 {
		var constant *Constant
		var expr Expression
		if okc1 {
			constant = c1
			expr = e.expr2
		} else {
			constant = c2
			expr = e.expr1
		}
		if constant.num == 0 {
			return expr
		}
		if okc1 && okc2 {
			return &Constant{c1.num + c2.num}
		}
	}
	return e
}

func (e *ExprsAdded) Derivative() Expression {
	simp := e.simplify()
	if simp != e {
		return simp.Derivative()
	}
	if _, ok := e.expr1.(*Constant); ok {
		if e.expr1.(*Constant).num == 0 {
			return e.expr2.Derivative()
		}
	}
	if _, ok := e.expr2.(*Constant); ok {
		if e.expr2.(*Constant).num == 0 {
			return e.expr1.Derivative()
		}
	}
	return &ExprsAdded{expr1: e.expr1.Derivative(), expr2: e.expr2.Derivative()}
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

	if _, ok := e.expr1.(*Constant); ok {
		if e.expr1.(*Constant).num == 0 {
			return e.expr2.String()
		}
	}
	if _, ok := e.expr2.(*Constant); ok {
		if e.expr2.(*Constant).num == 0 {
			return e.expr1.String()
		}
	}
	return e.expr1.String() + " + " + e.expr2.String()
}

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
			return &Constant{c1.num - c2.num}
		}
		if okc1 {
			if c1.num == 0 {
				return &ExprsMultiplied{&Constant{-1}, e.expr2}
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
	return &ExprsSubtracted{expr1: e.expr1.Derivative(), expr2: e.expr2.Derivative()}
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
	expr2Str := e.expr2.String()
	switch e.expr2.(type) {
	case *ExprsAdded, *ExprsSubtracted:
		expr2Str = "(" + expr2Str + ")"
	}
	return e.expr1.String() + " - " + expr2Str
}

type ExprsDivided struct {
	high Expression
	low  Expression
}

func (e *ExprsDivided) simplify() Expression {
	e.high = e.high.simplify()
	e.low = e.low.simplify()
	c1, okc1 := e.high.(*Constant)
	c2, okc2 := e.low.(*Constant)
	if okc1 && okc2 {
		return &Constant{c1.num / c2.num}
	}
	if okc1 {
		if c1.num == 0 {
			return &Constant{0}
		}
		if polyn, ok := e.low.(*Polynomial); ok { // constant divided by polynomial
			return (&Polynomial{map[float64]float64{-1: c1.num}, polyn}).simplify()
		}
	}
	return e
}

func (e *ExprsDivided) Derivative() Expression {
	simp := e.simplify()
	if simp != e {
		return simp.Derivative()
	}
	return &ExprsDivided{
		high: &ExprsSubtracted{
			expr1: &ExprsMultiplied{expr1: e.low, expr2: e.high.Derivative()},
			expr2: &ExprsMultiplied{expr1: e.high, expr2: e.low.Derivative()},
		},
		low: &Polynomial{powerToCoeff: map[float64]float64{2: 1}, inside: e.low},
	}
}

func (e *ExprsDivided) String() (str string) {
	simp := e.simplify()
	if simp != e {
		return simp.String()
	}
	if debug {
		defer func() {
			str = "div[" + str + "]"
		}()
	}
	highStr := e.high.String()
	lowStr := e.low.String()
	switch e.high.(type) {
	case *ExprsAdded, *ExprsSubtracted:
		highStr = "(" + highStr + ")"
	}
	switch e.low.(type) {
	case *ExprsAdded, *ExprsSubtracted, *ExprsMultiplied, *ExprsDivided:
		lowStr = "(" + lowStr + ")"
	}
	return highStr + " / " + lowStr
}

type Polynomial struct {
	powerToCoeff map[float64]float64
	inside       Expression
}

func (p *Polynomial) simplify() Expression {
	p.inside = p.inside.simplify()
	if len(p.powerToCoeff) == 1 {
		for power, coeff := range p.powerToCoeff {
			if power == 0 {
				return &Constant{coeff}
			}
			if power == 1 {
				return &ExprsMultiplied{expr1: &Constant{coeff}, expr2: p.inside}
			}
		}
	}
	if polyn, ok := p.inside.(*Polynomial); ok {
		if len(polyn.powerToCoeff) == 1 && len(p.powerToCoeff) == 1 {
			for inpower, incoeff := range polyn.powerToCoeff {
				for outpower, outcoeff := range p.powerToCoeff {
					return &Polynomial{map[float64]float64{inpower * outpower: outcoeff * math.Pow(incoeff, outpower)}, polyn.inside}
				}
			}
		}
	}
	return p
}

func (p *Polynomial) Derivative() Expression {
	simp := p.simplify()
	if simp != p {
		return simp.Derivative()
	}
	if _, ok := p.inside.(*Constant); ok {
		return &Constant{0}
	}
	if len(p.powerToCoeff) == 1 {
		for power, coeff := range p.powerToCoeff {
			if power == 0 {
				return &Constant{0}
			}
			if power == 1 {
				return &ExprsMultiplied{expr1: &Constant{coeff}, expr2: p.inside.Derivative()}
			}
		}
	}
	derivative := &Polynomial{make(map[float64]float64, len(p.powerToCoeff)), p.inside}
	for power, coeff := range p.powerToCoeff {
		if power != 0 {
			derivative.powerToCoeff[power-1] = power * coeff
		}
	}
	return &ExprsMultiplied{expr1: derivative, expr2: p.inside.Derivative()}
}

func (p *Polynomial) String() (str string) {
	if debug {
		defer func() {
			str = "polyn[" + str + "]"
		}()
	}

	insideStr := p.inside.String()
	if _, ok := p.inside.(*Constant); ok {
		if p.inside.(*Constant).num == 0 {
			return "0"
		}
		insideStr = "(" + insideStr + ")"
	}
	// Sort the powers in ascending order
	powers := make([]float64, 0, len(p.powerToCoeff))
	for power := range p.powerToCoeff {
		powers = append(powers, power)
	}
	slices.Sort(powers)
	for _, power := range powers {
		coeff := p.powerToCoeff[power]
		if coeff == 0 {
			continue
		}
		coeffStr := ""
		if coeff != 1 {
			coeffStr = ftoa(coeff)
		}
		if power == 0 {
			if coeffStr == "" {
				coeffStr = "1"
			}
			str += coeffStr + " + "
		} else if power == 1 {
			str += coeffStr + "" + insideStr + " + "
		} else {
			str += coeffStr + "" + insideStr + "^" + ftoa(power) + " + "
		}
	}
	// Remove the trailing " + "
	if len(str) > 0 {
		str = str[:len(str)-3]
	}
	if len(p.powerToCoeff) == 1 { // if there's only a single term don't wrap it in brackets
		return str
	}
	return "(" + str + ")"
}

type Log struct {
	base float64
	expr Expression
}

func (l *Log) simplify() Expression {
	l.expr = l.expr.simplify()
	//if c, ok := l.expr.(*Constant); ok {
	//	return &Constant{math.Log(c.num) / math.Log(l.base)}
	//}
	return l
}

var E = math.E

func (l *Log) Derivative() Expression {
	simp := l.simplify()
	if simp != l {
		return simp.Derivative()
	}
	if l.base == E {
		return &ExprsDivided{l.expr.Derivative(), l.expr}
	}
	return &ExprsDivided{
		l.expr.Derivative(),
		&ExprsMultiplied{
			l.expr,
			&Log{E, &Constant{l.base}},
		},
	}
}

func (l *Log) String() string {
	simp := l.simplify()
	if simp != l {
		return simp.String()
	}
	if l.base == E {
		return "ln(" + l.expr.String() + ")"
	}
	return "log_" + ftoa(l.base) + "(" + l.expr.String() + ")"
}

type Exponential struct {
	base  float64
	power Expression
}

func (e *Exponential) simplify() Expression {
	e.power = e.power.simplify()
	if c, ok := e.power.(*Constant); ok {
		return &Constant{math.Pow(e.base, c.num)}
	}
	return e
}

func (e *Exponential) Derivative() Expression {
	simp := e.simplify()
	if simp != e {
		return simp.Derivative()
	}
	if e.base == E {
		return &ExprsMultiplied{
			&Exponential{E, e.power},
			e.power.Derivative(),
		}
	}
	return &ExprsMultiplied{
		expr1: &ExprsMultiplied{
			&Log{base: E, expr: &Constant{e.base}},
			&Exponential{
				base:  e.base,
				power: e.power,
			},
		},
		expr2: e.power.Derivative(),
	}
}

func (e *Exponential) String() string {
	simp := e.simplify()
	if simp != e {
		return simp.String()
	}
	if e.base == E {
		return "e^(" + e.power.String() + ")"
	}
	return ftoa(e.base) + "^(" + e.power.String() + ")"
}

func main() {
	//p := &Polynomial{map[float64]float64{0: 1, 1: 2, 2: 3},
	//	&Polynomial{map[float64]float64{30: 1, 31: 1}, &X{}},
	//}
	//p := &Polynomial{map[float64]float64{2: 1, 3: 1, 4: 0},
	//	&Polynomial{map[float64]float64{30: 1, 31: 1}, &X{}},
	//}
	// 1 + 2(x+3) + 3(x+3)^2
	//p := &ExprsDivided{
	//	&Polynomial{
	//		powerToCoeff: map[float64]float64{0: 3, 5: 1},
	//		inside:       &X{},
	//	},
	//	&Polynomial{
	//		powerToCoeff: map[float64]float64{-1: 3, 5: 1},
	//		inside:       &X{},
	//	},
	//}
	p := &Exponential{
		base: 3,
		//power: &Log{base: 2, expr: &X{}},
		//power: &X{},
		power: &Polynomial{
			powerToCoeff: map[float64]float64{0: 3, 5: 1},
			inside:       &X{},
		},
	}
	//fmt.Println(p)
	//fmt.Println(p.Derivative())
	//fmt.Println(p.Derivative().Derivative())
	//fmt.Println(p.Derivative().Derivative().Derivative())
	//fmt.Println(p.Derivative().Derivative().Derivative().Derivative())

	//p := &ExprsMultiplied{
	//	&Polynomial{map[int]int{0: 1}, &X{}},
	//	&Polynomial{map[int]int{30: 1}, &X{}},
	//}
	fmt.Println(p)
	fmt.Println(p.Derivative())
	//legible(p.Derivative().String())
	fmt.Println(p.Derivative().Derivative())
	//legible(p.Derivative().Derivative().String())
	//legible(p.Derivative().Derivative().String())
	//fmt.Println(p.Derivative().Derivative())
	//fmt.Println(p.Derivative().Derivative().Derivative())
	//fmt.Println(p.Derivative().Derivative().Derivative().Derivative())

	// 1 + 2(x+2) + 3(x+3)^2
	//p := &Polynomial{map[int]int{0: 1},
	//	&X{},
	//}
	//fmt.Println(p)
	//fmt.Println(p.Derivative())
	//fmt.Println(p.Derivative().Derivative())
	//fmt.Println(p.Derivative().Derivative().Derivative())

	//p := &ExprsMultiplied{&Constant{6}, &Polynomial{
	//	powerToCoeff: map[int]int{0: 1, 1: 2, 2: 3},
	//	inside:       &X{},
	//}}
	//fmt.Println(p)
	//fmt.Println(p.Derivative())
	//fmt.Println(p.Derivative().Derivative())
}

func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func minusMulSimplify(e *ExprsMultiplied) (result string, negative bool) {
	c, expr, ok := getConstExpr(e.expr1, e.expr2)
	if ok {
		if c.num == 1 {
			return expr.String(), false
		}
		if c.num == -1 {
			return expr.String(), true
		}
	}
	return e.String(), false
}

func getConstExpr(e1 Expression, e2 Expression) (c *Constant, expr Expression, ok bool) {
	c1, okc1 := e1.(*Constant)
	c2, okc2 := e2.(*Constant)
	if okc1 {
		return c1, e2, true
	}
	if okc2 {
		return c2, e1, true
	}
	return nil, nil, false
}

func legible(e string) {
	// example: (3 + x^5) / (3x^-1 + x^5) + 3(3 + x^0.5 + x)^-20
	// get the maximum depth of the expression

	maxdepth := 0
	currdepth := 0
	for _, c := range e {
		if c == '(' {
			currdepth++
		}
		if c == ')' {
			currdepth--
		}
		if currdepth > maxdepth {
			maxdepth = currdepth
		}
	}
	//fmt.Println("maxdepth", maxdepth)

	// iterate through th expression agaib, this time adding spaces depending on the depth
	currdepth = 0
	newstr := ""
	for i, c := range e {
		if c == '(' {
			currdepth++
		}
		if c == ')' {
			currdepth--
		}
		//fmt.Print(currdepth)

		if currdepth == 0 || currdepth == 1 {
			if (c == '+' || c == '-') && e[i+1] == ' ' {
				newstr += "\n"
			} else {

			}
		}
		newstr += string(c)

		//if c == '+' || c == '-' && e[i+1] == ' ' { // || c == '*' || c == '/'
		//	newstr += strings.Repeat(" ", (maxdepth-currdepth)) + string(c) + strings.Repeat(" ", (maxdepth-currdepth))
		//} else {
		//	newstr += string(c)
		//}
	}
	//fmt.Println()
	fmt.Println(newstr)
}
