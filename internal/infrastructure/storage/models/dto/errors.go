package dto

import "errors"

var (
	ErrForbidden     = errors.New("access denied")
	ErrActionDenied  = errors.New("action denied")
	ErrAiUnavailable = errors.New("ai model currently unavailable")
	ErrNoAds         = errors.New("no ads available for this client")
	ErrWrongFile     = errors.New("\"Недопустимый формат файла. ")
)
