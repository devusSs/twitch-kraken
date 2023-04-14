package postgres

import (
	"github.com/devusSs/twitch-kraken/internal/database"
	"github.com/devusSs/twitch-kraken/internal/database/postgres/statements"
)

func (p *psql) LoadGateKeeperSettings() (database.GateKeeperSettings, error) {
	row := p.db.QueryRow(statements.GetGatekeeperSettings)

	var s database.GateKeeperSettings

	err := row.Scan(&s.ID, &s.FilterChat, &s.FilterLinks,
		&s.IgnoreMods, &s.IgnoreSubs, &s.SymbolsMax, &s.EmotesMax, &s.BadWords,
		&s.SetTime)

	return s, err
}

func (p *psql) UpdateGateKeeperSettings(settings database.GateKeeperSettings) error {
	row := p.db.QueryRow(statements.UpdateGatekeeperSettings, settings.FilterChat,
		settings.FilterLinks, settings.IgnoreMods, settings.IgnoreSubs,
		settings.SymbolsMax, settings.EmotesMax, settings.BadWords,
		settings.SetTime)

	err := row.Scan(&settings.ID)

	return err
}
