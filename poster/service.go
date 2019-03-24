package poster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/go-kit/kit/log"

	m "github.com/jozuenoon/biblia2y/models"
)

type Service interface {
	ProcessMessages(senderId string, messages []string, tag string, messagingType string, notificationType string) error
}

func New(PageAccessToken, FaceBookAPI string, logger log.Logger) Service {
	return &service{
		logger:          logger,
		PageAccessToken: PageAccessToken,
		FaceBookAPI:     FaceBookAPI,
	}
}

type service struct {
	PageAccessToken string
	FaceBookAPI     string
	logger          log.Logger
}

// MessageSplitter will divide long messages into smaller pieces,
// facebook API will accept only 2000 characters at once.
func messageSplitter(input string, msgLen int, r *regexp.Regexp) []string {
	if len(input) == 0 {
		return nil
	}
	if len(input) < msgLen {
		return []string{input}
	}

	idx := r.FindIndex([]byte(input[msgLen:]))
	// If can't find any delimiter...
	if len(idx) < 1 {
		return []string{input}
	}
	return append([]string{input[:msgLen+idx[0]+1]}, messageSplitter(input[msgLen+idx[0]+1:], msgLen, r)...)
}

func (p *service) ProcessMessages(senderId string, messages []string, tag string, messagingType string, notificationType string) error {
	client := &http.Client{}
	responses := make([]m.Response, 0, len(messages))

	r := regexp.MustCompile("([.])")

	for _, verses := range messages {
		for _, msg := range messageSplitter(verses, 1000, r) {
			response := m.Response{
				Recipient: m.User{
					ID: senderId,
				},
				Message: m.Message{
					Text: msg,
				},
				Tag:           tag,
				MessagingType: messagingType,
			}
			responses = append(responses, response)
		}
	}

	for _, response := range responses {
		body := new(bytes.Buffer)
		json.NewEncoder(body).Encode(&response)

		url := fmt.Sprintf(p.FaceBookAPI, p.PageAccessToken)
		req, err := http.NewRequest("POST", url, body)
		if err != nil {
			continue
		}
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			p.logger.Log("msg", "post error to FB API", "err", err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			respInfo, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				p.logger.Log("msg", "post error to FB API", "err", err)
				continue
			}
			p.logger.Log("msg", "post returned", "err", string(respInfo))

		}
		time.Sleep(2 * time.Second)
	}
	return nil
}
