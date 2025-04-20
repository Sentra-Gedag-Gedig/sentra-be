package entity

type KTPPosition struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

type KTPDetectionResult struct {
	Message     string       `json:"message"`
	BBox        []float64    `json:"bbox,omitempty"`
	Confidence  float64      `json:"conf,omitempty"`
	Center      []float64    `json:"center,omitempty"`
	BoxSize     []float64    `json:"box_size,omitempty"`
	Error       string       `json:"error,omitempty"`
	KTPPosition *KTPPosition `json:"ktp_position,omitempty"`
}
