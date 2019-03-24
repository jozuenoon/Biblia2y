package messenger

import (
	"context"
	"encoding/json"
	"net/http"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	m "github.com/jozuenoon/biblia2y/models"
)

func MakeHandler(bs Service, logger kitlog.Logger, verifyToken string) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
	}

	r := mux.NewRouter()
	r.Methods("POST").Path("/webhook").Handler(
		handlers.ContentTypeHandler(kithttp.NewServer(
			makeMessagesEndPoint(bs),
			decodeMessage,
			encodeResponse,
			opts...), "application/json"))

	r.Methods("GET").Path("/webhook").HandlerFunc(makeVerificationEndPoint(verifyToken))
	return r
}

func decodeMessage(ctx context.Context, r *http.Request) (interface{}, error) {
	var callback m.Callback
	err := json.NewDecoder(r.Body).Decode(&callback)
	if err != nil {
		return nil, err
	}
	return callback, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if response == nil {
		w.WriteHeader(200)
		w.Write([]byte("Got your message"))
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Message not supported"))
	}
	return nil
}

func makeVerificationEndPoint(verifyToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		challenge := r.URL.Query().Get("hub.challenge")
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		if mode != "" && token == verifyToken {
			w.WriteHeader(200)
			w.Write([]byte(challenge))
		} else {
			w.WriteHeader(404)
			w.Write([]byte("Error, wrong validation token"))
		}
	}
}
