
package nlp

import (
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type NLPProcessor struct {
	pageMappings map[string]PageMappingData
	stopWords    map[string]bool
}

func NewProcessor(config interface{}) INLPProcessor {
	
	stopWords := map[string]bool{
		"saya": true, "ke": true, "di": true, "dan": true, "atau": true,
		"untuk": true, "dari": true, "dengan": true, "pada": true, "dalam": true,
		"yang": true, "adalah": true, "akan": true, "telah": true, "sudah": true,
		"bisa": true, "dapat": true, "mau": true, "ingin": true, "pergi": true,
		"buka": true, "lihat": true, "tampilkan": true, "tunjukkan": true,
		"the": true, "to": true, "go": true, "open": true, "show": true,
		"view": true, "see": true, "display": true, "navigate": true,
		"page": true, "halaman": true, "menu": true,
	}

	
	defaultMappings := getDefaultPageMappings()

	return &NLPProcessor{
		pageMappings: defaultMappings,
		stopWords:    stopWords,
	}
}

func (nlp *NLPProcessor) ProcessCommand(text string) (*IntentResult, error) {
	startTime := time.Now()
	
	
	cleanText := nlp.cleanText(text)
	
	
	tokens := nlp.extractTokens(cleanText)
	
	
	matches := nlp.findBestMatches(tokens, cleanText)
	
	processingTime := time.Since(startTime)
	
	if len(matches) == 0 {
		return &IntentResult{
			Intent:         "unknown",
			Confidence:     0.0,
			ProcessingTime: processingTime.String(),
		}, nil
	}

	
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Confidence > matches[j].Confidence
	})

	bestMatch := matches[0]
	bestMatch.ProcessingTime = processingTime.String()

	return bestMatch, nil
}

func (nlp *NLPProcessor) findBestMatches(tokens []string, fullText string) []*IntentResult {
	var results []*IntentResult

	for pageID, mapping := range nlp.pageMappings {
		confidence := nlp.calculatePageConfidence(tokens, fullText, mapping)
		
		if confidence.Confidence > 0.2 { 
			result := &IntentResult{
				Intent:          "navigate",
				Page:            pageID,
				PageURL:         mapping.URL,
				PageDisplayName: mapping.DisplayName,
				PageDescription: mapping.Description,
				Confidence:      confidence.Confidence,
				Matches:         confidence.Matches,
			}
			results = append(results, result)
		}
	}

	return results
}

func (nlp *NLPProcessor) calculatePageConfidence(tokens []string, fullText string, mapping PageMappingData) *confidenceResult {
	var matches []MatchResult
	totalScore := 0.0
	maxPossibleScore := 0.0

	
	for _, keyword := range mapping.Keywords {
		for _, token := range tokens {
			if strings.EqualFold(token, keyword) {
				matches = append(matches, MatchResult{
					Keyword: keyword,
					Score:   1.0,
					Type:    "exact",
				})
				totalScore += 1.0
			}
		}
		maxPossibleScore += 1.0
	}

	
	for _, synonym := range mapping.Synonyms {
		similarity := nlp.calculateSimilarity(fullText, synonym)
		if similarity > 0.6 {
			matches = append(matches, MatchResult{
				Keyword: synonym,
				Score:   similarity,
				Type:    "synonym",
			})
			totalScore += similarity * 1.2 
		}
	}

	
	for _, keyword := range mapping.Keywords {
		for _, token := range tokens {
			similarity := nlp.calculateSimilarity(token, keyword)
			if similarity > 0.5 && similarity < 1.0 { 
				matches = append(matches, MatchResult{
					Keyword: keyword,
					Score:   similarity * 0.7, 
					Type:    "fuzzy",
				})
				totalScore += similarity * 0.7
			}
		}
	}

	
	categoryBonus := nlp.getCategoryBonus(tokens, mapping.Category)
	totalScore += categoryBonus

	
	confidence := totalScore / math.Max(maxPossibleScore, 1.0)
	if len(matches) > 1 {
		confidence *= 1.1 
	}
	confidence = math.Min(confidence, 1.0)

	return &confidenceResult{
		Confidence: confidence,
		Matches:    matches,
	}
}

func (nlp *NLPProcessor) calculateSimilarity(text1, text2 string) float64 {
	
	norm1 := nlp.cleanText(text1)
	norm2 := nlp.cleanText(text2)

	
	if norm1 == norm2 {
		return 1.0
	}

	
	if strings.Contains(norm1, norm2) || strings.Contains(norm2, norm1) {
		shorter := norm1
		longer := norm2
		if len(norm1) > len(norm2) {
			shorter = norm2
			longer = norm1
		}
		return float64(len(shorter)) / float64(len(longer))
	}

	
	distance := nlp.levenshteinDistance(norm1, norm2)
	maxLen := math.Max(float64(len(norm1)), float64(len(norm2)))
	
	if maxLen == 0 {
		return 0.0
	}
	
	return math.Max(0, 1.0 - (float64(distance) / maxLen))
}

func (nlp *NLPProcessor) levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      
				matrix[i][j-1]+1,      
				matrix[i-1][j-1]+cost, 
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	} else if b < c {
		return b
	}
	return c
}

func (nlp *NLPProcessor) getCategoryBonus(tokens []string, category string) float64 {
	categoryKeywords := map[string][]string{
		"finance": {"uang", "bayar", "payment", "money", "cash", "dana", "kredit", "debit", "saldo", "keuangan"},
		"user":    {"saya", "my", "pribadi", "personal", "data", "akun", "profil"},
		"system":  {"atur", "setting", "config", "ubah", "ganti", "pengaturan", "konfigurasi"},
		"support": {"tolong", "help", "bingung", "tidak tahu", "gimana", "how", "bantuan"},
		"communication": {"pesan", "message", "notif", "pemberitahuan", "kabar", "info"},
		"navigation": {"ke", "pergi", "buka", "lihat", "tampil", "navigasi"},
	}

	keywords, exists := categoryKeywords[category]
	if !exists {
		return 0.0
	}

	bonus := 0.0
	for _, token := range tokens {
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(token), strings.ToLower(keyword)) {
				bonus += 0.1
			}
		}
	}

	return math.Min(bonus, 0.3) 
}

func (nlp *NLPProcessor) cleanText(text string) string {
	
	text = strings.ToLower(text)
	
	
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, text)
	
	
	result = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			return r
		}
		return ' '
	}, result)
	
	
	words := strings.Fields(result)
	return strings.Join(words, " ")
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) 
}

func (nlp *NLPProcessor) extractTokens(text string) []string {
	words := strings.Fields(text)
	var tokens []string
	
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) > 1 && !nlp.stopWords[word] {
			tokens = append(tokens, word)
		}
	}
	
	return tokens
}

func (nlp *NLPProcessor) GetPageMapping(pageID string) (PageMappingData, bool) {
	mapping, exists := nlp.pageMappings[pageID]
	return mapping, exists
}

func (nlp *NLPProcessor) GetAllMappings() map[string]PageMappingData {
	return nlp.pageMappings
}

func (nlp *NLPProcessor) AddPageMapping(pageID string, mapping PageMappingData) error {
	nlp.pageMappings[pageID] = mapping
	return nil
}

func (nlp *NLPProcessor) GenerateResponseText(result *IntentResult) string {
	if result.Confidence < 0.3 {
		return "Maaf, saya tidak yakin dengan perintah yang Anda maksud. Bisa dijelaskan lebih spesifik?"
	}
	
	if result.Confidence > 0.8 {
		return "Menuju ke " + result.PageDisplayName
	} else if result.Confidence > 0.5 {
		return "Saya rasa Anda ingin ke " + result.PageDisplayName + ". Betul?"
	} else {
		return "Apakah Anda maksud " + result.PageDisplayName + "? Saya kurang yakin dengan perintah tersebut."
	}
}


type confidenceResult struct {
	Confidence float64
	Matches    []MatchResult
}


func getDefaultPageMappings() map[string]PageMappingData {
	return map[string]PageMappingData{
		"home": {
			PageID:      "home",
			URL:         "/",
			DisplayName: "Beranda",
			Keywords:    []string{"beranda", "home", "utama", "awal", "depan"},
			Synonyms:    []string{"halaman utama", "halaman awal", "halaman depan", "menu utama"},
			Category:    "navigation",
			Description: "Halaman utama aplikasi",
		},
		"profile": {
			PageID:      "profile",
			URL:         "/profile",
			DisplayName: "Profil",
			Keywords:    []string{"profil", "profile", "akun", "account"},
			Synonyms:    []string{"profil saya", "akun saya", "data diri", "informasi pribadi"},
			Category:    "user",
			Description: "Halaman profil pengguna",
		},
		"transaction_history": {
			PageID:      "transaction_history",
			URL:         "/transactions",
			DisplayName: "Riwayat Transaksi",
			Keywords:    []string{"riwayat", "transaksi", "history", "transaction"},
			Synonyms: []string{
				"riwayat transaksi", "historis transaksi", "sejarah transaksi",
				"history transaksi", "catatan transaksi", "log transaksi",
				"riwayat pembayaran", "historis pembayaran", "sejarah pembayaran",
				"riwayat pengeluaran", "historis pengeluaran", "catatan pengeluaran",
				"riwayat pembelian", "historis pembelian", "sejarah pembelian",
				"catatan keuangan", "data keuangan", "laporan keuangan",
			},
			Category:    "finance",
			Description: "Riwayat semua transaksi keuangan",
		},
		"wallet": {
			PageID:      "wallet",
			URL:         "/wallet",
			DisplayName: "Dompet",
			Keywords:    []string{"dompet", "wallet", "saldo", "balance"},
			Synonyms: []string{
				"dompet digital", "e-wallet", "saldo dompet", "balance dompet",
				"uang elektronik", "dana", "kredit", "deposit", "tabungan",
			},
			Category:    "finance",
			Description: "Dompet digital dan saldo",
		},
		"settings": {
			PageID:      "settings",
			URL:         "/settings",
			DisplayName: "Pengaturan",
			Keywords:    []string{"pengaturan", "settings", "konfigurasi", "config"},
			Synonyms: []string{
				"pengaturan aplikasi", "setelan", "konfigurasi aplikasi",
				"preferensi", "opsi", "kontrol", "atur aplikasi",
			},
			Category:    "system",
			Description: "Pengaturan aplikasi",
		},
		"notifications": {
			PageID:      "notifications",
			URL:         "/notifications",
			DisplayName: "Notifikasi",
			Keywords:    []string{"notifikasi", "notifications", "pemberitahuan", "alert"},
			Synonyms: []string{
				"pemberitahuan", "pesan masuk", "inbox", "peringatan",
				"info terbaru", "update", "kabar", "berita",
			},
			Category:    "communication",
			Description: "Daftar notifikasi dan pemberitahuan",
		},
		"help": {
			PageID:      "help",
			URL:         "/help",
			DisplayName: "Bantuan",
			Keywords:    []string{"bantuan", "help", "panduan", "guide"},
			Synonyms: []string{
				"pusat bantuan", "help center", "tutorial", "cara menggunakan",
				"petunjuk", "instruksi", "dokumentasi", "faq", "tanya jawab",
			},
			Category:    "support",
			Description: "Pusat bantuan dan panduan",
		},
	}
}

