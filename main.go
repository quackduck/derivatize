package main

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

var debug = true

type Expression interface {
	Derivative() Expression
	String() string
	simplify() Expression
}

type Constant struct{ num float64 }

func (n *Constant) Derivative() Expression { return &Constant{0} }
func (n *Constant) String() string         { return ftoa(n.num) }
func (n *Constant) simplify() Expression   { return n }
func num(num float64) *Constant            { return &Constant{num} }

type X struct{}

func (x *X) Derivative() Expression { return &Constant{1} }
func (x *X) String() string         { return "x" }
func (x *X) simplify() Expression   { return x }
func x() *X                         { return &X{} }

type Y struct{ derivnum int }

func (y *Y) Derivative() Expression { return &Y{y.derivnum + 1} }
func (y *Y) String() string         { return "y" + strings.Repeat("'", y.derivnum) }
func (y *Y) simplify() Expression   { return y }
func y() *Y                         { return &Y{} }

type ExprsMultiplied struct {
	es []Expression
}

func (e *ExprsMultiplied) simplify() Expression {
	merged := make([]Expression, 0, len(e.es))
	// combine constants
	constant := 1.0
	changed := false
	constCount := 0
	newTerms := 0
	for _, expr := range e.es {
		if mult2, ok := expr.(*ExprsMultiplied); ok {
			changed = true
			for i, expr2 := range mult2.es {
				mult2.es[i] = expr2.simplify()
			}
			merged = append(merged, mult2.es...)
			newTerms += len(mult2.es) - 1
		} else {
			merged = append(merged, expr.simplify())
		}
	}

	noConsts := make([]Expression, 0, len(merged))
	for _, expr := range merged {
		if c, ok := expr.(*Constant); ok {
			if c.num == 0 {
				return &Constant{0}
			}
			constant *= c.num
			constCount++
		} else {
			noConsts = append(noConsts, expr)
		}
	}
	if len(noConsts) == 0 {
		return num(constant)
	}
	if constCount > 1 {
		changed = true
	}
	if len(e.es) > 0 {
		if c, ok := e.es[0].(*Constant); !ok { // if the first element is not a constant
			if constCount != 0 { // if there are constants
				changed = true // we changed the order of the constants
			}
		} else {
			// if the first element is a constant and is 1 then a change has been made
			if c.num == 1 {
				changed = true
			}
		}
	}
	if constant != 1.0 {
		noConsts = append([]Expression{&Constant{constant}}, noConsts...)
	}
	if changed {
		return &ExprsMultiplied{noConsts}
	} else {
		return e
	}
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
			result = append(result, mult(e.es[i].Derivative(), mult(e.es[i+1:]...)))
			continue
		}
		if i == len(e.es)-1 {
			result = append(result, mult(mult(e.es[:i]...), e.es[i].Derivative()))
			continue
		}
		result = append(result,
			mult(
				mult(e.es[:i]...),
				e.es[i].Derivative(),
				mult(e.es[i+1:]...),
			),
		)
	}
	return add(result...)
}

//func prodRule(e1 Expression, e2 Expression) Expression {
//	return &ExprsAdded{
//		expr1: &ExprsMultiplied{[]Expression{e1, e2.Derivative()}},
//		expr2: &ExprsMultiplied{[]Expression{e2, e1.Derivative()}},
//	}
//}

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
	result := ""
	for i, expr := range e.es {
		if i == 0 {
			result += expr.String()
		} else {
			result += " * " + expr.String()
		}
	}
	return result
}

func mult(es ...Expression) Expression {
	return &ExprsMultiplied{es}
}

type ExprsAdded struct {
	es []Expression
}

func (e *ExprsAdded) simplify() Expression {
	//e.expr1 = e.expr1.simplify()
	//e.expr2 = e.expr2.simplify()
	//c1, okc1 := e.expr1.(*Constant)
	//c2, okc2 := e.expr2.(*Constant)
	//if okc1 || okc2 {
	//	var constant *Constant
	//	var expr Expression
	//	if okc1 {
	//		constant = c1
	//		expr = e.expr2
	//	} else {
	//		constant = c2
	//		expr = e.expr1
	//	}
	//	if constant.num == 0 {
	//		return expr
	//	}
	//	if okc1 && okc2 {
	//		return &Constant{c1.num + c2.num}
	//	}
	//}
	//return e
	merged := make([]Expression, 0, len(e.es))
	// combine constants
	constant := 0.0
	changed := false
	constCount := 0
	newTerms := 0
	for _, expr := range e.es {
		if add2, ok := expr.(*ExprsAdded); ok {
			changed = true
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
	if constCount > 1 {
		changed = true
	}
	if len(e.es) > 0 {
		if c, ok := e.es[0].(*Constant); !ok { // if the first element is not a constant
			if constCount != 0 { // if there are constants
				changed = true // we changed the order of the constants
			}
		} else {
			// if the first element is a constant and is 1 then a change has been made
			if c.num == 0 {
				changed = true
			}
		}
	}
	if constant != 0.0 {
		noConsts = append([]Expression{&Constant{constant}}, noConsts...)
	}
	if changed {
		return &ExprsAdded{noConsts}
	} else {
		return e
	}
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
	result := ""
	for i, expr := range e.es {
		if i == 0 {
			result += expr.String()
		} else {
			result += " + " + expr.String()
		}
	}
	return "(" + result + ")"
}

func add(es ...Expression) Expression {
	//if len(es) == 0 {
	//	return &Constant{0}
	//}
	//if len(es) == 1 {
	//	return es[0]
	//}
	return &ExprsAdded{es}
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
			return num(c1.num - c2.num)
		}
		if okc1 {
			if c1.num == 0 {
				return mult(num(-1), e.expr2)
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
	expr2Str := e.expr2.String()
	return "(" + e.expr1.String() + " - " + expr2Str + ")"
}

func subtract(e1 Expression, e2 Expression) Expression { return &ExprsSubtracted{e1, e2} }

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
		//fmt.Println("c1", c1.num)
		if c1.num == 0 {
			//fmt.Println("zero")
			return &Constant{0}
		}
		if polyn, ok := e.low.(*Polynomial); ok { // constant divided by polynomial
			return (poly(map[float64]float64{-1: c1.num}, polyn)).simplify()
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
			expr1: mult(e.low, e.high.Derivative()),
			expr2: mult(e.high, e.low.Derivative()),
		},
		low: poly(map[float64]float64{2: 1}, e.low),
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
	//switch e.high.(type) {
	//case *ExprsAdded, *ExprsSubtracted:
	//	highStr = "(" + highStr + ")"
	//}
	switch e.low.(type) {
	//case *ExprsAdded, *ExprsSubtracted, *ExprsMultiplied, *ExprsDivided:
	case *ExprsMultiplied, *ExprsDivided:
		lowStr = "[ " + lowStr + " ]"
	}
	return highStr + " / " + lowStr
}

func div(e1 Expression, e2 Expression) Expression { return &ExprsDivided{e1, e2} }

type Polynomial struct {
	powerToCoeff map[float64]float64
	inside       Expression
}

func (p *Polynomial) simplify() Expression {
	p.inside = p.inside.simplify()
	if len(p.powerToCoeff) == 1 {
		for power, coeff := range p.powerToCoeff {
			if power == 0 {
				return num(coeff)
			}
			if power == 1 {
				return mult(num(coeff), p.inside)
			}
		}
	}
	if polyn, ok := p.inside.(*Polynomial); ok {
		if len(polyn.powerToCoeff) == 1 && len(p.powerToCoeff) == 1 {
			for inpower, incoeff := range polyn.powerToCoeff {
				for outpower, outcoeff := range p.powerToCoeff {
					return poly(map[float64]float64{inpower * outpower: outcoeff * math.Pow(incoeff, outpower)}, polyn.inside)
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
		return num(0)
	}
	if len(p.powerToCoeff) == 1 {
		for power, coeff := range p.powerToCoeff {
			if power == 0 {
				return num(0)
			}
			if power == 1 {
				return mult(num(coeff), p.inside.Derivative())
			}
		}
	}
	derivative := poly(make(map[float64]float64, len(p.powerToCoeff)), p.inside)
	for power, coeff := range p.powerToCoeff {
		if power != 0 {
			derivative.powerToCoeff[power-1] = power * coeff
		}
	}
	return mult(derivative, p.inside.Derivative())
}

func (p *Polynomial) String() (str string) {
	if debug {
		defer func() {
			str = "polyn[" + str + "]"
		}()
	}

	insideStr := p.inside.String()
	if c, ok := p.inside.(*Constant); ok {
		if c.num == 0 {
			return "0"
		}
	}
	switch p.inside.(type) {
	case *ExprsMultiplied, *ExprsDivided:
		insideStr = "[ " + insideStr + " ]"
	}
	// Sort the powers in ascending order
	powers := make([]float64, 0, len(p.powerToCoeff))
	for power := range p.powerToCoeff {
		powers = append(powers, power)
	}
	slices.Sort(powers)
	slices.Reverse(powers)
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

func poly(powerToCoeff map[float64]float64, inside Expression) *Polynomial {
	return &Polynomial{powerToCoeff, inside}
}

func polyParse(ps string, inside Expression) *Polynomial {
	powerToCoeff := make(map[float64]float64)
	terms := strings.Split(ps, "+")
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		coeff := 1.0
		power := 1.0
		if !strings.Contains(term, "x") {
			power = 0
		}
		var err error
		split := strings.Split(term, "x")
		fmt.Println(split)
		coeffStr := strings.TrimSpace(split[0])
		powerStr := ""
		if len(split) > 1 {
			powerStr = strings.TrimSpace(strings.Replace(split[1], "^", "", 1))
		}
		if coeffStr != "" {
			coeff, err = strconv.ParseFloat(coeffStr, 64)
			if err != nil {
				panic("invalid polyParse: " + term)
			}
		}
		if powerStr != "" {
			power, err = strconv.ParseFloat(powerStr, 64)
			if err != nil {
				panic("invalid polyParse: " + term)
			}
		}
		powerToCoeff[power] += coeff
	}
	return poly(powerToCoeff, inside)
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
		return div(l.expr.Derivative(), l.expr).simplify()
	}

	return div(
		l.expr.Derivative(),
		mult(
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
	if _, ok := l.expr.(*Polynomial); !ok {
		if _, ok := l.expr.(*Constant); ok {
			insideStr = "(" + insideStr + ")"
		} else {
			insideStr = "[ " + insideStr + " ]"
		}
	}
	if l.base == E {
		return "ln" + insideStr
	}
	return "log_" + ftoa(l.base) + insideStr
}

func log(base float64, expr Expression) *Log { return &Log{base, expr} }

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
		return mult(
			exp(E, e.power),
			e.power.Derivative(),
		)
	}
	return mult(
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
	if _, ok := e.power.(*Polynomial); !ok {
		insideStr = "[ " + insideStr + " ]"
	}
	if e.base == E {
		return "e^" + insideStr
	}
	return ftoa(e.base) + "^" + insideStr
}

func exp(base float64, power Expression) *Exponential { return &Exponential{base, power} }

func main() {
	p := add(
		x(),
		exp(E, x()),
		mult(num(2), log(2, y())),
		//poly(map[float64]float64{2: 3, 1: 2, 0: 1}, x()),
		polyParse("3x^2 + 12x + 144 + 4x^3 + 2x", x()),
	)

	fmt.Println(p)
	fmt.Println(p.Derivative())
	fmt.Println(p.Derivative().Derivative())
	fmt.Println(p.Derivative().Derivative().Derivative())
}

func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

//func minusMulSimplify(e *ExprsMultiplied) (result string, negative bool) {
//	c, expr, ok := getConstExpr(e.expr1, e.expr2)
//	if ok {
//		if c.num == 1 {
//			return expr.String(), false
//		}
//		if c.num == -1 {
//			return expr.String(), true
//		}
//	}
//	return e.String(), false
//}
//
//func getConstExpr(e1 Expression, e2 Expression) (c *Constant, expr Expression, ok bool) {
//	c1, okc1 := e1.(*Constant)
//	c2, okc2 := e2.(*Constant)
//	if okc1 {
//		return c1, e2, true
//	}
//	if okc2 {
//		return c2, e1, true
//	}
//	return nil, nil, false
//}

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
