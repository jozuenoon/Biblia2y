package bible

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
)

// LoadBookIndex - return book maps, find names by book number, find number by book name.
func LoadBookIndex(path string) (map[int][]string, map[string]int, error) {

	if path == "" {
		path = "../data/ksiegi.txt"
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.Comma = ' '

	BooksName := make(map[int][]string)
	BooksValue := make(map[string]int)

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, nil, err
		}
		bookNum, err := strconv.Atoi(row[0])
		if err != nil {
			return nil, nil, err
		}
		BooksValue[row[1]] = bookNum
		BooksName[bookNum] = append(BooksName[bookNum], row[1])
	}

	return BooksName, BooksValue, nil
}

// LoadPlanReferences - return book maps, find names by book number, find number by book name.
func LoadPlanReferences(path string) (map[int][]string, error) {

	if path == "" {
		path = "../data/plan.csv"
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.Comma = ';'
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1 // For variable length of records.

	planRef := make(map[int][]string)

	var idx int

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}
		planRef[idx] = row
		idx++
	}

	return planRef, nil
}

type LoadTextResponse struct {
	IndexMap map[Label]int
	LabelMap map[int]Label
	TextMap  map[int]string
	MaxIndex int
}

func LoadText(path string) (*LoadTextResponse, error) {

	if path == "" {
		path = "../data/bt.txt"
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)

	var idx int

	idxMap := make(map[Label]int)
	labelMap := make(map[int]Label)
	textMap := make(map[int]string)

	for sc.Scan() {
		t := sc.Text()
		sp := strings.SplitN(t, " ", 2)
		idxMap[Label(sp[0])] = idx // save ref for index
		labelMap[idx] = Label(sp[0])
		textMap[idx] = sp[1]
		idx++
	}

	return &LoadTextResponse{
		IndexMap: idxMap,
		LabelMap: labelMap,
		TextMap:  textMap,
		MaxIndex: idx - 1,
	}, nil
}
