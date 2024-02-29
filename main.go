package main

import (
	"fmt"
	"math/rand"
	"strconv"
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

	// ln(x + y)
	//p := add(x(), polyParse("x^2 + 2x + 1", x()))

	//p := randomExpressionGenerator()

	// ln(x) - x
	p := subtract(log(E, x()), x())
	//p := mul(x(), x(), polyParse("x^2 + 2x + 1", x()), polyParse("x^2 + 2x + 1", x()))

	fmt.Print("f(x)   = ")
	legible(p)
	//fmt.Println(p.structure())
	//fmt.Println(p.simplify().structure())
	fmt.Print("f'(x)  = ")
	legible(p.Derivative())
	fmt.Println(p.Derivative().structure())
	fmt.Print("f''(x) = ")
	legible(p.Derivative().Derivative())

	//d := p.Derivative().Derivative()
	//fmt.Println(d.structure())
	//fmt.Println(d.structure())
	//fmt.Println(d.structure())
	//fmt.Println(d.simplify().structure())

	//return

	//f, err := os.Create("derivatize.prof")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	//var e Expression
	//e = p
	//fmt.Println("f(x)")
	//legible(e)
	//for i := 0; i < 4; i++ {
	//	fmt.Println(i+1, "th derivative:")
	//	e = e.Derivative().simplify().simplify().simplify().simplify()
	//	legible(e)
	//	//fmt.Println(e.structure())
	//	//fmt.Println(e.simplify().structure())
	//}
	//fmt.Print("f(10)(x) = ")
	//legible(e)
}

func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

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
