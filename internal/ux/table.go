package ux

import (
	"strings"
	"unicode/utf8"
)

type Table struct {
	Header []string
	Rows   [][]string
}

type Row struct {
	Field string
	Value string
}

func NewRow(field, value string) Row {
	return Row{
		Field: field,
		Value: value,
	}
}

type ColFactory[T any] func(T) []Row

func TableFromObjects[T any](objects []T, rowFactory ColFactory[T]) Table {
	var (
		headers []string
		rows    = make([][]string, 0, len(objects))
	)

	for i, object := range objects {
		objRows := rowFactory(object)
		row := make([]string, 0, len(objRows))

		for _, objRow := range objRows {
			if i == 0 {
				headers = append(headers, objRow.Field)
			}

			row = append(row, objRow.Value)
		}

		rows = append(rows, row)
	}

	return Table{
		Header: headers,
		Rows:   rows,
	}
}

func (t *Table) calculateColumnWidths() map[int]int {
	colWidths := make(map[int]int)

	// calculate the longest string in each column
	for i, row := range t.Rows {
		for j, cell := range row {
			cellWidth := utf8.RuneCountInString(cell)
			if cellWidth > colWidths[j] {
				colWidths[j] = cellWidth
			}

			headerWidth := utf8.RuneCountInString(t.Header[j])
			if headerWidth > colWidths[j] {
				colWidths[j] = headerWidth
			}
		}

		if i == 0 {
			for j, header := range t.Header {
				headerWidth := utf8.RuneCountInString(header)
				if headerWidth > colWidths[j] {
					colWidths[j] = headerWidth
				}
			}
		}
	}

	return colWidths
}

func (t *Table) String() string {
	longestInCol := t.calculateColumnWidths()

	var sb strings.Builder

	// print padded cells
	for i, header := range t.Header {
		sb.WriteString(header)
		sb.WriteString(" ")

		headerWidth := utf8.RuneCountInString(header)

		sb.WriteString(strings.Repeat(" ", longestInCol[i]-headerWidth))
		sb.WriteString("\t")
	}

	sb.WriteString("\n")

	for _, row := range t.Rows {
		for i, cell := range row {
			sb.WriteString(cell)
			sb.WriteString(" ")

			cellWidth := utf8.RuneCountInString(cell)

			sb.WriteString(strings.Repeat(" ", longestInCol[i]-cellWidth))
			sb.WriteString("\t")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
