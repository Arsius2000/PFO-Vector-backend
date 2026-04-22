package service

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

func ValidationTelegramUsername(username string) (error) {

	raw := strings.TrimPrefix(username, "@")

	// 2. Проверка на длину
	if len(raw) < 5 {
		return errors.New("слишком короткий (минимум 5 символов)")
	}
	if len(raw) > 32 {
		return errors.New("слишком длинный (максимум 32 символа)")
	}
	// 4. Проверка на первый символ (не может быть цифрой или _)
	first := rune(raw[0])
	if unicode.IsDigit(first) || first == '_' {
		return errors.New("должен начинаться с буквы")
	}

	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validPattern.MatchString(raw) {
		return errors.New("разрешены только латинские буквы, цифры и '_'")
	}

	return  nil
}