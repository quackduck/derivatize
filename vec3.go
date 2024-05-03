package main

type Vec3 struct {
	f, g, h Expression
}

func (v *Vec3) fDerivative() Expression {
	xwrt(true)
	a := v.f.Derivative()
	xwrt(false)
	return a
}

func (v *Vec3) gDerivative() Expression {
	ywrt(true)
	a := v.g.Derivative()
	ywrt(false)
	return a
}

func (v *Vec3) hDerivative() Expression {
	zwrt(true)
	a := v.h.Derivative()
	zwrt(false)
	return a
}

func (v *Vec3) fDerivativeWrt(h *Var) Expression {
	h.wrt = true
	a := v.f.Derivative()
	h.wrt = false
	return a
}

func (v *Vec3) gDerivativeWrt(h *Var) Expression {
	h.wrt = true
	a := v.g.Derivative()
	h.wrt = false
	return a
}

func (v *Vec3) hDerivativeWrt(h *Var) Expression {
	h.wrt = true
	a := v.h.Derivative()
	h.wrt = false
	return a
}

func (v *Vec3) Gradient() *Vec3 {
	return &Vec3{v.fDerivative(), v.gDerivative(), v.hDerivative()}
}

func (v *Vec3) Simplify() {
	v.f = v.f.simplify()
	v.g = v.g.simplify()
	v.h = v.h.simplify()
}

func (v *Vec3) String() string {
	return "<" + legibleStr(v.f) + ", " + legibleStr(v.g) + ", " + legibleStr(v.h) + ">"
}

func (v *Vec3) structure() string {
	return "vec3{" + v.f.structure() + ", " + v.g.structure() + ", " + v.h.structure() + "}"
}

func vec3(f, g, h Expression) *Vec3 {
	return &Vec3{f, g, h}
}

func (v *Vec3) Curl() *Vec3 {
	return &Vec3{
		subtract(v.hDerivativeWrt(y), v.gDerivativeWrt(z)).simplify(),
		subtract(v.fDerivativeWrt(z), v.hDerivativeWrt(x)).simplify(),
		subtract(v.gDerivativeWrt(x), v.fDerivativeWrt(y)).simplify(),
	}
}

func (v *Vec3) Div() Expression {
	return add(v.fDerivative(), v.gDerivative(), v.hDerivative()).simplify()
}

func (v *Vec3) Dot(v2 *Vec3) Expression {
	return add(mul(v.f, v2.f), mul(v.g, v2.g), mul(v.h, v2.h)).simplify()
}

func (v *Vec3) Cross(v2 *Vec3) *Vec3 {
	return &Vec3{
		subtract(mul(v.g, v2.h), mul(v.h, v2.g)).simplify(),
		subtract(mul(v.h, v2.f), mul(v.f, v2.h)).simplify(),
		subtract(mul(v.f, v2.g), mul(v.g, v2.f)).simplify(),
	}
}

func xwrt(on bool) {
	x.wrt = on
}

func ywrt(on bool) {
	y.wrt = on
}

func zwrt(on bool) {
	z.wrt = on
}
