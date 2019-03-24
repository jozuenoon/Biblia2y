package messenger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jasonlvhit/gocron"
	"github.com/jozuenoon/biblia2y/bible"
	"github.com/jozuenoon/biblia2y/models"
	"github.com/jozuenoon/biblia2y/poster"
	"github.com/syndtr/goleveldb/leveldb"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

type ParseMessageInput struct {
	SenderID  string
	TimeStamp int
	Message   string
}

type ParseMessageOutput struct {
	SenderID string
	Message  []string
}

type Service interface {
	// set time 8:30 - set time of daily event
	// set day 1 - set day of schedule
	// show day 1 - show day 1 verses
	// start - schedule sender for bible plan
	// stop - remove sender from bible plan
	ParseMessage(*ParseMessageInput) *ParseMessageOutput

	// Recover after down time...
	Recover() error

	// Get response sing...
	ResponseSink() chan<- *ParseMessageOutput
}

func New(
	dbPath,
	pageAccessToken,
	facebookAPI,
	booksPath,
	textPath,
	planPath string,
	log log.Logger,
	done <-chan struct{},
) (Service, error) {
	// Create level db, this is concurrently safe....
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}

	// Get bible service...
	bsvc, err := bible.New(booksPath, textPath, planPath, log)
	if err != nil {
		return nil, err
	}
	// Get poster service...
	psvc := poster.New(pageAccessToken, facebookAPI, log)

	messages := make(chan *ParseMessageOutput, 50)

	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-messages:
				err := psvc.ProcessMessages(msg.SenderID, msg.Message, "", "RESPONSE", "REGULAR")
				if err != nil {
					log.Log("msg", "failed to process message", "err", err)
				}
			}
		}
	}()

	s := &service{
		DB:              db,
		psvc:            psvc,
		bsvc:            bsvc,
		log:             log,
		Schedulers:      make(map[string]*SchedulerTask),
		responses:       messages,
		pageAccessToken: pageAccessToken,
	}

	if err := s.Recover(); err != nil {
		return nil, err
	}
	return s, nil
}

type service struct {
	// Persistent database...
	DB              *leveldb.DB
	Schedulers      map[string]*SchedulerTask
	log             log.Logger
	bsvc            bible.Service
	psvc            poster.Service
	responses       chan *ParseMessageOutput
	schLock         sync.RWMutex
	pageAccessToken string
}

const (
	startCommand   = "start"
	stopCommand    = "stop"
	helpCommand    = "help"
	setTimeCommand = "set time"
	showDayCommand = "show day"
	setDayCommand  = "set day"
	infoCommand    = "info"
)

var help = `*Help:*
- *set time 8:30* - set time of daily event
- *set day 1* - set day of schedule
- *show day 1* - show day 1 verses
- *start* - start my schedule
- *stop* - remove me from bible plan
- *dz 1,1* - write this verse
- *info* - show current schedule information
`

func (s *service) ParseMessage(in *ParseMessageInput) *ParseMessageOutput {
	if in == nil {
		s.log.Log("msg", "empty message")
		return nil
	}
	out := make([]string, 0)
	add := func(in string) {
		out = append(out, in)
	}

	in.Message = strings.ToLower(in.Message)

	// Check if message parses to verse...
	verseText, _ := s.bsvc.GetTextByReference(in.Message)

	switch {
	case verseText != "":
		add(verseText)
	case in.Message == startCommand:
		add(s.Start(in.SenderID))
	case in.Message == stopCommand:
		add(s.Stop(in.SenderID))
	case in.Message == helpCommand:
		add(help)
	case strings.HasPrefix(in.Message, setTimeCommand):
		add(s.SetTime(in.Message, in.SenderID))
	case strings.HasPrefix(in.Message, showDayCommand):
		for _, showDayMsg := range s.ShowDay(in.Message) {
			add(showDayMsg)
		}
	case strings.HasPrefix(in.Message, setDayCommand):
		add(s.SetDay(in.Message, in.SenderID))
	case strings.HasPrefix(in.Message, infoCommand):
		add(s.Info(in.SenderID))
	default:
		add("Sorry I don't understand: \n" + in.Message)
		add(help)
	}

	s.log.Log("msg", in.Message, "senderID", in.SenderID)
	return &ParseMessageOutput{
		SenderID: in.SenderID,
		Message:  out,
	}
}

func (s *service) ResponseSink() chan<- *ParseMessageOutput {
	return s.responses
}

func (s *service) Stop(senderID string) string {
	err := s.DB.Delete([]byte(senderID), nil)
	if err != nil {
		return fmt.Sprintf("Error while deleting user %s", err)
	}
	s.schLock.Lock()
	delete(s.Schedulers, senderID)
	s.schLock.Unlock()
	return "Your subscription was successfully removed."
}

// Start will persist sender and schedule tasks...
func (s *service) Start(senderID string) string {
	var message string
	// Check if it doesn't exists...
	data, err := s.DB.Get([]byte(senderID), nil)
	if err == nil {
		var userData User
		err = Unmarshal(data, &userData)
		if err == nil {
			message = fmt.Sprintf(
				`You have bible verses scheduled at %s, currently you are at day %d.
If you want to reset your schedule unsubscribe with *stop* command first.`,
				userData.ScheduleTime.Format("15:04"),
				userData.CurrentDay)
		} else {
			s.log.Log("msg", "error while unmarshalling", "user_id", senderID, "err", err)
			message = fmt.Sprintf("Your user exists but seems to have some error %s", err)
		}
		return message
	}

	// Save new user for recovery...
	userData := User{
		SenderID:     senderID,
		ScheduleTime: time.Now().Add(1 * time.Minute),
		CurrentDay:   0,
	}

	err = PutUserData(&userData, s.DB)
	if err != nil {
		s.log.Log("msg", "error while saving user", "user_id", senderID, "err", err)
		return fmt.Sprintf("Your user can't be saved %s", err.Error())
	}

	// Create scheduler...
	err = s.AddScheduler(&userData)
	if err != nil {
		s.log.Log("msg", "error while adding scheduler", "user_id", senderID, "err", err)
		return fmt.Sprintf("Can't create scheduler, please retry: %s", err.Error())
	}

	s.log.Log("msg", "user saved and scheduled", "user_id", senderID)
	message = fmt.Sprintf(
		"You have bible read plan scheduled at %s, currently you are at day %d",
		userData.ScheduleTime.Format("15:04"),
		userData.CurrentDay,
	)
	return message
}

// https://graph.facebook.com/v2.6/USER_ID?fields=first_name,last_name,profile_pic,locale,timezone,gender&access_token=PAGE_ACCESS_TOKEN
func (s *service) getUserDetails(senderID string) *models.User {
	client := &http.Client{}
	defaultUser := &models.User{
		Timezone: 1,
	}
	fbAPI := fmt.Sprintf("https://graph.facebook.com/v2.6/%s?fields=name,first_name,last_name,timezone&access_token=%s", senderID, s.pageAccessToken)
	req, err := http.NewRequest("GET", fbAPI, nil)
	if err != nil {
		// Assume CET.
		return defaultUser
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		s.log.Log("msg", "user details get error to FB API", "err", err)
		// Assume CET.
		return defaultUser
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.log.Log("msg", "can't read body", "err", err)
		return defaultUser
	}

	user := &models.User{}

	if err := json.Unmarshal(body, user); err != nil {
		s.log.Log("msg", "body unmarshall failed", "err", err)
		return defaultUser
	}
	return user
}

func MakeTask(senderID string, log log.Logger, db *leveldb.DB, bsvc bible.Service, psvc poster.Service) func() {
	return func() {
		log.Log("msg", "sending message", "user_id", senderID)

		userData, err := GetUserData(senderID, db)
		if err != nil {
			log.Log("msg", "error while getting user data", "user_id", senderID, "err", err)
			return
		}

		verses, err := bsvc.GetDay(userData.CurrentDay)
		if err != nil {
			log.Log("msg", "error while getting verses", "user_id", senderID, "err", err)
			// Reset day to 0 and try again...
			userData.CurrentDay = 0
			verses, err = bsvc.GetDay(userData.CurrentDay)
			if err != nil {
				log.Log("msg", "error while getting verses for day 0", "user_id", senderID, "err", err)
				return
			}
		}

		userData.CurrentDay++
		err = PutUserData(userData, db)
		if err != nil {
			log.Log("msg", "error while saving user progress", "user_id", senderID, "err", err)
		}

		err = psvc.ProcessMessages(userData.SenderID, verses, "NON_PROMOTIONAL_SUBSCRIPTION", "MESSAGE_TAG", "SILENT_PUSH")
		if err != nil {
			log.Log("msg", "error while sending verses", "user_id", senderID, "err", err)
			return
		}
	}
}

func (s *service) Info(senderID string) string {
	userData, err := GetUserData(senderID, s.DB)
	if err != nil {
		return "Can't find your user in database, maybe you want to `start` your schedule."
	}
	return fmt.Sprintf(
		"You have bible read plan scheduled at %s, currently you are at day %d",
		userData.ScheduleTime.Format("15:04"),
		userData.CurrentDay,
	)
}

func (s *service) ShowDay(message string) []string {
	shouldBeNumber := strings.TrimSpace(strings.TrimPrefix(message, showDayCommand))

	day, err := strconv.Atoi(shouldBeNumber)
	if err != nil {
		return []string{err.Error()}
	}

	verses, err := s.bsvc.GetDay(day)
	if err != nil {
		s.log.Log("msg", "show day error", "err", err)
		return []string{"Sorry! Something gone wrong, can't find day to show."}
	}
	return verses
}

func GetUserData(senderID string, db *leveldb.DB) (*User, error) {
	data, err := db.Get([]byte(senderID), nil)
	if err != nil {
		return nil, err
	}
	var userData User
	err = Unmarshal(data, &userData)
	if err != nil {
		return nil, err
	}
	return &userData, nil
}

func PutUserData(userData *User, db *leveldb.DB) error {
	data, err := Marshal(userData)
	if err != nil {
		return fmt.Errorf("failed to marshal data %s", err.Error())
	}
	err = db.Put([]byte(userData.SenderID), data, nil)
	if err != nil {
		return fmt.Errorf("failed to save your subscription %s", err.Error())
	}
	return nil
}

func (s *service) SetDay(message string, senderID string) string {
	userData, err := GetUserData(senderID, s.DB)
	if err != nil {
		return "Can't find your user in database, maybe you want to `start` your schedule."
	}

	shouldBeNumber := strings.TrimSpace(strings.TrimPrefix(message, setDayCommand))

	day, err := strconv.Atoi(shouldBeNumber)
	if err != nil {
		return err.Error()
	}

	userData.CurrentDay = day

	err = PutUserData(userData, s.DB)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("New schedule is set at day: %d. At %s", userData.CurrentDay, userData.ScheduleTime.Format("15:04"))
}

func (s *service) SetTime(msg string, senderID string) string {
	userData, err := GetUserData(senderID, s.DB)
	if err != nil {
		return "Can't find your user in database, maybe you want to `start` your schedule."
	}

	newTime, err := parseSetTimeCommand(msg)
	if err != nil {
		return err.Error()
	}

	userData.ScheduleTime = newTime

	s.schLock.RLock()
	currentSched := s.Schedulers[senderID]
	currentSched.Kill()
	s.schLock.RUnlock()

	err = s.AddScheduler(userData)
	if err != nil {
		return err.Error()
	}

	err = PutUserData(userData, s.DB)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("New schedule is set at: %s", newTime.Format("15:04"))
}

func parseSetTimeCommand(msg string) (time.Time, error) {
	timeString := strings.TrimPrefix(msg, setTimeCommand)
	timeString = strings.Trim(timeString, " ;[]{}'.,/\\|?")
	t, err := time.Parse("15:04", timeString)
	if err != nil {
		return time.Time{}, fmt.Errorf("can't parse time: %s", err)
	}
	return t, nil
}

func (s *service) AddScheduler(userData *User) error {

	task := MakeTask(userData.SenderID, s.log, s.DB, s.bsvc, s.psvc)

	sched, err := NewSchedulerTask(userData.SenderID, s.log, userData.ScheduleTime, task)
	if err != nil {
		return fmt.Errorf("error while creating scheduler %s", err.Error())
	}
	s.schLock.Lock()
	s.Schedulers[userData.SenderID] = sched
	s.schLock.Unlock()
	return nil
}

// Recover from down time...
func (s *service) Recover() error {
	iter := s.DB.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()
		var userData User
		err := Unmarshal(val, &userData)
		if err != nil {
			s.log.Log("msg", "failed to unmarshall", "key", key, "err", err)
			continue
		}
		err = s.AddScheduler(&userData)
		if err != nil {
			s.log.Log("msg", "failed to add scheduler", "key", key, "err", err)
			continue
		}
		s.log.Log("msg", "revovered user", "senderID", userData.SenderID, "scheduled_at", userData.ScheduleTime)
	}
	return nil
}

func Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func Unmarshal(b []byte, v interface{}) error {
	return msgpack.Unmarshal(b, v)
}

func NewSchedulerTask(senderID string, log log.Logger, time time.Time, task func()) (*SchedulerTask, error) {
	s := SchedulerTask{
		SenderID: senderID,
		Time:     time,
		Task:     task,
		log:      log,
	}

	err := s.Reload()
	if err != nil {
		return nil, err
	}

	return &s, nil
}

type SchedulerTask struct {
	SenderID string
	Time     time.Time
	Task     func()
	CronJob  *gocron.Scheduler
	done     chan struct{}
	log      log.Logger
}

func (s *SchedulerTask) Reload() error {
	// Kill old cron...
	s.Kill()
	s.CronJob = nil

	// Recreate done channel...
	s.done = make(chan struct{})

	// Recreate cron job...
	s.CronJob = gocron.NewScheduler()
	// Assume daily display...
	timeString := s.Time.Format("15:04")
	if s.Task != nil {
		s.log.Log("msg", "scheduling event", "time", timeString)
		s.CronJob.Every(1).Day().At(timeString).Do(s.Task)
	} else {
		return fmt.Errorf("missing task function")
	}

	go s.run()

	return nil
}

func (s *SchedulerTask) Kill() {
	if s.done != nil {
		close(s.done)
	}
}

func (s *SchedulerTask) run() {
	for {
		select {
		case <-s.CronJob.Start():
			s.CronJob.Clear()
			s.log.Log("msg", "cron exited unexpectedly", "sender_id", s.SenderID)
		case <-s.done:
			s.CronJob.Clear()
			return
		}
	}
}

type User struct {
	_msgpack     struct{} `msgpack:",omitempty"`
	SenderID     string
	ScheduleTime time.Time
	CurrentDay   int

	Name      string
	FirstName string
	LastName  string
	Timezone  int
}
