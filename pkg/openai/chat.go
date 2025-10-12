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
	Type       string                 `json:"type"` // "navigation", "transaction", "query"
	Action     string                 `json:"action,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Confidence float64                `json:"confidence"`
	Order      int                    `json:"order"` // execution order for multi-intent
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

RULES:
- Return ONLY valid JSON
- Detect multiple intents if present
- Order intents by execution priority
- Use fuzzy matching for page names
- If page doesn't exist, suggest closest match

RESPONSE FORMAT:
{
  "intents": [
    {
      "type": "transaction",
      "action": "create_expense",
      "data": {
        "amount": 50000,
        "description": "makan siang",
        "category": "makanan"
      },
      "confidence": 0.95,
      "order": 1
    },
    {
      "type": "navigation",
      "action": "navigate",
      "data": {
        "page_id": "profiles",
        "url": "/profiles",
        "display_name": "Profil"
      },
      "confidence": 0.9,
      "order": 2
    }
  ],
  "confidence": 0.92,
  "needs_clarification": false,
  "clarification_question": ""
}

NAVIGATION DATA FIELDS:
- page_id: exact page_id from available pages
- url: exact URL path
- display_name: user-friendly name
- matched_keyword: which keyword matched (optional)

TRANSACTION DATA FIELDS:
- amount: numeric value in IDR
- description: what the transaction is for
- category: one of: makanan, transportasi, belanja, kesehatan, hiburan, gaji, bonus, investasi
- type: "income" or "expense"

EXAMPLES:

Input: "catat pengeluaran makan 50 ribu lalu ke profil"
Output: {
  "intents": [
    {"type":"transaction","action":"create_expense","data":{"amount":50000,"description":"makan","category":"makanan","type":"expense"},"confidence":0.95,"order":1},
    {"type":"navigation","action":"navigate","data":{"page_id":"profiles","url":"/profiles","display_name":"Profil"},"confidence":0.9,"order":2}
  ],
  "confidence":0.92,
  "needs_clarification":false
}

Input: "pindah ke beranda"
Output: {
  "intents":[{"type":"navigation","action":"navigate","data":{"page_id":"home","url":"/home","display_name":"Beranda"},"confidence":0.98,"order":1}],
  "confidence":0.98,
  "needs_clarification":false
}

Input: "ke halaman yang ada qr nya"
Output: {
  "intents":[{"type":"navigation","action":"navigate","data":{"page_id":"qr","url":"/qr","display_name":"QR Code"},"confidence":0.85,"order":1}],
  "confidence":0.85,
  "needs_clarification":false
}

Input: "buka halaman xyz"
Output: {
  "intents":[],
  "confidence":0.3,
  "needs_clarification":true,
  "clarification_question":"Maaf, halaman 'xyz' tidak ditemukan. Apakah maksud Anda: Beranda, Profil, atau Sentra Pay?"
}`, string(pagesJSON))

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
	
	// Clean response if needed
	responseContent = strings.TrimSpace(responseContent)
	
	if err := json.Unmarshal([]byte(responseContent), &result); err != nil {
		return nil, fmt.Errorf("failed to parse multi-intent result: %w - response: %s", err, responseContent)
	}

	return &result, nil
}