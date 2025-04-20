package detectionService

import (
	"ProjectGolang/internal/api/detection"
	"ProjectGolang/internal/entity"
	"encoding/json"
	"errors"
	"golang.org/x/net/context"
	"strings"
)

func (s *detectionService) ProcessFrame(frame []byte) (*entity.DetectionResult, error) {
	result, err := s.websocketPkg.ProcessFaceFrame(frame)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *detectionService) ProcessKTPFrame(frame []byte) (*entity.KTPDetectionResult, error) {
	result, err := s.websocketPkg.ProcessKTPFrame(frame)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *detectionService) ProcessQRISFrame(frame []byte) (*entity.QRISDetectionResult, error) {
	result, err := s.websocketPkg.ProcessQRISFrame(frame)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *detectionService) ExtractAndSaveKTP(ctx context.Context, base64Image string) (*detection.KTP, error) {
	prompt := `
	Ekstrak semua informasi dari KTP Indonesia ini dan berikan hasilnya dalam format JSON.
	Format output yang diinginkan:
	{
		"nik": "1234567890123456",
		"nama": "NAMA LENGKAP",
		"tempat_lahir": "KOTA",
		"tanggal_lahir": "01-01-1990",
		"jenis_kelamin": "LAKI-LAKI/PEREMPUAN",
		"golongan_darah": "A/B/AB/O",
		"alamat": "ALAMAT LENGKAP",
		"rt": "001",
		"rw": "002",
		"kelurahan": "NAMA KELURAHAN",
		"kecamatan": "NAMA KECAMATAN",
		"agama": "ISLAM/KRISTEN/KATOLIK/dll",
		"status_perkawinan": "BELUM KAWIN/KAWIN/CERAI HIDUP/CERAI MATI",
		"pekerjaan": "JENIS PEKERJAAN",
		"kewarganegaraan": "WNI/WNA",
		"berlaku_hingga": "SEUMUR HIDUP/31-12-2025"
	}
	Berikan HANYA respons JSON, tanpa teks tambahan apapun.
	`

	result, err := s.gemini.AnalyzeImage(ctx, base64Image, prompt)
	if err != nil {
		return nil, err
	}

	ktp, err := parseGeminiResponse(result)

	return ktp, nil
}

func parseGeminiResponse(response string) (*detection.KTP, error) {
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, errors.New("cannot find valid JSON in response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	var ktp detection.KTP
	err := json.Unmarshal([]byte(jsonStr), &ktp)
	if err != nil {
		return nil, err
	}

	if ktp.NIK == "" || ktp.Nama == "" {
		return nil, errors.New("failed to extract essential KTP information")
	}

	return &ktp, nil
}

func (s *detectionService) DetectMoney(ctx context.Context, base64Image string) (*detection.MoneyDetectionResponse, error) {
	prompt := `
	Identifikasi uang Rupiah pada gambar ini dan berikan hasilnya dalam format JSON.
	
	Hal-hal yang perlu diidentifikasi:
	1. Nominal/nilai uang dalam angka (contoh: 50000, 100000)
	2. Jenis uang (kertas/logam)
	3. Gambar/tokoh yang ada di uang tersebut
	
	Format output yang diinginkan:
	{
		"nominal": 50000,
		"jenis": "kertas",
		"tokoh": "I Gusti Ngurah Rai",
		"warna_dominan": "Biru"
	}
	
	Jika terdapat lebih dari satu uang dalam gambar, identifikasi semua uang tersebut dan jumlahkan nilai totalnya. 
	Contoh output jika terdapat lebih dari satu uang:
	{
		"detail": [
			{
				"nominal": 50000,
				"jenis": "kertas",
				"tokoh": "I Gusti Ngurah Rai",
				"warna_dominan": "Biru"
			},
			{
				"nominal": 100000,
				"jenis": "kertas",
			}
		],
		"total": 150000
	}
	
	Berikan HANYA respons JSON, tanpa teks tambahan apapun.
	`

	result, err := s.gemini.AnalyzeImage(ctx, base64Image, prompt)
	if err != nil {
		return nil, err
	}

	return parseMoneyResponse(result)
}

func parseMoneyResponse(response string) (*detection.MoneyDetectionResponse, error) {
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, errors.New("cannot find valid JSON in response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	var multipleResponse detection.MoneyDetectionMultiple
	err := json.Unmarshal([]byte(jsonStr), &multipleResponse)
	if err == nil && len(multipleResponse.Detail) > 0 {
		return &detection.MoneyDetectionResponse{
			Multiple: true,
			Details:  multipleResponse.Detail,
			Total:    multipleResponse.Total,
		}, nil
	}

	var singleResponse detection.MoneyDetail
	err = json.Unmarshal([]byte(jsonStr), &singleResponse)
	if err != nil {
		return nil, errors.New("failed to parse Gemini response as valid JSON")
	}

	if singleResponse.Nominal == 0 {
		return nil, errors.New("failed to extract money value information")
	}

	return &detection.MoneyDetectionResponse{
		Multiple: false,
		Details:  []detection.MoneyDetail{singleResponse},
		Total:    singleResponse.Nominal,
	}, nil
}
