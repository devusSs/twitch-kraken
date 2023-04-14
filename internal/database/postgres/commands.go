package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/devusSs/twitch-kraken/internal/database"
	"github.com/devusSs/twitch-kraken/internal/database/postgres/statements"
	"github.com/devusSs/twitch-kraken/internal/logging"
)

func (p *psql) AddTwitchCommand(command database.TwitchCommand) error {
	row := p.db.QueryRow(statements.AddCommand, command.Name, command.Output, command.Userlevel,
		command.Cooldown, command.Added, command.Edited)

	err := row.Scan(&command.ID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return sql.ErrTxDone
		}
		return err
	}

	return nil
}

func (p *psql) GetAllTwitchCommands() ([]database.TwitchCommand, error) {
	rows, err := p.db.Query(statements.GetAllCommands)
	if err != nil {
		return nil, err
	}

	commands := []database.TwitchCommand{}

	for rows.Next() {
		c := database.TwitchCommand{}

		if err := rows.Scan(&c.ID, &c.Name, &c.Output, &c.Userlevel,
			&c.Cooldown, &c.Added, &c.Edited); err != nil {
			return nil, err
		}

		commands = append(commands, c)
	}

	return commands, nil
}

func (p *psql) GetOneTwitchCommand(name string) (database.TwitchCommand, error) {
	row := p.db.QueryRow(statements.GetCommand, name)

	var c database.TwitchCommand

	err := row.Scan(&c.ID, &c.Name, &c.Output, &c.Userlevel, &c.Cooldown, &c.Added, &c.Edited)

	return c, err
}

func (p *psql) UpdateTwitchCommand(command database.TwitchCommand) (database.TwitchCommand, error) {
	row := p.db.QueryRow(statements.UpdateCommand, command.Output, command.Userlevel, command.Cooldown,
		command.Edited.Time, command.Name)

	err := row.Scan(&command.ID)

	return command, err
}

func (p *psql) DeleteTwitchCommand(name string) error {
	res, err := p.db.Exec(statements.DeleteCommand, name)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		logging.WriteError(err)
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	if aff > 0 && aff != 1 {
		return fmt.Errorf("error: multiple rows affected (%d)", aff)
	}
	return nil
}
