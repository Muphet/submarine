package server

import (
	"context"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/shiwano/submarine/server/battle/lib/typhenapi"
	webapi "github.com/shiwano/submarine/server/battle/lib/typhenapi/web/submarine"
	websocketapi "github.com/shiwano/submarine/server/battle/lib/typhenapi/websocket/submarine"
	"github.com/shiwano/submarine/server/battle/src/logger"
	"github.com/shiwano/submarine/server/battle/test"
	conn "github.com/shiwano/websocket-conn"
)

type clientSession struct {
	api          *websocketapi.WebSocketAPI
	disconnected chan struct{}
	conn         *conn.Conn
	cancel       context.CancelFunc
}

func newClientSession() *clientSession {
	serializer := new(typhenapi.MessagePackSerializer)
	session := &clientSession{
		disconnected: make(chan struct{}, 1),
	}
	session.api = websocketapi.New(session, serializer, nil)
	return session
}

func (s *clientSession) Send(data []byte) error {
	return s.conn.SendBinaryMessage(data)
}

func (s *clientSession) connect(url string) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	c, _, err := conn.Connect(ctx, conn.DefaultSettings(), strings.Replace(url, "http", "ws", 1), nil)
	if err != nil {
		return err
	}
	s.conn = c
	go func() {
		for d := range s.conn.Stream() {
			if d.EOS {
				break
			}
			switch d.Message.MessageType {
			case conn.BinaryMessageType:
				if err := s.api.DispatchMessageEvent(d.Message.Data); err != nil {
					logger.Log.Error(err)
				}
			}
		}
		s.disconnected <- struct{}{}
	}()
	return nil
}

func (s *clientSession) close() {
	s.cancel()
}

func (s *clientSession) waitForDisconnected() {
	if s.disconnected != nil {
		<-s.disconnected
	}
}

func newWebAPIMock(url string) *webapi.WebAPI {
	WebAPIRoundTripper = test.NewWebAPITransporter()
	return newWebAPI(url)
}

func newTestServer() *httptest.Server {
	WebAPIRoundTripper = test.NewWebAPITransporter()
	gin.SetMode(gin.TestMode)
	s := httptest.NewServer(New())
	return s
}
