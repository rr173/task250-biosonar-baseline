package echo

import (
	"math"

	"task250-biosonar/internal/model"
)

// Extract turns a corrected echo window's multispectral channels into the
// 4-dimensional FeatureVector used for substrate classification.
//
//	[0] bs_mean       mean backscatter amplitude across channels (dB)
//	[1] bs_std        backscatter spread across channels (dB)
//	[2] penetration   mean penetration depth across channels (m)
//	[3] angular_slope slope of backscatter vs incidence angle (dB/deg)
func Extract(e *model.EchoWindow) (model.FeatureVector, error) {
	if len(e.Channels) == 0 {
		return nil, model.ErrEmptyChannels
	}
	var bsVals, penVals, incAngles, bsAtAngle []float64
	for _, c := range e.Channels {
		if len(c.Amplitudes) == 0 || len(c.Depths) != len(c.Amplitudes) {
			return nil, model.ErrEmptyChannels
		}
		for i, d := range c.Depths {
			if math.IsNaN(d) || math.IsInf(d, 0) || math.IsNaN(c.Amplitudes[i]) || math.IsInf(c.Amplitudes[i], 0) ||
				(i > 0 && d <= c.Depths[i-1]) {
				return nil, model.ErrEmptyChannels
			}
		}
		m := mean(c.Amplitudes)
		p := penetration(c.Amplitudes, c.Depths)
		bsVals = append(bsVals, m)
		penVals = append(penVals, p)
		incAngles = append(incAngles, c.IncidenceDeg)
		bsAtAngle = append(bsAtAngle, m)
	}
	if len(bsVals) == 0 {
		return nil, model.ErrEmptyChannels
	}
	bsMean := mean(bsVals)
	bsStd := std(bsVals, bsMean)
	pen := mean(penVals)
	slope := linearSlope(incAngles, bsAtAngle)
	return model.FeatureVector{bsMean, bsStd, pen, slope}, nil
}
