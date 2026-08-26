package geometry

import "math"

// dot returns the inner product of two equal-length vectors.
func dot(a, b []float64) float64 {
	var s float64
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}

// sub returns a - b element-wise.
func sub(a, b []float64) []float64 {
	out := make([]float64, len(a))
	for i := range a {
		out[i] = a[i] - b[i]
	}
	return out
}

// norm2 returns the squared Euclidean norm.
func norm2(a []float64) float64 { return dot(a, a) }

// maskedMean computes the mean of values whose corresponding mask is true.
// It returns 0 when no value is selected.
func maskedMean(values []float64, mask []bool) float64 {
	var sum float64
	var n int
	for i, v := range values {
		if i < len(mask) && mask[i] {
			sum += v
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

// safeDiv divides a by b, returning fallback when b is near zero.
func safeDiv(a, b, fallback float64) float64 {
	if math.Abs(b) < 1e-9 {
		return fallback
	}
	return a / b
}
