package postgres

import (
	"github.com/devusSs/twitch-kraken/internal/database"
	"github.com/devusSs/twitch-kraken/internal/database/postgres/statements"
)

func (p *psql) AddAuthEvent(event database.AuthEvent) (database.AuthEvent, error) {
	row := p.db.QueryRow(statements.AddEvent, event.Type, event.Data, event.Timestamp)

	err := row.Scan(&event.ID)

	return event, err
}
