package helpers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/goccy/go-json"
)

type (
	IParser interface {
		ToString(v interface{}) string
		ToInt(v interface{}) (int, error)
		Marshal(source interface{}) ([]byte, error)
		Unmarshal(src []byte, dest interface{}) error
		CsvToJson(pool bool, record []byte) ([]map[string]interface{}, error)
	}

	parser struct{}
)

func NewParser() IParser {
	return &parser{}
}

func (h *parser) ToString(v interface{}) string {
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}

func (h *parser) ToInt(v interface{}) (int, error) {
	parse, err := strconv.Atoi(h.ToString(v))
	if err != nil {
		return 0, nil
	}

	return parse, nil
}

func (h *parser) Marshal(src interface{}) ([]byte, error) {
	return json.Marshal(src)
}

func (h *parser) Unmarshal(src []byte, dest interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(src))

	for decoder.More() {
		if err := decoder.Decode(dest); err != nil {
			return err
		}
	}

	return nil
}

func (h *parser) CsvToJson(pool bool, record []byte) ([]map[string]interface{}, error) {
	csvReader := csv.NewReader(bytes.NewBuffer(record))
	csvRecords, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	headers := csvRecords[:1][0]
	recordsLength := len(csvRecords)
	records := csvRecords[1:recordsLength]

	switch pool {
	case true:
		return h.reformatCsvWithPool(headers, records), nil

	default:
		return h.reformatCsvWithoutPool(headers, records), nil
	}
}

func (h *parser) reformatCsvWithoutPool(headers []string, record [][]string) []map[string]interface{} {
	var (
		csvRecord       map[string]interface{} = make(map[string]interface{})
		csvRecordLength int                    = len(record)
	)

	if csvRecordLength == 0 {
		return nil
	}

	if csvRecordLength == 1 {
		for i, v := range headers {
			csvRecord[v] = record[0][i]
		}

		return []map[string]interface{}{csvRecord}
	}

	mid := len(record) / 2
	left := h.reformatCsvWithoutPool(headers, record[:mid])
	right := h.reformatCsvWithoutPool(headers, record[mid:])

	return append(left, right...)
}

func (h *parser) reformatCsvWithPool(headers []string, record [][]string) []map[string]interface{} {
	var (
		po              *sync.Pool             = new(sync.Pool)
		csvRecord       map[string]interface{} = make(map[string]interface{})
		csvRecordLength int                    = len(record)
	)

	if csvRecordLength == 0 {
		return nil
	}

	if csvRecordLength == 1 {
		po.New = func() interface{} {
			return csvRecord
		}

		for i, v := range headers {
			csvRecord[v] = record[0][i]
			po.Put(csvRecord)
		}

		return []map[string]interface{}{po.Get().(map[string]interface{})}
	}

	mid := len(record) / 2
	left := h.reformatCsvWithPool(headers, record[:mid])
	right := h.reformatCsvWithPool(headers, record[mid:])

	return append(left, right...)
}
