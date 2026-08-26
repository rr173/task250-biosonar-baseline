package model

// SubstrateType is a seafloor class described by a feature centroid and a
// diagonal covariance in feature space. Classification uses a Gaussian
// likelihood against these parameters.
type SubstrateType struct {
	ID       int64     `json:"id"`
	Code     string    `json:"code"`
	Name     string    `json:"name"`
	Centroid []float64 `json:"centroid"`
	CovDiag  []float64 `json:"cov_diag"`
}

// FeatureVector is the extracted descriptor of an echo window.
//
// Dimension order (must stay consistent everywhere):
//
//	[0] bs_mean       mean backscatter amplitude (dB)
//	[1] bs_std        backscatter spread (dB)
//	[2] penetration   mean penetration depth (m)
//	[3] angular_slope slope of backscatter vs incidence angle (dB/deg)
type FeatureVector []float64

// FeatureDim is the expected length of a FeatureVector.
const FeatureDim = 4

// Validate checks that the substrate model is well formed.
func (s *SubstrateType) Validate() error {
	if len(s.Centroid) != FeatureDim || len(s.CovDiag) != FeatureDim {
		return ErrUnknownSubstrate
	}
	for _, v := range s.CovDiag {
		if v <= 0 {
			return ErrUnknownSubstrate
		}
	}
	return nil
}
