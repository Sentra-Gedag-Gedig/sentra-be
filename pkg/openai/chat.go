package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

type IChatGPT interface {
	ProcessConversation(ctx context.Context, userMessage string, conversationHistory []ConversationMessage) (string, error)
	ProcessTransactionIntent(ctx context.Context, userMessage string) (*TransactionIntent, error)
}

type ConversationMessage struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

type TransactionIntent struct {
	IsTransaction   bool    `json:"is_transaction"`
	Type            string  `json:"type"`            // "income" or "expense"
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	SuggestedCategory string `json:"suggested_category"`
	Confidence      float64 `json:"confidence"`
}

type chatGPTService struct {
	client *openai.Client
	model  string
}

func NewChatGPT() IChatGPT {
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("OPENAI_CHAT_MODEL")
	
	if model == "" {
		model = openai.GPT4 // or GPT3Dot5Turbo for cheaper option
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

	// Add conversation history
	for _, msg := range conversationHistory {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add current user message
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

	// Parse JSON response
	var intent TransactionIntent
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &intent); err != nil {
		return nil, fmt.Errorf("failed to parse transaction intent: %w", err)
	}

	return &intent, nil
}