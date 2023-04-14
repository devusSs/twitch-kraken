package postgres

import (
	"database/sql"
	"time"

	"github.com/devusSs/twitch-kraken/internal/database"
	"github.com/devusSs/twitch-kraken/internal/database/postgres/statements"
)

func (p *psql) RegisterTwitchUser(username string, firstSeen time.Time) error {
	row := p.db.QueryRow(statements.RegisterTwitchUser, username, firstSeen, firstSeen)

	var id int

	err := row.Scan(&id)

	// Indicates successfull insert since we do not return anything there yet.
	if err == sql.ErrNoRows {
		return nil
	}

	return err
}

func (p *psql) UpdateTwitchUserDC(username string, lastSeen time.Time) error {
	row := p.db.QueryRow(statements.UpsertTwitchUserDC, username, lastSeen, lastSeen)

	var id int

	err := row.Scan(&id)

	// Indicates successfull insert since we do not return anything there yet.
	if err == sql.ErrNoRows {
		return nil
	}

	return err
}

func (p *psql) UpdateTwitchUserBaseDetails(user database.TwitchUser) error {
	row := p.db.QueryRow(statements.UpsertTwitchUserBaseDetails, user.Username, user.TwitchID, user.DisplayName,
		user.IsMod.Bool, user.LastSeen.Time, user.LastSeen.Time)

	err := row.Scan(&user.ID)

	// Indicates successfull insert since we do not return anything there yet.
	if err == sql.ErrNoRows {
		return nil
	}

	return err
}

func (p *psql) UpdateTwitchUserOnBan(user database.TwitchUser) error {
	row := p.db.QueryRow(statements.UpsertTwitchUserBanOrTimeout, user.TwitchID, user.Username,
		user.LastSeen.Time, user.HasBeenBanned.Bool, user.LastBan.Time)

	err := row.Scan(&user.ID)

	// Indicates successfull insert since we do not return anything there yet.
	if err == sql.ErrNoRows {
		return nil
	}

	return err
}
