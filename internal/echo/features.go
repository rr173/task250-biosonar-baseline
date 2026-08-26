package echo

import "math"

// mean returns the arithmetic mean of xs (0 for empty input).
func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var s float64
	for _, v := range xs {
		s += v
	}
	return s / float64(len(xs))
}

// std returns the population standard deviation of xs about m.
func std(xs []float64, m float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var s float64
	for _, v := range xs {
		d := v - m
		s += d * d
	}
	return math.Sqrt(s / float64(len(xs)))
}

// penetration estimates how far the acoustic signal penetrates the substrate
// before dropping 3 dB below its peak return.
func penetration(amps, depths []float64) float64 {
	if len(amps) == 0 || len(depths) == 0 {
		return 0
	}
	peak := amps[0]
	for _, a := range amps {
		if a > peak {
			peak = a
		}
	}
	threshold := peak - 3.0
	last := depths[len(depths)-1]
	for i, a := range amps {
		if a < threshold {
			if i == 0 {
				return depths[0]
			}
			return depths[i-1]
		}
		last = depths[i]
	}
	return last
}

// linearSlope returns the least-squares slope of ys against xs. When the
// independent variable has no spread it returns 0.
func linearSlope(xs, ys []float64) float64 {
	n := len(xs)
	if n < 2 {
		return 0
	}
	mx := mean(xs)
	my := mean(ys)
	var num, den float64
	for i := range xs {
		dx := xs[i] - mx
		num += dx * (ys[i] - my)
		den += dx * dx
	}
	if math.Abs(den) < 1e-9 {
		return 0
	}
	return num / den
}
