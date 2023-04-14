package postgres

import (
	"github.com/devusSs/twitch-kraken/internal/database"
	"github.com/devusSs/twitch-kraken/internal/database/postgres/statements"
)

func (p *psql) AddMessageEvent(event database.MessageEvent) (database.MessageEvent, error) {
	row := p.db.QueryRow(statements.AddMessage, event.Issuer, event.Content, event.Sent)

	err := row.Scan(&event.ID)

	return event, err
}
