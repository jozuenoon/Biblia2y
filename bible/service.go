package bible

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/go-kit/kit/log"
)

type Service interface {
	GetDay(day int) ([]string, error)
	GetDayReferences(day int) ([]string, error)
	GetBookNumber(bookName string) (int, error)
	GetText(idx int) (string, error)
	GetVerseFromIndex(idx int) (*Verse, error)
	GetBookNames(int) ([]string, error)
	GetTextByReference(ref string) (string, error)
	GetIndexFromLabel(Label) (int, error)
	GetChapterStartIndex(int) (int, error)
	GetChapterEndIndex(int) int
	NewVerseFromSingleLabel(Label) (*Verse, error)
	NewVerseFromDualLabel(Label, Label) (*Verse, error)
	GetVerseText(*Verse) ([]string, error)
	GetLabel(int) (Label, error)
}

var _ Service = (*service)(nil)

type service struct {
	// Get book names by book number.
	bookName map[int][]string
	// Get book number by book name.
	bookValue map[string]int
	// Mapping from label (semantic index) into sequential index.
	idxMap map[Label]int
	// Label map - maps index back to label...
	labelLock sync.Mutex
	labelMap  map[int]Label
	// Sequential index to verses.
	textMap map[int]string
	// Plan with references only...
	planRef map[int][]string
	// Plan mapping from plan day into set of verses.
	plan map[int][]string

	// Maximum index value (sequential index).
	maxIndex int

	log log.Logger
}

func (s *service) GetLabel(index int) (Label, error) {
	s.labelLock.Lock()
	defer s.labelLock.Unlock()
	if label, ok := s.labelMap[index]; ok {
		return label, nil
	}
	return Label(""), fmt.Errorf("could not find index")
}

func (s *service) NewVerseFromSingleLabel(label Label) (*Verse, error) {
	start, err := s.GetIndexFromLabel(label)
	if err != nil {
		return nil, err
	}
	if len(label) == 6 { // Handle chapter label.
		end := s.GetChapterEndIndex(start)
		return &Verse{start, end}, nil
	}
	return &Verse{start, 0}, nil
}

func (s *service) NewVerseFromDualLabel(startLabel, endLabel Label) (*Verse, error) {
	start, err := s.GetIndexFromLabel(startLabel)
	if err != nil {
		return nil, err
	}
	end, err := s.GetIndexFromLabel(endLabel)
	if err != nil {
		return nil, err
	}
	if len(endLabel) == 6 {
		end = s.GetChapterEndIndex(end)
		return &Verse{start, end}, nil
	}
	return &Verse{start, end}, nil
}

func (s *service) GetVerseFromIndex(idx int) (*Verse, error) {
	if label, ok := s.labelMap[idx]; ok {
		v, err := s.NewVerseFromSingleLabel(label)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, fmt.Errorf("%d index -> label not found", idx)
}

func (s *service) MaxDay() int {
	return len(s.plan) - 1
}

func (s *service) GetDay(day int) ([]string, error) {
	// Make days to rotate over and over...
	maxDay := s.MaxDay()

	if day > maxDay {
		day %= maxDay
	}

	verses, ok := s.plan[day]
	if !ok {
		return nil, fmt.Errorf("plan day does not exists")
	}
	return verses, nil
}

func (s *service) GetBookNumber(bookName string) (int, error) {
	if num, ok := s.bookValue[bookName]; ok {
		return num, nil
	}
	return 0, fmt.Errorf("book does not exist")
}

func (s *service) GetBookNames(bookNumber int) ([]string, error) {
	bNames, ok := s.bookName[bookNumber]
	if !ok {
		return nil, fmt.Errorf("book does not exist")
	}
	return bNames, nil
}

func (s *service) GetTextByReference(ref string) (string, error) {
	parser := NewParser(strings.NewReader(ref), s)

	verse, err := parser.Parse()
	if err != nil {
		return "", err
	}

	t, err := s.GetVerseText(verse)
	if err != nil {
		return "", err
	}

	return strings.Join(t, " "), nil
}

func (s *service) getIndexFromLabel(label Label) (int, error) {
	if idx, ok := s.idxMap[label]; ok {
		return idx, nil
	}
	return 0, fmt.Errorf("given label does not exist")
}

func (s *service) getIndexFromChapterLabel(label Label) (int, error) {
	// Try to hit any existing verse in label map.
	// Verse 001 should always exist as from my research.
	verseTry := []string{"001", "002", "003"}
	for _, verse := range verseTry {
		if idx, ok := s.idxMap[label+Label(verse)]; ok {
			return idx, nil
		}
	}
	return 0, fmt.Errorf("given label does not exist")
}

// If label is verse label, just returns index directly
// If label is chapter label, it will return start of
// chapter index (which is not always associated with verse 1).
func (s *service) GetIndexFromLabel(label Label) (int, error) {
	if len(label) == 6 { // Handle chapter label.
		idx, err := s.getIndexFromChapterLabel(label)
		if err != nil {
			return 0, err
		}
		return s.GetChapterStartIndex(idx)

	} else if len(label) == 9 { // Handle regular verse label.
		return s.getIndexFromLabel(label)
	}
	return 0, fmt.Errorf("invalid label length: %s", label)
}

// Rewinds to chapter start, passed index must exist.
func (s *service) GetChapterStartIndex(index int) (int, error) {
	currentLabel, ok := s.labelMap[index]
	if !ok {
		return 0, fmt.Errorf("index does not exist: %d", index)
	}
	currentChapter := currentLabel.GetChapter()

	chapterChanged := func(iterIndex int) bool {
		label := s.labelMap[iterIndex]
		if currentChapter != label.GetChapter() {
			return true
		}
		return false
	}

	for i := 0; index-i >= 0; i++ {
		if chapterChanged(index - i) {
			return index - i + 1, nil
		}
	}
	return 0, nil
}

// Rewinds to chapter end, passed index must exist.
func (s *service) GetChapterEndIndex(index int) int {
	currentLabel := s.labelMap[index]
	currentChapter := currentLabel.GetChapter()

	chapterChanged := func(iterIndex int) bool {
		label := s.labelMap[iterIndex]
		if currentChapter != label.GetChapter() {
			return true
		}
		return false
	}

	for i := 0; index+i <= s.maxIndex; i++ {
		if chapterChanged(index + i) {
			return index + i - 1
		}
	}
	return 0
}

func (s *service) GetText(idx int) (string, error) {
	if text, ok := s.textMap[idx]; ok {
		return text, nil
	}
	return "", fmt.Errorf("index does not exist")
}

func (s *service) GetVerseText(verse *Verse) ([]string, error) {
	header, err := s.VerseHeader(verse)
	if err != nil {
		return nil, err
	}
	if verse.IsSingle() {
		vt, err := s.GetText(verse.Start())
		if err != nil {
			return nil, err
		}
		l, err := s.GetLabel(verse.Start())
		if err != nil {
			return nil, err
		}
		return []string{header + "\n", s.getVerseFromLabel(l), vt}, nil
	}

	var texts []string
	for i := verse.Start(); i <= verse.End(); i++ {
		vt, err := s.GetText(i)
		if err != nil {
			return nil, err
		}
		ilvn, err := s.inlineVerseNumber(i)
		if err != nil {
			return nil, err
		}
		texts = append(texts, []string{ilvn, vt}...)
	}
	return append([]string{header + "\n"}, texts...), nil
}

func (s *service) inlineVerseNumber(index int) (string, error) {
	l, err := s.GetLabel(index)
	if err != nil {
		return "", err
	}
	if s.getVerseFromLabel(l) == "1" {
		return strings.Join([]string{s.getChapterFromLabel(l), s.getVerseFromLabel(l)}, ","), nil
	}
	return s.getVerseFromLabel(l), nil
}

// GetIndexHeader returns human readable verse header from index.
// Example: 1 Kor 2,13
func (s *service) GetIndexHeader(idx int) ([]string, error) {
	var header []string
	label, ok := s.labelMap[idx]
	if !ok {
		return nil, fmt.Errorf("can't find verse index")
	}
	bookName, err := s.getBookFromLabel(label)
	if err != nil {
		return nil, err
	}
	return append(header, bookName, s.getChapterFromLabel(label)+",", s.getVerseFromLabel(label)), nil
}

func (s *service) getVerseFromLabel(l Label) string {
	verse := l.GetVerse()
	return strings.TrimLeft(verse, "0")
}

func (s *service) getChapterFromLabel(l Label) string {
	chapter := l.GetChapter()
	return strings.TrimLeft(chapter, "0")
}

func (s *service) getBookFromLabel(l Label) (string, error) {
	book, err := strconv.ParseInt(l.GetBook(), 10, 64)
	if err != nil {
		return "", err
	}

	books, err := s.GetBookNames(int(book))
	if err != nil {
		return "", err
	}
	return books[0], nil
}

func (s *service) VerseHeader(v *Verse) (string, error) {
	var header []string
	if v.IsSingle() {
		label := s.labelMap[v.Start()]
		bookName, err := s.getBookFromLabel(label)
		if err != nil {
			return "", err
		}
		header = append(header, bookName, " ", s.getChapterFromLabel(label), ",", s.getVerseFromLabel(label))
		return strings.Join(header, ""), nil
	}

	startLabel := s.labelMap[v.Start()]
	endLabel := s.labelMap[v.End()]

	switch {
	case startLabel.GetChapter() == endLabel.GetChapter() && startLabel.GetBook() == endLabel.GetBook():
		// Same chapter same book...
		bookName, err := s.getBookFromLabel(startLabel)
		if err != nil {
			return "", err
		}
		header = append(header,
			bookName,
			" ",
			s.getChapterFromLabel(startLabel),
			",",
			s.getVerseFromLabel(startLabel),
			"-",
			s.getVerseFromLabel(endLabel))
		return strings.Join(header, ""), nil
	case startLabel.GetBook() == endLabel.GetBook():
		// Same book...
		bookName, err := s.getBookFromLabel(startLabel)
		if err != nil {
			return "", err
		}
		header = append(header,
			bookName,
			" ",
			s.getChapterFromLabel(startLabel),
			",",
			s.getVerseFromLabel(startLabel),
			"-",
			s.getChapterFromLabel(endLabel),
			",",
			s.getVerseFromLabel(endLabel))
		return strings.Join(header, ""), nil
	default:
		// Maybe cross book...
		startBookName, err := s.getBookFromLabel(startLabel)
		if err != nil {
			return "", err
		}
		endBookName, err := s.getBookFromLabel(startLabel)
		if err != nil {
			return "", err
		}
		header = append(header,
			startBookName,
			" ",
			s.getChapterFromLabel(startLabel),
			",",
			s.getVerseFromLabel(startLabel),
			" - ",
			endBookName,
			" ",
			s.getChapterFromLabel(endLabel),
			",",
			s.getVerseFromLabel(endLabel))
		return strings.Join(header, ""), nil
	}
}

func (s *service) GetDayReferences(day int) ([]string, error) {
	refs, ok := s.planRef[day]
	if !ok {
		return nil, fmt.Errorf("plan day does not exists")
	}
	return refs, nil
}

func (s *service) LoadPlan() error {
	plan := make(map[int][]string)
	for day, refs := range s.planRef {
		var planText []string
		for _, ref := range refs {
			// Omit empty references immediately
			if ref == "" {
				continue
			}
			s.log.Log("msg", "processing", "ref", ref)
			p := NewParser(strings.NewReader(ref), s)
			verse, err := p.Parse()
			if err != nil {
				s.log.Log("msg", "error while parsing ref %s", ref, "err", err)
				continue
			}
			text, err := s.GetVerseText(verse)
			if err != nil {
				s.log.Log("msg", "error while getting text", "err", err, "ref", ref)
				continue
			}
			planText = append(planText, strings.Join(text, " "))
		}
		plan[day] = planText
	}
	s.plan = plan
	return nil
}

func New(booksPath, textPath, planPath string, log log.Logger) (Service, error) {
	// Get books...
	bookName, bookValue, err := LoadBookIndex(booksPath)
	if err != nil {
		return nil, err
	}

	// Load text ...
	text, err := LoadText(textPath)
	if err != nil {
		return nil, err
	}

	// Load plan references ...
	planRef, err := LoadPlanReferences(planPath)
	if err != nil {
		return nil, err
	}

	s := &service{
		planRef:   planRef,
		bookName:  bookName,
		bookValue: bookValue,
		idxMap:    text.IndexMap,
		labelMap:  text.LabelMap,
		textMap:   text.TextMap,
		maxIndex:  text.MaxIndex,
		log:       log,
	}
	// Generate plan enteries...
	err = s.LoadPlan()
	if err != nil {
		return nil, err
	}
	return s, nil
}
