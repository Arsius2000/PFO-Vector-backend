package service

import (
    "errors"
    "fmt"
    "strings"
    "unicode"
)

var (
    ErrPhoneEmpty         = errors.New("phone_number is empty")
    ErrPhoneInvalidChars  = errors.New("phone_number has invalid characters")
    ErrPhoneInvalidLength = errors.New("phone_number must contain 10 or 11 digits")
    ErrPhoneInvalidPrefix = errors.New("phone_number must start with 7 or 8")
)

func NormalizePhone(raw string) (string, error) {
    s := strings.TrimSpace(raw)
    if s == "" {
        return "", ErrPhoneEmpty
    }

    var digits []rune
    for _, r := range s {
        switch {
        case unicode.IsDigit(r):
            digits = append(digits, r)
        case r == '+' || r == ' ' || r == '-' || r == '(' || r == ')':
            // допустимые разделители
        default:
            return "", ErrPhoneInvalidChars
        }
    }

    d := string(digits)
    switch len(d) {
    case 10:
        d = "8" + d
    case 11:
        if d[0] == '7' {
            d = "8" + d[1:]
        }
        if d[0] != '8' {
            return "", ErrPhoneInvalidPrefix
        }
    default:
        return "", ErrPhoneInvalidLength
    }

    return fmt.Sprintf("8%s%s%s%s", d[1:4], d[4:7], d[7:9], d[9:11]), nil
}