// This file was generated by typhen-api

package configuration

import (
	"errors"
	"github.com/shiwano/submarine/server/battle/lib/typhenapi"
)

var _ = errors.New

// Server is a kind of TyphenAPI type.
type Server struct {
	ApiServerBaseUri    string                `json:"api_server_base_uri" msgpack:"api_server_base_uri"`
	BattleServerBaseUri string                `json:"battle_server_base_uri" msgpack:"battle_server_base_uri"`
	Database            *ServerDatabaseObject `json:"database" msgpack:"database"`
}

// Coerce the fields.
func (t *Server) Coerce() error {
	if t.Database == nil {
		return errors.New("Database should not be empty")
	}
	return nil
}

// Bytes creates the byte array.
func (t *Server) Bytes(serializer typhenapi.Serializer) ([]byte, error) {
	if err := t.Coerce(); err != nil {
		return nil, err
	}

	data, err := serializer.Serialize(t)
	if err != nil {
		return nil, err
	}

	return data, nil
}
