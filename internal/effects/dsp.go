package effects

import (
	"math"
	"math/cmplx"
)

func computeFFT(a []complex128) {
	n := len(a)
	if n <= 1 {
		return
	}

	a0 := make([]complex128, n/2)
	a1 := make([]complex128, n/2)
	for i := 0; i < n/2; i++ {
		a0[i] = a[2*i]
		a1[i] = a[2*i+1]
	}

	computeFFT(a0)
	computeFFT(a1)

	for i := 0; i < n/2; i++ {
		w := cmplx.Exp(complex(0, -2*math.Pi*float64(i)/float64(n)))
		a[i] = a0[i] + w*a1[i]
		a[i+n/2] = a0[i] - w*a1[i]
	}
}

func getBands(samples []int16) (bass, mid, high float64) {
	n := len(samples)
	data := make([]complex128, n)
	
	for i := 0; i < n; i++ {
		window := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1)))
		data[i] = complex(float64(samples[i])*window, 0)
	}

	computeFFT(data)

	// Используем Max вместо среднего, чтобы не терять яркость
	for i := 1; i < n/2; i++ {
		magnitude := cmplx.Abs(data[i])
		
		if i <= 4 { // ~0-370 Гц (Бас)
			if magnitude > bass { bass = magnitude }
		} else if i <= 40 { // ~370-3700 Гц (Средние/Вокал)
			if magnitude > mid { mid = magnitude }
		} else { // 3700+ Гц (Высокие)
			if magnitude > high { high = magnitude }
		}
	}

	return bass, mid, high
}
