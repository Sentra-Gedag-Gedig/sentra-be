package voice

import "ProjectGolang/pkg/response"

var (
	ErrInvalidAudioFile     = response.NewError(400, "invalid audio file")
	ErrAudioFileTooLarge    = response.NewError(400, "audio file too large")
	ErrUnsupportedFormat    = response.NewError(400, "unsupported audio format")
	ErrTranscriptionFailed  = response.NewError(500, "failed to transcribe audio")
	ErrNLPProcessingFailed  = response.NewError(500, "failed to process natural language")
	ErrTTSGenerationFailed  = response.NewError(500, "failed to generate speech")
	ErrSessionNotFound      = response.NewError(404, "session not found")
	ErrInvalidSession       = response.NewError(400, "invalid session state")
	ErrCommandNotRecognized = response.NewError(400, "command not recognized")
	ErrPageMappingNotFound  = response.NewError(404, "page mapping not found")
	ErrVoiceCommandFailed   = response.NewError(500, "failed to process voice command")
	ErrUnauthorizedAccess   = response.NewError(403, "unauthorized access to voice features")
	ErrRateLimitExceeded    = response.NewError(429, "rate limit exceeded")
)
