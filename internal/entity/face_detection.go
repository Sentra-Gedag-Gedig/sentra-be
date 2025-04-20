package entity

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type DetectionResult struct {
	Status       string             `json:"status"`
	Instructions []string           `json:"instructions"`
	FacePosition *Position          `json:"face_position,omitempty"`
	FaceSize     *float64           `json:"face_size,omitempty"`
	FrameCenter  Position           `json:"frame_center"`
	Deviations   map[string]float64 `json:"deviations,omitempty"`
}
