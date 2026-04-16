package service

import (
	"errors"
	"fmt"

	"pfo-vector/internal/model"

	"strings"

	"mime/multipart"

	"github.com/xuri/excelize/v2"
)

var (
    ErrOpenExcelReader      = errors.New("open excel reader")
    ErrEmptyWorkbook        = errors.New("empty workbook")
    ErrFailedReadRows       = errors.New("failed to read rows")
    ErrNoDataRows           = errors.New("no data rows")
    ErrMissingRequiredCol   = errors.New("missing required column")
    ErrTelegramNotUnique    = errors.New("telegram not unique")
)

type UserRow struct {
    FullName    string `json:"full_name"`
    Telegram    string `json:"telegram"`
    PhoneNumber string `json:"phone_number,omitempty"`
}

func ParseExcel(file multipart.File)([]UserRow,[]model.RowError,error){
	xlsx,err := excelize.OpenReader(file)
	if err!=nil{
		return nil,nil,fmt.Errorf("%w: %v",ErrOpenExcelReader,err)
	}
	defer xlsx.Close()


	sheet :=xlsx.GetSheetName(0)
	if sheet == ""{
		return nil,nil,ErrEmptyWorkbook
	}

	rows,err := xlsx.GetRows(sheet)
	if err != nil{
		return nil,nil,ErrFailedReadRows
	}

	if len(rows)<2{
		return  nil,nil,ErrNoDataRows
	}
	headerIndex :=map[string]int{}
	for i,h := range rows[0]{
		headerIndex[strings.TrimSpace(strings.ToLower(h))]=i
	}
	
	required := []string{"фио","тг","телефон"}
	for _,col := range required{
		if _,ok := headerIndex[col];!ok{
			return nil,nil,fmt.Errorf("%w : %s",ErrMissingRequiredCol,col)
			
		}
	}
	rowError := []model.RowError{}
	users := make([]UserRow, 0, len(rows)-1)
	for id,row := range rows[1:]{
		fullname := cell(row,headerIndex,"фио")
		telegram := cell(row,headerIndex,"тг")
		phonenumber := cell(row,headerIndex,"телефон")

		isValid := true
		if strings.TrimSpace(fullname) == ""  {
			rowError = append(rowError, model.RowError{
				Row : id+2,
				Field: "fullname",
				Reason: "Fullname empty",
			})
            isValid = false
        }
		if strings.TrimSpace(telegram) == ""  {
			rowError = append(rowError, model.RowError{
				Row : id+2,
				Field: "telegram",
				Reason: "Telegram empty",
			})
            isValid = false
        }
		if strings.TrimSpace(phonenumber) == ""  {
			rowError = append(rowError, model.RowError{
				Row : id+2,
				Field: "phonenumber",
				Reason: "phonenumber empty",
			})
            isValid = false
        }

		if isValid{
			users = append(users, UserRow{
				FullName: fullname,
				Telegram: telegram,
				PhoneNumber: phonenumber,
			})
		}


	}
	return users,rowError,nil
}




func cell(row []string, idx map[string]int, column string) string {
    i, ok := idx[column]
    if !ok || i >= len(row) {
        return ""
    }
    return strings.TrimSpace(row[i])
}