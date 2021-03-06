package room

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/shiwano/submarine/server/battle/lib/typhenapi"
	api "github.com/shiwano/submarine/server/battle/lib/typhenapi/type/submarine"
	battleAPI "github.com/shiwano/submarine/server/battle/lib/typhenapi/type/submarine/battle"
	webAPI "github.com/shiwano/submarine/server/battle/lib/typhenapi/web/submarine"
	"github.com/shiwano/submarine/server/battle/src/battle"
	"github.com/shiwano/submarine/server/battle/src/logger"
	"github.com/shiwano/submarine/server/battle/src/resource"
	"github.com/shiwano/submarine/server/battle/src/session"
)

type createSessionRequest struct {
	roomMember *battleAPI.RoomMember
	w          http.ResponseWriter
	r          *http.Request
	done       chan struct{}
}

type messageWithSession struct {
	message typhenapi.Type
	session *session.Session
}

// Room represents a room of the battle server.
type Room struct {
	id                  int64
	webAPI              *webAPI.WebAPI
	info                *battleAPI.PlayableRoom
	sessions            map[int64]*session.Session
	bots                map[int64]*api.Bot
	battle              *battle.Battle
	lastCreatedBotID    int64
	joinRequested       chan createSessionRequest
	leaveRequested      chan *session.Session
	roomMessageReceived chan messageWithSession
	closed              chan struct{}
}

func openRoom(ctx context.Context, webAPI *webAPI.WebAPI, id int64) (*Room, error) {
	// TODO: Validate whether the battle server can create the room.
	res, err := webAPI.Battle.FindRoom(id)
	if err != nil {
		return nil, err
	}
	if res.Room == nil {
		return nil, fmt.Errorf("No room(%v) found", id)
	}

	// TODO: Load the specified stage mesh and the light map.
	stageMesh, err := resource.Loader.LoadMesh(1)
	if err != nil {
		return nil, err
	}
	lightMap, err := resource.Loader.LoadLightMap(1)
	if err != nil {
		return nil, err
	}

	r := &Room{
		id:                  id,
		webAPI:              webAPI,
		info:                res.Room,
		sessions:            make(map[int64]*session.Session),
		bots:                make(map[int64]*api.Bot),
		battle:              battle.New(ctx, time.Second*300, stageMesh, lightMap),
		joinRequested:       make(chan createSessionRequest),
		leaveRequested:      make(chan *session.Session),
		roomMessageReceived: make(chan messageWithSession),
		closed:              make(chan struct{}),
	}
	go r.run(ctx)
	return r, nil
}

func (r *Room) String() string {
	return fmt.Sprintf("Room(%v)", r.id)
}

// Join creates a new session and join to the room.
func (r *Room) Join(roomMember *battleAPI.RoomMember, w http.ResponseWriter, hr *http.Request) bool {
	req := createSessionRequest{roomMember, w, hr, make(chan struct{})}
	select {
	case <-r.closed:
		return false
	case r.joinRequested <- req:
		<-req.done
		return true
	}
}

func (r *Room) run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer close(r.closed)

	logger.Log.Infof("%v opened", r)

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-r.battle.Finished():
			r.closeBattle()
			break loop
		case req := <-r.joinRequested:
			r.join(ctx, req)
		case session := <-r.leaveRequested:
			r.leave(session)
		case m := <-r.roomMessageReceived:
			switch t := m.message.(type) {
			case *battleAPI.StartRequest:
				r.startBattle(m.session)
			case *battleAPI.AddBotRequest:
				r.addBot()
			case *battleAPI.RemoveBotRequest:
				r.removeBot(t.BotId)
			}
		case output := <-r.battle.Gateway.Output:
			r.sendBattleMessageToSessions(output)
		}
	}

	cancel()
	for _, s := range r.sessions {
		<-s.Closed()
	}
	logger.Log.Infof("%v closed", r)
}

func (r *Room) toRoomAPIType() *battleAPI.Room {
	members := make([]*api.User, len(r.sessions))
	i := 0
	for _, s := range r.sessions {
		members[i] = s.ToUserAPIType()
		i++
	}
	bots := make([]*api.Bot, len(r.bots))
	i = 0
	for _, b := range r.bots {
		bots[i] = b
		i++
	}
	return &battleAPI.Room{Id: r.id, Members: members, Bots: bots}
}

func (r *Room) broadcastRoom() {
	message := r.toRoomAPIType()
	for _, s := range r.sessions {
		s.SendRoom(message)
	}
}

func (r *Room) startBattle(s *session.Session) {
	// TODO: Validate that can the session starts the battle.
	if r.battle.Start() {
		logger.Log.Infof("%v's battle started", r)
	}
}

func (r *Room) addBot() {
	r.lastCreatedBotID--
	bot := &api.Bot{Id: r.lastCreatedBotID, Name: "BOT"}
	if r.battle.EnterBot(bot) {
		r.bots[bot.Id] = bot
		r.broadcastRoom()
		logger.Log.Infof("Bot(%v) is added to %v", bot.Id, r)
	}
}

func (r *Room) removeBot(botID int64) {
	if bot, ok := r.bots[botID]; ok {
		if r.battle.LeaveBot(bot) {
			delete(r.bots, bot.Id)
			r.broadcastRoom()
			logger.Log.Infof("Bot(%v) is removed from %v", bot.Id, r)
		}
	}
}

func (r *Room) join(ctx context.Context, req createSessionRequest) {
	defer close(req.done)

	s, err := session.Open(ctx, req.roomMember, req.w, req.r)
	if err != nil {
		logger.Log.Errorf("Failed to create a session: %v", err)
		return
	}
	r.sessions[s.ID()] = s
	r.battle.EnterUser(s.ID())
	s.SynchronizeTime()
	r.broadcastRoom()
	logger.Log.Infof("%v joined into %v", s, r)

	go func() {
	loop:
		for {
			select {
			case <-s.Closed():
				break loop
			case m := <-s.RoomMessageReceived():
				r.roomMessageReceived <- messageWithSession{m, s}
			case m := <-s.BattleMessageReceived():
				r.battle.Gateway.InputMessage(s.ID(), m)
			}
		}
		select {
		case <-r.closed:
		case r.leaveRequested <- s:
			return
		}
	}()
}

func (r *Room) leave(s *session.Session) {
	logger.Log.Infof("%v leaved from %v", s, r)
	delete(r.sessions, s.ID())
	r.battle.LeaveUser(s.ID())
	r.broadcastRoom()
}

func (r *Room) closeBattle() {
	for c := 1; true; c++ {
		if _, err := r.webAPI.Battle.CloseRoom(r.id); err != nil {
			logger.Log.Errorf("%v failed %v times to use closeRoom API: %v", r, c, err)
			time.Sleep(time.Duration(c) * time.Second)
			continue
		}
		break
	}
}

func (r *Room) sendBattleMessageToSessions(output *battle.GatewayOutput) {
	if output.UserIDs == nil {
		for _, s := range r.sessions {
			s.SendBattleMessage(output.Message)
		}
	} else {
		for _, s := range r.sessions {
			for _, userID := range output.UserIDs {
				if s.ID() == userID {
					s.SendBattleMessage(output.Message)
				}
			}
		}
	}
}
