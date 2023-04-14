package statements

const (
	AddCommand = `
		INSERT INTO twitch_commands (name, output, userlevel, cooldown, added, edited) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;
	`

	UpdateCommand = `
		UPDATE twitch_commands SET output = $1, userlevel = $2, cooldown = $3, edited = $4 
		WHERE name = $5 
		RETURNING id;
	`

	DeleteCommand = `
		DELETE FROM twitch_commands WHERE name = $1;
	`

	GetAllCommands = `
		SELECT * FROM twitch_commands;
	`

	GetCommand = `
		SELECT * FROM twitch_commands WHERE name = $1;
	`
)
