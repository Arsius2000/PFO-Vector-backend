package service

import (
	"context"

	"mime/multipart"
	"pfo-vector/internal/model"
	"pfo-vector/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserImportService struct {
	queries repository.Querier
}

func NewUserImportService(q repository.Querier) *UserImportService {
	return &UserImportService{queries: q}
}

func (s *UserImportService) ImportFromExcel(ctx context.Context, file multipart.File) (model.ImportResult, error) {

	rows, rowsErrors, err := ParseExcel(file)
	if err != nil {
		return model.ImportResult{}, err
	}

	created := 0
	failed := 0
	for id, row := range rows {

		args := repository.CreateUserParams{
			FullName: row.FullName,
			Telegram: row.Telegram,
			PhoneNumber: pgtype.Text{
				String: row.PhoneNumber,
				Valid:  row.PhoneNumber != "", // false => NULL в БД
			},
			Gender: pgtype.Text{
				String: row.Gender,
				Valid: row.Gender!="",
			},
			
		}

		_, err := s.queries.CreateUser(ctx, args)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				rowsErrors = append(rowsErrors, model.RowError{Row: id, Field: "Telegram", Reason: "Telegram not unique"})
				failed++
				continue
			}
			return model.ImportResult{}, err

		}
		created++

	}
	return model.ImportResult{TotalRows: created + failed, Created: created, Failed: failed, Errors: rowsErrors}, nil
}
