package main

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

//var debug = true

var debug = false

type Expression interface {
	Derivative() Expression
	String() string
	simplify() Expression
	structure() string
}

func main() {
	//p := div(num(1), log(E, mul(x(), add(x(), x()))))

	// x ln x - ln y
	//p := subtract(mul(x(), log(E, x())), log(E, add(y(), y().Derivative())))

	// 1 / sqrt(x^2 + 1)
	//p := div(num(1), polyParse("x^0.5", polyParse("x^2 + 1", x())))
	// x / ln(x)
	//p := div(x(), log(E, x()))

	// (x^2 - 4) / (x + 2)
	//p := div(polyParse("x^2 + -4", x()), add(x(), num(2)))

	//fmt.Println(p.structure())
	//fmt.Println(p.Derivative().structure())
	//fmt.Println(p.structure())

	//p := polyParse("1", x())

	//p := num(1)

	//p := randomExpressionGenerator()

	// e ^ (ln(x^2))
	//p := exp(E, log(E, polyParse("x^2", x())))

	//eggholder function

	//s1 := polyParse("x^0.5", abs(
	//	add(c(), div(b(), num(2)), num(47)),
	//))
	//s2 := polyParse("x^0.5", abs(
	//	subtract(b(), add(c(), num(47))),
	//))
	//
	//p := subtract(
	//	mul(num(-1), add(c(), num(47), sin(s1))),
	//	mul(b(), sin(s2)),
	//)

	//p := sin(cos(x()))

	//p := sin(cos(x()))

	s1 := polyParse("x^0.5",
		polyParse("0.5x + 47", x()),
	)
	s2 := polyParse("x^0.5",
		polyParse("x + -47", x()),
	)

	p := subtract(
		mul(num(-47), sin(s1)),
		mul(x(), sin(s2)),
	)

	// sqrt(25 - b^2 - c^2)
	//p := polyParse("x^0.5",
	//	subtract(
	//		num(25),
	//		add(
	//			polyParse("x^2", b()),
	//			polyParse("x^2", c()),
	//		),
	//	),
	//)

	//p := mul(num(3), abs(subtract(x(), num(2))))

	//p := mul(num(-47), sin(s1))

	// (sin(x)/cos(x))^2 + e^((1/sin(x))) * log_8(1/cos((3x)^(1/3)))

	//p := add(
	//	polyParse("x^2", div(sin(x()), cos(x()))),
	//	mul(
	//		exp(E, div(num(1), sin(x()))),
	//		log(8, div(num(1), cos(polyParse("x^0.3333333333333333", mul(num(3), x()))))),
	//	),
	//)

	//p := div(x(), x())

	// (x + y) * e^(2xy)
	//p := mul(add(x(), y()), exp(E, mul(num(2), x(), y())))

	// 1 / ln(x+y)
	//p := div(num(1), log(E, add(x(), y())))

	//p := exp(E, mul(num(-1), x()))

	//fmt.Print("f(x)   = ")
	//legible(p)
	//fmt.Print("f'(x)  = ")
	//legible(p.Derivative())
	//fmt.Print("f''(x) = ")
	//legible(p.Derivative().Derivative())

	//return

	//f, err := os.Create("derivatize.prof")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	var e Expression
	e = p
	fmt.Println("f(x)")
	legible(e)
	for i := 0; i < 4; i++ {
		fmt.Println(i+1, "th derivative:")
		e = e.Derivative().simplify().simplify()
		legible(e)
		//fmt.Println(e.structure())
		//fmt.Println(e.simplify().structure())
	}
	//fmt.Print("f(10)(x) = ")
	//legible(e)
}

type Constant struct{ num float64 }

func (n *Constant) Derivative() Expression { return &Constant{0} }
func (n *Constant) String() string         { return ftoa(n.num) }
func (n *Constant) simplify() Expression   { return n }
func num(num float64) *Constant            { return &Constant{num} }
func (n *Constant) structure() string      { return "const{" + n.String() + "}" }

type X struct{}

func (x *X) Derivative() Expression { return &Constant{1} }
func (x *X) String() string         { return "x" }
func (x *X) simplify() Expression   { return x }
func x() Expression                 { return &X{} }
func (x *X) structure() string      { return "x{}" }

type Y struct{ derivnum int }

func (y *Y) Derivative() Expression { return &Y{y.derivnum + 1} }
func (y *Y) String() string         { return "y" + strings.Repeat("'", y.derivnum) }
func (y *Y) simplify() Expression   { return y }
func y() Expression                 { return &Y{} }
func (y *Y) structure() string      { return "y{}" }

type B struct{}

func (b *B) Derivative() Expression { return &Constant{1} }
func (b *B) String() string         { return "b" }
func (b *B) simplify() Expression   { return b }
func b() Expression                 { return &B{} }
func (b *B) structure() string      { return "b{}" }

type C struct{}

func (c *C) Derivative() Expression { return &Constant{0} }
func (c *C) String() string         { return "c" }
func (c *C) simplify() Expression   { return c }
func c() Expression                 { return &C{} }
func (c *C) structure() string      { return "c{}" }

type Abs struct {
	inside Expression
}

func (a *Abs) Derivative() Expression {
	return mul(div(a.inside, abs(a.inside)), a.inside.Derivative())
}

func (a *Abs) String() string {
	return "|" + a.inside.String() + "|"
}

func (a *Abs) simplify() Expression {
	a.inside = a.inside.simplify()
	return a
}

func abs(e Expression) Expression {
	return &Abs{e}
}

func (a *Abs) structure() string {
	return "abs{" + a.inside.structure() + "}"
}

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
	e.es = mergeBasedOnReflect(e.es, true)
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
	noConsts = mulPolysAndVars(noConsts, &Constant{constant})
	//if constant != 1.0 {
	//	noConsts = append([]Expression{&Constant{constant}}, noConsts...)
	//}
	e.es = noConsts
	return e
}

func mulPolysAndVars(es []Expression, c *Constant) (rest []Expression) {
	es = mulMergeDivides(es)

	// split into polynomials, x, y and the rest
	var polys []*Polynomial
	polys, rest = splitByType[*Polynomial](es)

	if len(polys) != 0 && c.num != 1.0 {
		for power, coeff := range polys[0].powerToCoeff {
			polys[0].powerToCoeff[power] = coeff * c.num
		}
	} else if len(polys) == 0 && c.num != 1.0 {
		defer func() {
			rest = append([]Expression{c}, rest...)
		}()
	}

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

	if len(xs) == 0 && len(ys) == 0 {
		for _, p := range polys {
			rest = append(rest, p)
		}
		return rest
	}

	xdone := false
	ydone := false

	if len(xs) == 0 {
		xdone = true
	}
	if len(ys) == 0 {
		ydone = true
	}

	for _, p := range polys {
		_, isX := p.inside.(*X)
		yin, isY := p.inside.(*Y)
		if isY && yin.derivnum != 0 {
			isY = false // ignore y' y'' etc for now at least
		}
		if !isX && !isY {
			//rest = append(rest, p)
			continue
		}
		newPowerToCoeff := make(map[float64]float64, len(p.powerToCoeff))
		for power := range p.powerToCoeff {
			if isX {
				//p.powerToCoeff[power] += float64(len(xs))
				newPowerToCoeff[power+float64(len(xs))] = p.powerToCoeff[power]
				//fmt.Println("added to x", power, coeff)
			} else if isY {
				//p.powerToCoeff[power] += float64(len(xs))
				newPowerToCoeff[power+float64(len(ys))] = p.powerToCoeff[power]
				//fmt.Println("added to y", power, coeff)
			}
		}
		p.powerToCoeff = newPowerToCoeff
		xdone = xdone || isX
		ydone = ydone || isY
		if xdone && ydone {
			break
		}
	}
	//rest = append(rest, polys...)
	if !xdone {
		if len(xs) == 1 {
			rest = append(rest, x())
		} else {
			rest = append(rest, poly(map[float64]float64{float64(len(xs)): 1}, x()))
		}
	}
	if !ydone {
		if len(ys) == 1 {
			rest = append(rest, y())
		} else {
			rest = append(rest, poly(map[float64]float64{float64(len(ys)): 1}, y()))
		}
	}
	for i := range polys {
		rest = append(rest, polys[i])
	}
	return rest
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
	for _, expr := range e.es {
		result += expr.structure() + " * "
	}
	return "mul{" + result[:len(result)-3] + "}"
}

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

type ExprsDivided struct {
	high Expression
	low  Expression
}

func (e *ExprsDivided) simplify() (ret Expression) {
	e.high = e.high.simplify()
	e.low = e.low.simplify()

	//fmt.Println("simplify", e.high, e.low)

	if mul1, ok := e.high.(*ExprsMultiplied); ok {
		// (a * b) / c = (a * b * 1/c)
		//fmt.Println("HELLO", mul1.es, e.low)
		mul1.es = append(mul1.es, div(num(1), e.low).simplify())
		//fmt.Println("HELLO", mul1.es, e.low)
		e.low = num(1)
		return mul1.simplify()
	}

	if p1, ok := e.high.(*Polynomial); ok {
		if _, ok = e.low.(*X); ok {
			newPowerToCoeff := make(map[float64]float64, len(p1.powerToCoeff))
			for power, coeff := range p1.powerToCoeff {
				newPowerToCoeff[power-1] = coeff
			}
			p1.powerToCoeff = newPowerToCoeff
			e.low = num(1)
			return p1.simplify()
		}
	}

	c1, okc1 := e.high.(*Constant)
	c2, okc2 := e.low.(*Constant)
	if okc1 && okc2 {
		return &Constant{c1.num / c2.num}
	}
	if okc2 {
		if c2.num == 1 {
			return e.high
		}
		if c2.num == 0 {
			panic("division by zero")
		}
		if c2.num == -1 {
			return mul(num(-1), e.high)
		}
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
		if x, ok := e.low.(*X); ok {
			return (poly(map[float64]float64{-1: c1.num}, x)).simplify()
		}
	}

	// check if the high is ExprsDivided
	if d, ok := e.high.(*ExprsDivided); ok {
		// (a/b) / c = a / (b*c)
		return div(d.high, mul(d.low, e.low)).simplify()
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
			expr1: mul(e.low, e.high.Derivative()),
			expr2: mul(e.high, e.low.Derivative()),
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

func div(e1 Expression, e2 Expression) Expression {
	//fmt.Println("div", e1, e2)
	return &ExprsDivided{e1, e2}
}

func (e *ExprsDivided) structure() string {
	return "div{" + e.high.structure() + " / " + e.low.structure() + "}"
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
				return num(coeff)
			}
			if power == 1 {
				return mul(num(coeff), p.inside)
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
	derivative := poly(make(map[float64]float64, len(p.powerToCoeff)), p.inside)
	for power, coeff := range p.powerToCoeff {
		if power != 0 {
			derivative.powerToCoeff[power-1] = power * coeff
		}
	}
	//fmt.Println("multiplying by", p.inside.Derivative())
	return mul(derivative, p.inside.Derivative())
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
		if coeff == -1 {
			coeffStr = "-"
		} else if coeff != 1 {
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

func (p *Polynomial) Equal(p2 *Polynomial) bool {

	return reflect.DeepEqual(p, p2) // hopefully this works

	//switch e := p.inside.(type) {
	//case *X:
	//case *Y:
	//	switch e2 := p2.inside.(type) {
	//
	//	}
	//}
	//
	//for power, coeff := range p.powerToCoeff {
	//	if p2.powerToCoeff[power] != coeff {
	//		return false
	//	}
	//}
	//return true
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

func (p *Polynomial) structure() string {
	result := ""
	for power, coeff := range p.powerToCoeff {
		result += ftoa(power) + ":" + ftoa(coeff) + ", "
	}
	return "poly{" + result[:len(result)-2] + " " + p.inside.structure() + "}"
}

type Log struct {
	base float64
	expr Expression
}

func (l *Log) simplify() Expression {
	l.expr = l.expr.simplify()

	if e, ok := l.expr.(*Exponential); ok {
		if e.base == l.base {
			return e.power
		}
	}
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
		mul(
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
	switch l.expr.(type) {
	case *ExprsAdded, *ExprsSubtracted:

	case *Polynomial:
		if insideStr[0] != '(' {
			insideStr = "(" + insideStr + ")"
		}
	case *Constant:
		insideStr = "(" + insideStr + ")"
	default:
		insideStr = "[ " + insideStr + " ]"
	}
	if l.base == E {
		return "ln" + insideStr
	}
	return "log_" + ftoa(l.base) + insideStr
}

func log(base float64, expr Expression) *Log { return &Log{base, expr} }

func (l *Log) structure() string {
	return "log_" + ftoa(l.base) + "{ " + l.expr.structure() + "}"
}

type Exponential struct {
	base  float64
	power Expression
}

func (e *Exponential) simplify() Expression {
	e.power = e.power.simplify()
	if c, ok := e.power.(*Constant); ok {
		if e.base != E {
			return &Constant{math.Pow(e.base, c.num)}
		}
	}
	if l, ok := e.power.(*Log); ok {
		if l.base == e.base {
			return l.expr
		}
	}
	return e
}

func (e *Exponential) Derivative() Expression {
	simp := e.simplify()
	if simp != e {
		return simp.Derivative()
	}
	if e.base == E {
		return mul(
			exp(E, e.power),
			e.power.Derivative(),
		)
	}
	return mul(
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

//
//type Reciprocal struct {
//	expr Expression
//}
//
//func (r *Reciprocal) simplify() Expression {
//	r.expr = r.expr.simplify()
//	//if c, ok := r.expr.(*Constant); ok {
//	//	if c.num == 0 {
//	//		panic("division by zero")
//	//	}
//	//	return &Constant{1 / c.num}
//	//}
//	switch e := r.expr.(type) {
//	case *Reciprocal:
//		return e.expr
//	case *Polynomial:
//		return poly(map[float64]float64{-1: 1}, e.inside).simplify()
//	}
//	return r
//}
//
//func (r *Reciprocal) Derivative() Expression {
//	simp := r.simplify()
//	if simp != r {
//		return simp.Derivative()
//	}
//	return poly(map[float64]float64{-2: 1}, r.expr).simplify()
//}
//
//func (r *Reciprocal) String() string {
//	simp := r.simplify()
//	if simp != r {
//		return simp.String()
//	}
//	insideStr := r.expr.String()
//	if _, ok := r.expr.(*Polynomial); !ok {
//		insideStr = "[ " + insideStr + " ]"
//	}
//	return "1 / " + insideStr
//}

func exp(base float64, power Expression) *Exponential { return &Exponential{base, power} }

func (e *Exponential) structure() string {
	return "exp_" + ftoa(e.base) + "{" + e.power.structure() + "}"
}

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

func legible(e Expression) {
	str := e.String()

	colors := []string{
		//"\033[31m",
		"\033[32m",
		"\033[33m",
		"\033[34m",
		//"\033[48,5,208m",
		//"\x1b[38;5;214m",
		"\033[35m",
		"\033[36m",
		"\033[37m",
		//"\033[91m",
		//"\033[92m",
		//"\033[93m",
	}
	//slices.Reverse(colors)
	result := "\033[0m" + colors[0]
	depth := 0
	for _, c := range str {
		// get depth
		if c == '(' || c == '[' {
			depth++
		}
		if c == ')' || c == ']' {
			depth--
		}
		if c == ')' || c == ']' {
			result += colors[depth%len(colors)]
		}
		result += string(c)
		if c == '(' || c == '[' {
			result += colors[depth%len(colors)]
		}
	}
	result += "\033[0m"
	fmt.Println(result)
}

func randomExpressionGenerator() Expression {
	funcs2input := []func(Expression, Expression) Expression{
		subtract,
		div,
	}
	vars := []func() Expression{
		x,
		//y,
	}
	funcsNinput := []func(...Expression) Expression{
		add,
		mul,
	}
	funcsCustom := []func() Expression{
		func() Expression { return log(E, randomExpressionGenerator()) },
		func() Expression { return exp(E, randomExpressionGenerator()) },
		func() Expression { return sin(randomExpressionGenerator()) },
		func() Expression { return cos(randomExpressionGenerator()) },
		func() Expression {
			// random polynomial
			powerToCoeff := make(map[float64]float64)
			numTerms := rand.Intn(3) + 1
			for i := 0; i < numTerms; i++ {
				powerToCoeff[float64(rand.Intn(14)-7)] = float64(rand.Intn(14) - 7)
			}
			return poly(powerToCoeff, x())
		},
	}

	// generate a random expression
	// 3 lists so pick a random one
	r := rand.Float64()
	weights := []float64{0.25, 0.25, 0.15, 0.35}
	if r < weights[0] {
		return funcs2input[rand.Intn(len(funcs2input))](randomExpressionGenerator(), randomExpressionGenerator())
	}
	if r < weights[0]+weights[1] {
		return vars[rand.Intn(len(vars))]()
	}
	if r < weights[0]+weights[1]+weights[2] {
		// pick a random number of inputs
		numInputs := rand.Intn(1) + 2
		inputs := make([]Expression, numInputs)
		for i := range inputs {
			inputs[i] = randomExpressionGenerator()
		}
		return funcsNinput[rand.Intn(len(funcsNinput))](inputs...)
	}
	return funcsCustom[rand.Intn(len(funcsCustom))]()
}
