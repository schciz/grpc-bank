package service

import (
	"bytes"
	"fmt"
	"time"

	"grpc-bank/internal/report/domain"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

func buildStatementContent(accountID, startDate, endDate, format string, txs []domain.StatementTransaction) ([]byte, error) {
	switch format {
	case "pdf":
		return buildPDF(accountID, startDate, endDate, txs)
	case "xlsx":
		fallthrough
	default:
		return buildXLSX(accountID, startDate, endDate, txs)
	}
}

func buildXLSX(accountID, startDate, endDate string, txs []domain.StatementTransaction) ([]byte, error) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := "Statement"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"Created At (UTC)", "Transaction ID", "From Account", "To Account", "Amount", "Status"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for idx, t := range txs {
		row := idx + 2
		values := []interface{}{
			t.CreatedAt.UTC().Format(time.RFC3339),
			t.ID, t.FromAccountID, t.ToAccountID, t.Amount, t.Status,
		}
		for i, v := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	_ = f.SetCellValue(sheet, "A"+fmt.Sprint(len(txs)+3), fmt.Sprintf("Account: %s, period: %s..%s", accountID, startDate, endDate))
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildPDF(accountID, startDate, endDate string, txs []domain.StatementTransaction) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Account statement")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 8, fmt.Sprintf("Account: %s  Period: %s..%s", accountID, startDate, endDate))
	pdf.Ln(10)
	headers := []string{"Created At (UTC)", "Tx ID", "From", "To", "Amount", "Status"}
	widths := []float64{55, 55, 45, 45, 30, 25}
	pdf.SetFont("Arial", "B", 9)
	for i, h := range headers {
		pdf.CellFormat(widths[i], 7, h, "1", 0, "", false, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 8)
	for _, t := range txs {
		row := []string{
			t.CreatedAt.UTC().Format("2006-01-02 15:04:05"),
			t.ID, t.FromAccountID, t.ToAccountID, t.Amount, t.Status,
		}
		for i, v := range row {
			pdf.CellFormat(widths[i], 6, truncate(v, 28), "1", 0, "", false, 0, "")
		}
		pdf.Ln(-1)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
