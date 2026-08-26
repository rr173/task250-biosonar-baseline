package classify

import "task250-biosonar/internal/model"

// DefaultSubstrates returns the built-in seafloor classes used when no
// operator-supplied catalogue is present. Each centroid follows the feature
// ordering [bs_mean, bs_std, penetration, angular_slope].
func DefaultSubstrates() []model.SubstrateType {
	return []model.SubstrateType{
		{Code: "SAND", Name: "Sand",
			Centroid: []float64{-12.0, 2.0, 2.5, 0.05},
			CovDiag:  []float64{1.0, 0.5, 0.4, 0.005}},
		{Code: "MUD", Name: "Mud",
			Centroid: []float64{-22.0, 1.5, 8.0, 0.02},
			CovDiag:  []float64{1.5, 0.4, 0.8, 0.004}},
		{Code: "ROCK", Name: "Rock",
			Centroid: []float64{-10.0, 4.0, 1.0, 0.25},
			CovDiag:  []float64{1.2, 0.8, 0.3, 0.020}},
		{Code: "COBBLE", Name: "Cobble",
			Centroid: []float64{-15.0, 5.0, 3.0, 0.15},
			CovDiag:  []float64{1.3, 1.0, 0.5, 0.010}},
	}
}
