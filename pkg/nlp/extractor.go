package nlp

import (
	"regexp"
	"strconv"
	"strings"
)

type TransactionData struct {
	Type        string  // "income" or "expense"
	Amount      float64
	Description string
	Category    string
	Confidence  float64
}

type NumberExtractor struct {
	numberWords map[string]float64
}

func NewNumberExtractor() *NumberExtractor {
	return &NumberExtractor{
		numberWords: map[string]float64{
			// Units
			"nol":       0,
			"satu":      1,
			"dua":       2,
			"tiga":      3,
			"empat":     4,
			"lima":      5,
			"enam":      6,
			"tujuh":     7,
			"delapan":   8,
			"sembilan":  9,
			"sepuluh":   10,
			"sebelas":   11,
			"seratus":   100,
			"seribu":    1000,
			"sejuta":    1000000,
			
			// Tens
			"puluh":     10,
			"belas":     10,
			"ratus":     100,
			"ribu":      1000,
			"juta":      1000000,
			"miliar":    1000000000,
			
			// Common shortcuts
			"rebu":      1000,
			"rebuan":    1000,
		},
	}
}

func (ne *NumberExtractor) ExtractAmount(text string) (float64, string) {
	text = strings.ToLower(text)
	
	// Pattern 1: Numeric with separator (50.000, 1.000.000)
	numPattern := regexp.MustCompile(`(\d{1,3}(?:\.\d{3})*(?:,\d+)?)`)
	if matches := numPattern.FindString(text); matches != "" {
		// Remove dots (thousand separator) and replace comma with dot
		cleaned := strings.ReplaceAll(matches, ".", "")
		cleaned = strings.ReplaceAll(cleaned, ",", ".")
		if amount, err := strconv.ParseFloat(cleaned, 64); err == nil {
			return amount, "numeric"
		}
	}
	
	// Pattern 2: Number + unit (50 ribu, 1 juta)
	unitPattern := regexp.MustCompile(`(\d+)\s*(ribu|rebu|juta|jt|rb|k)`)
	if matches := unitPattern.FindStringSubmatch(text); len(matches) > 0 {
		num, _ := strconv.ParseFloat(matches[1], 64)
		unit := matches[2]
		
		multiplier := 1.0
		switch unit {
		case "ribu", "rebu", "rb", "k":
			multiplier = 1000
		case "juta", "jt":
			multiplier = 1000000
		}
		
		return num * multiplier, "unit"
	}
	
	// Pattern 3: Indonesian number words
	amount := ne.parseIndonesianNumber(text)
	if amount > 0 {
		return amount, "words"
	}
	
	return 0, "none"
}

func (ne *NumberExtractor) parseIndonesianNumber(text string) float64 {
	words := strings.Fields(text)
	total := 0.0
	current := 0.0
	
	for _, word := range words {
		word = strings.ToLower(word)
		
		if val, exists := ne.numberWords[word]; exists {
			if val >= 1000 {
				if current == 0 {
					current = 1
				}
				total += current * val
				current = 0
			} else if val == 100 {
				if current == 0 {
					current = 1
				}
				current *= val
			} else {
				current += val
			}
		}
	}
	
	return total + current
}

func (ne *NumberExtractor) ExtractTransactionType(text string) string {
	text = strings.ToLower(text)
	
	incomeKeywords := []string{
		"pemasukan", "masuk", "terima", "dapat", "pendapatan", 
		"gaji", "bonus", "income", "kredit", "transfer masuk",
	}
	
	expenseKeywords := []string{
		"pengeluaran", "keluar", "bayar", "beli", "belanja", 
		"expense", "debit", "transfer keluar", "kirim",
	}
	
	for _, keyword := range incomeKeywords {
		if strings.Contains(text, keyword) {
			return "income"
		}
	}
	
	for _, keyword := range expenseKeywords {
		if strings.Contains(text, keyword) {
			return "expense"
		}
	}
	
	return "unknown"
}

func (ne *NumberExtractor) ExtractDescription(text string, amount float64) string {
	text = strings.ToLower(text)
	
	// Remove transaction type keywords
	removeKeywords := []string{
		"tambah", "catat", "buat", "bikin",
		"pemasukan", "pengeluaran", "transaksi",
		"untuk", "buat", "keperluan",
	}
	
	for _, keyword := range removeKeywords {
		text = strings.ReplaceAll(text, keyword, "")
	}
	
	// Remove amount-related words
	amountPattern := regexp.MustCompile(`\d+\s*(ribu|rebu|juta|rb|jt|k)?`)
	text = amountPattern.ReplaceAllString(text, "")
	
	// Clean up
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	
	if text == "" {
		return "Transaksi"
	}
	
	return text
}

func (ne *NumberExtractor) IdentifyCategory(description string, transactionType string) string {
	description = strings.ToLower(description)
	
	if transactionType == "income" {
		categoryMap := map[string][]string{
			"gaji": {"gaji", "salary", "upah"},
			"bonus": {"bonus", "thr", "insentif"},
			"investasi": {"investasi", "dividen", "saham", "return"},
			"part time": {"freelance", "sampingan", "part time"},
		}
		
		for category, keywords := range categoryMap {
			for _, keyword := range keywords {
				if strings.Contains(description, keyword) {
					return category
				}
			}
		}
		
		return "gaji" // default for income
	}
	
	if transactionType == "expense" {
		categoryMap := map[string][]string{
			"makanan": {"makan", "food", "lapar", "jajan", "snack", "kopi", "minum", "resto", "warung"},
			"transportasi": {"transport", "ojek", "grab", "gojek", "bensin", "parkir", "tol"},
			"belanja": {"belanja", "shopping", "beli", "toko"},
			"hiburan": {"hiburan", "nonton", "film", "game", "konser"},
			"kesehatan": {"obat", "dokter", "rumah sakit", "apotek"},
		}
		
		for category, keywords := range categoryMap {
			for _, keyword := range keywords {
				if strings.Contains(description, keyword) {
					return category
				}
			}
		}
		
		return "sehari-hari" // default for expense
	}
	
	return ""
}

func (ne *NumberExtractor) ExtractTransaction(text string) (*TransactionData, error) {
	// Extract amount
	amount, amountType := ne.ExtractAmount(text)
	if amount == 0 {
		return nil, nil
	}
	
	// Extract transaction type
	txType := ne.ExtractTransactionType(text)
	if txType == "unknown" {
		return nil, nil
	}
	
	// Extract description
	description := ne.ExtractDescription(text, amount)
	
	// Identify category
	category := ne.IdentifyCategory(description, txType)
	
	// Calculate confidence
	confidence := 0.7
	if amountType == "numeric" {
		confidence = 0.9
	} else if amountType == "unit" {
		confidence = 0.8
	}
	
	if category != "" {
		confidence += 0.1
	}
	
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return &TransactionData{
		Type:        txType,
		Amount:      amount,
		Description: description,
		Category:    category,
		Confidence:  confidence,
	}, nil
}