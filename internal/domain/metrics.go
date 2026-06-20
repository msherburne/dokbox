package domain

type MetricSample struct {
	Name  string
	Value float64
	Limit *float64
	Unit  string
	Label string
}

func (m MetricSample) Ratio() *float64 {
	if m.Limit == nil || *m.Limit == 0 {
		return nil
	}

	ratio := m.Value / *m.Limit
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	return &ratio
}
