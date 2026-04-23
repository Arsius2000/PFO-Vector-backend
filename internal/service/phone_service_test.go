package service

import "testing"

func TestNormalizePhone(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr error
    }{
        {
            name:    "valid +7 format",
            input:   "+7 (999) 123-45-67",
            want:    "89991234567",
            wantErr: nil,
        },
        {
            name:    "valid 8 format",
            input:   "8(999)123-45-67",
            want:    "89991234567",
            wantErr: nil,
        },
        {
            name:    "empty string",
            input:   "",
            want:    "",
            wantErr: ErrPhoneEmpty,
        },
        {
            name:    "invalid chars",
            input:   "8-999-123-45-67#",
            want:    "",
            wantErr: ErrPhoneInvalidChars,
        },
        {
            name:    "invalid length",
            input:   "8999123456",
            want:    "",
            wantErr: ErrPhoneInvalidLength,
        },
        {
            name:    "invalid prefix",
            input:   "99991234567",
            want:    "",
            wantErr: ErrPhoneInvalidPrefix,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NormalizePhone(tt.input)

            if err != tt.wantErr {
                t.Fatalf("want err %v, got %v", tt.wantErr, err)
            }
            if got != tt.want {
                t.Fatalf("want %q, got %q", tt.want, got)
            }
        })
    }
}