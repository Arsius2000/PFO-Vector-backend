package service

import "testing"

func TestValidationTelegramUsername(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {
            name:    "valid username with @",
            input:   "@ivan_123",
            wantErr: false,
        },
        {
            name:    "valid username without @",
            input:   "ivan_123",
            wantErr: false,
        },
        {
            name:    "too short",
            input:   "ab",
            wantErr: true,
        },
        {
            name:    "starts with digit",
            input:   "1ivan",
            wantErr: true,
        },
        {
            name:    "starts with underscore",
            input:   "_ivan",
            wantErr: true,
        },
        {
            name:    "invalid chars",
            input:   "ivan-123",
            wantErr: true,
        },
        {
            name:    "empty string",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidationTelegramUsername(tt.input)

            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
            }
        })
    }
}