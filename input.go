package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func read(filename string) ([][]float64, []float64, error) {
	const paramCount = 2
	var (
		x [][]float64
		y []float64
	)
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open %q", filename)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = paramCount + 1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read the file %q", filename)
	}

	for _, record := range records {
		params := make([]float64, paramCount)
		for i := 0; i < paramCount; i++ {
			params[i], err = strconv.ParseFloat(record[i], 64)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse parameter %q", record[i])
			}
		}
		x = append(x, params)
		label, err := strconv.Atoi(record[paramCount])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse label %q", record[paramCount])
		}
		y = append(y, float64(label))
	}
	return x, y, nil
}
