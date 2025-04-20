package detection

type KTP struct {
	NIK              string `json:"nik"`
	Nama             string `json:"nama"`
	TempatLahir      string `json:"tempat_lahir"`
	TanggalLahir     string `json:"tanggal_lahir"`
	JenisKelamin     string `json:"jenis_kelamin"`
	GolonganDarah    string `json:"golongan_darah"`
	Alamat           string `json:"alamat"`
	RT               string `json:"rt"`
	RW               string `json:"rw"`
	Kelurahan        string `json:"kelurahan"`
	Kecamatan        string `json:"kecamatan"`
	Agama            string `json:"agama"`
	StatusPerkawinan string `json:"status_perkawinan"`
	Pekerjaan        string `json:"pekerjaan"`
	Kewarganegaraan  string `json:"kewarganegaraan"`
	BerlakuHingga    string `json:"berlaku_hingga"`
}

type OCRRequest struct {
	ImageBase64 string `json:"image_base64"`
}

type OCRResponse struct {
	Data  KTP    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type GeminiConfig struct {
	APIKey    string
	ModelName string
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text  string       `json:"text,omitempty"`
	Image *GeminiImage `json:"inlineData,omitempty"`
}

type GeminiImage struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

type MoneyDetectionRequest struct {
	ImageBase64 string `json:"image_base64" validate:"required"`
}

type MoneyDetail struct {
	Nominal int    `json:"nominal"`
	Jenis   string `json:"jenis"`
}

type MoneyDetectionMultiple struct {
	Detail []MoneyDetail `json:"detail"`
	Total  int           `json:"total"`
}

type MoneyDetectionResponse struct {
	Multiple bool          `json:"multiple"`
	Details  []MoneyDetail `json:"details"`
	Total    int           `json:"total"`
}

type MoneyResponse struct {
	Data  MoneyDetectionResponse `json:"data,omitempty"`
	Error string                 `json:"error,omitempty"`
}

type KTPDetectionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
type DetectionType string

const (
	KTPDetection  DetectionType = "KTP"
	FaceDetection DetectionType = "FACE"
	QRISDetection DetectionType = "QRIS"
)
