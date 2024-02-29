package main

import (
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

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
			if coeffStr == "-" {
				coeffStr = "-1"
			}
			str += coeffStr + " + "
		} else if power == 1 {
			str += coeffStr + insideStr + " + "
		} else {
			str += coeffStr + insideStr + "^" + ftoa(power) + " + "
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
