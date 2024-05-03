package ux

import "strings"

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
		var row []string

		objRows := rowFactory(object)

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
			if len(cell) > colWidths[j] {
				colWidths[j] = len(cell)
			}

			if len(t.Header[j]) > colWidths[j] {
				colWidths[j] = len(t.Header[j])
			}
		}

		if i == 0 {
			for j, header := range t.Header {
				if len(header) > colWidths[j] {
					colWidths[j] = len(header)
				}
			}
		}
	}

	return colWidths
}

func (t *Table) String() string {
	var (
		s            string
		longestInCol = t.calculateColumnWidths()
	)

	// print padded cells
	for i, header := range t.Header {
		s += header + " "
		s += strings.Repeat(" ", longestInCol[i]-len(header)) + "\t"
	}

	s += "\n"

	for _, row := range t.Rows {
		for i, cell := range row {
			s += cell + " "
			s += strings.Repeat(" ", longestInCol[i]-len(cell)) + "\t"
		}

		s += "\n"
	}

	return s
}
