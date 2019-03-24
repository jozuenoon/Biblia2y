package messenger

import (
	"context"
	"reflect"

	"github.com/go-kit/kit/endpoint"
	m "github.com/jozuenoon/biblia2y/models"
)

func makeMessagesEndPoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		callback, ok := request.(m.Callback)
		if ok && callback.Object == "page" {
			for _, entry := range callback.Entry {
				for _, event := range entry.Messaging {
					if !reflect.DeepEqual(event.Message, m.Message{}) && event.Message.Text != "" {
						input := ParseMessageInput{
							Message:   event.Message.Text,
							SenderId:  event.Sender.ID,
							TimeStamp: event.Timestamp}
						s.ResponseSink() <- s.ParseMessage(&input)
					}
				}
			}
			return nil, nil
		}
		return "not supported", nil
	}
}
