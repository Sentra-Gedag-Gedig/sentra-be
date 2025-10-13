package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type IChatGPT interface {
	ProcessConversation(ctx context.Context, userMessage string, conversationHistory []ConversationMessage) (string, error)
	ProcessTransactionIntent(ctx context.Context, userMessage string) (*TransactionIntent, error)
	ProcessMultiIntent(ctx context.Context, userMessage string, availablePages []PageInfo) (*MultiIntentResult, error)
}

type ConversationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TransactionIntent struct {
	IsTransaction     bool    `json:"is_transaction"`
	Type              string  `json:"type"`
	Amount            float64 `json:"amount"`
	Description       string  `json:"description"`
	SuggestedCategory string  `json:"suggested_category"`
	Confidence        float64 `json:"confidence"`
}

type MultiIntentResult struct {
	Intents    []Intent `json:"intents"`
	Confidence float64  `json:"confidence"`
	NeedsClarification bool `json:"needs_clarification"`
	ClarificationQuestion string `json:"clarification_question,omitempty"`
}

type Intent struct {
	Type       string                 `json:"type"` 
	Action     string                 `json:"action,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Confidence float64                `json:"confidence"`
	Order      int                    `json:"order"` 
}

type PageInfo struct {
	PageID      string   `json:"page_id"`
	URL         string   `json:"url"`
	DisplayName string   `json:"display_name"`
	Keywords    []string `json:"keywords"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
}

type chatGPTService struct {
	client *openai.Client
	model  string
}

func NewChatGPT() IChatGPT {
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("OPENAI_CHAT_MODEL")
	
	if model == "" {
		model = openai.GPT4
	}

	return &chatGPTService{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

func (c *chatGPTService) ProcessConversation(
	ctx context.Context,
	userMessage string,
	conversationHistory []ConversationMessage,
) (string, error) {
	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: `Kamu adalah Sentra AI, asisten suara untuk aplikasi keuangan Sentra yang membantu tunanetra dan low vision. 

Tugas kamu:
1. Membantu pengguna navigasi aplikasi dengan natural conversation
2. Membantu mencatat transaksi keuangan (pemasukan/pengeluaran)
3. Memberikan ringkasan keuangan
4. Menjawab pertanyaan seputar fitur aplikasi

Aturan penting:
- SELALU jawab dalam Bahasa Indonesia yang ringkas dan jelas
- Jawaban maksimal 2-3 kalimat
- Jika pengguna ingin mencatat transaksi, extract: tipe (pemasukan/pengeluaran), nominal, dan deskripsi
- Jika ada yang tidak jelas, tanyakan dengan spesifik
- Untuk navigasi, sebutkan nama menu dengan jelas

Contoh:
User: "Tolong catat tadi saya beli donat 5 ribu"
Assistant: "Baik, saya catat pengeluaran Rp5.000 untuk beli donat, kategori Makanan. Apakah sudah benar?"

User: "Saya mau lihat riwayat transaksi"
Assistant: "Baik, membuka halaman Riwayat Transaksi."`,
		},
	}

	for _, msg := range conversationHistory {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userMessage,
	})

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       c.model,
			Messages:    messages,
			Temperature: 0.7,
			MaxTokens:   150,
		},
	)

	if err != nil {
		return "", fmt.Errorf("ChatGPT API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from ChatGPT")
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *chatGPTService) ProcessTransactionIntent(
	ctx context.Context,
	userMessage string,
) (*TransactionIntent, error) {
	systemPrompt := `You are a transaction analyzer. Detect if user wants to record financial transaction.

IMPORTANT: Return ONLY valid JSON, nothing else.

Format:
{
  "is_transaction": true,
  "type": "expense",
  "amount": 15000,
  "description": "beli kopi di starbucks",
  "suggested_category": "makanan",
  "confidence": 0.9
}

Rules:
- type: "income" or "expense"
- amount: numeric value in IDR
- suggested_category: one of: "makanan", "transportasi", "belanja", "kesehatan", "hiburan", "gaji", "bonus", "investasi"

Income keywords: pemasukan, terima, dapat, gaji, bonus, pendapatan
Expense keywords: pengeluaran, beli, bayar, belanja

Example input: "tadi saya beli kopi 15 ribu di starbucks"
Example output: {"is_transaction":true,"type":"expense","amount":15000,"description":"beli kopi di starbucks","suggested_category":"makanan","confidence":0.9}`

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userMessage,
		},
	}

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       c.model,
			Messages:    messages,
			Temperature: 0.3,
			MaxTokens:   200,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("ChatGPT API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from ChatGPT")
	}

	var intent TransactionIntent
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &intent); err != nil {
		return nil, fmt.Errorf("failed to parse transaction intent: %w", err)
	}

	return &intent, nil
}

func (c *chatGPTService) ProcessMultiIntent(
	ctx context.Context,
	userMessage string,
	availablePages []PageInfo,
) (*MultiIntentResult, error) {
	pagesJSON, _ := json.Marshal(availablePages)
	
	systemPrompt := fmt.Sprintf(`You are an intent analyzer for Sentra AI voice assistant. Analyze user command and detect ALL intents.

AVAILABLE PAGES:
%s

INTENT TYPES:
1. navigation - User wants to go to a page
2. transaction - User wants to record income/expense
3. query - General question or conversation
4. delete_transaction - User wants to delete transaction(s) by AMOUNT and optional DESCRIPTION
5. logout - User wants to logout (ALWAYS needs confirmation)

TRANSACTION DETECTION RULES:
**INCOME keywords**: pemasukan, terima, dapat, gaji, bonus, pendapatan, masuk, diterima
**EXPENSE keywords**: pengeluaran, beli, bayar, belanja, keluar, keluarkan, catat, habis, spent

**Amount patterns**:
- "Rp200.000" / "Rp 200.000" / "Rp200000"
- "200 ribu" / "200ribu"
- "200rb" / "200k"
- "dua ratus ribu"

**Category inference**:
- Starbucks, kopi, cafe → "makanan"
- Grab, Gojek, bensin → "transportasi"
- baju, sepatu → "pakaian"
- obat, dokter → "kesehatan"
- If unclear → "lainnya"

TRANSACTION DATA FIELDS:
- type: "income" or "expense" (REQUIRED)
- amount: numeric value in IDR (REQUIRED)
- description: what the transaction is about (REQUIRED)
- suggested_category: best matching category from valid list

VALID CATEGORIES:
**Income**: gaji, bonus, investasi, part time, lainnya
**Expense**: makanan, sehari-hari, transportasi, sosial, perumahan, hadiah, komunikasi, pakaian, hiburan, tampilan, kesehatan, pajak, pendidikan, investasi, peliharaan, liburan, lainnya

RESPONSE FORMAT:
{
  "intents": [
    {
      "type": "transaction",
      "action": "create_transaction",
      "data": {
        "type": "expense",
        "amount": 200000,
        "description": "beli kopi di Starbucks",
        "suggested_category": "makanan"
      },
      "confidence": 0.95,
      "order": 1
    }
  ],
  "confidence": 0.95,
  "needs_clarification": false
}

TRANSACTION EXAMPLES:

Input: "Tolong dong, catat pengeluaran saya sebesar Rp200.000 di Starbucks."
Output: {
  "intents":[{
    "type":"transaction",
    "action":"create_transaction",
    "data":{
      "type":"expense",
      "amount":200000,
      "description":"beli kopi di Starbucks",
      "suggested_category":"makanan"
    },
    "confidence":0.95,
    "order":1
  }],
  "confidence":0.95,
  "needs_clarification":false
}

Input: "Catat pengeluaran 50 ribu untuk beli nasi goreng"
Output: {
  "intents":[{
    "type":"transaction",
    "action":"create_transaction",
    "data":{
      "type":"expense",
      "amount":50000,
      "description":"beli nasi goreng",
      "suggested_category":"makanan"
    },
    "confidence":0.98,
    "order":1
  }],
  "confidence":0.98,
  "needs_clarification":false
}

Input: "Tadi saya belanja di Indomaret 100 ribu"
Output: {
  "intents":[{
    "type":"transaction",
    "action":"create_transaction",
    "data":{
      "type":"expense",
      "amount":100000,
      "description":"belanja di Indomaret",
      "suggested_category":"sehari-hari"
    },
    "confidence":0.92,
    "order":1
  }],
  "confidence":0.92,
  "needs_clarification":false
}

Input: "Terima gaji 5 juta"
Output: {
  "intents":[{
    "type":"transaction",
    "action":"create_transaction",
    "data":{
      "type":"income",
      "amount":5000000,
      "description":"gaji bulanan",
      "suggested_category":"gaji"
    },
    "confidence":0.99,
    "order":1
  }],
  "confidence":0.99,
  "needs_clarification":false
}

Input: "Bayar Grab 25 ribu"
Output: {
  "intents":[{
    "type":"transaction",
    "action":"create_transaction",
    "data":{
      "type":"expense",
      "amount":25000,
      "description":"bayar Grab",
      "suggested_category":"transportasi"
    },
    "confidence":0.97,
    "order":1
  }],
  "confidence":0.97,
  "needs_clarification":false
}

DELETE TRANSACTION EXAMPLES:

Input: "Hapus transaksi 15 ribu"
Output: {
  "intents":[{
    "type":"delete_transaction",
    "action":"delete",
    "data":{
      "amount":15000
    },
    "confidence":0.95,
    "order":1
  }],
  "confidence":0.95,
  "needs_clarification":false
}

Input: "Hapus transaksi 50 ribu beli kopi"
Output: {
  "intents":[{
    "type":"delete_transaction",
    "action":"delete",
    "data":{
      "amount":50000,
      "description":"kopi"
    },
    "confidence":0.95,
    "order":1
  }],
  "confidence":0.95,
  "needs_clarification":false
}

NAVIGATION EXAMPLES:

Input: "Buka Profil"
Output: {
  "intents":[{"type":"navigation","action":"navigate","data":{"page_id":"profiles","url":"/profiles","display_name":"Profil"},"confidence":0.98,"order":1}],
  "confidence":0.98,
  "needs_clarification":false
}

Input: "Buka Deteksi"
Output: {
  "intents":[{"type":"navigation","action":"navigate","data":{"page_id":"money-detection","url":"/deteksi/money-detection","display_name":"Deteksi Uang"},"confidence":0.95,"order":1}],
  "confidence":0.95,
  "needs_clarification":false
}

LOGOUT EXAMPLES:

Input: "Logout"
Output: {
  "intents":[{"type":"logout","action":"request_logout","data":{"page_id":"logout","url":"/logout","display_name":"Logout"},"confidence":0.99,"order":1}],
  "confidence":0.99,
  "needs_clarification":false
}

IMPORTANT RULES:
1. ALWAYS detect "catat pengeluaran X di Y" as transaction
2. ALWAYS extract amount even with "Rp" prefix
3. ALWAYS infer category based on description keywords
4. If user says "pengeluaran" → type is "expense"
5. If user says "pemasukan"/"terima"/"dapat" → type is "income"
6. Return JSON ONLY, no other text

CLARIFICATION NEEDED ONLY IF:
- Amount is completely unclear or missing
- User says something ambiguous like "catat transaksi" without details`, string(pagesJSON))

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userMessage,
		},
	}

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       c.model,
			Messages:    messages,
			Temperature: 0.3, // Lebih rendah untuk lebih konsisten
			MaxTokens:   500,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("ChatGPT API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from ChatGPT")
	}

	var result MultiIntentResult
	responseContent := resp.Choices[0].Message.Content
	responseContent = strings.TrimSpace(responseContent)
	
	if err := json.Unmarshal([]byte(responseContent), &result); err != nil {
		return nil, fmt.Errorf("failed to parse multi-intent result: %w - response: %s", err, responseContent)
	}

	return &result, nil
}