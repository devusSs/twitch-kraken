package statements

const (
	CreateGateKeeperSettingsStore = `
		CREATE TABLE IF NOT EXISTS gatekeeper_settings (
			id bigserial, 
			filter_chat boolean DEFAULT TRUE, 
			filter_links boolean DEFAULT TRUE, 
			ignore_mods boolean DEFAULT TRUE, 
			ignore_subs boolean DEFAULT FALSE, 
			symbols_max integer DEFAULT 5, 
			emotes_max integer DEFAULT 3, 
			bad_words text[],
			set timestamp
		);
	`

	CreateUsersTable = `
		CREATE TABLE IF NOT EXISTS twitch_users (
			id bigserial,
			twitchid text UNIQUE,
			twitchusername text NOT NULL UNIQUE,
			displayname text,
			ismod boolean,
			firstseen timestamp,
			lastseen timestamp,
			hasbeenbanned boolean DEFAULT FALSE,
			lastban timestamp
		);
	`

	CreateCommandsTable = `
		CREATE TABLE IF NOT EXISTS twitch_commands (
			id bigserial,
			name text NOT NULL UNIQUE,
			output text NOT NULL,
			userlevel integer NOT NULL,
			cooldown integer NOT NULL,
			added timestamp,
			edited timestamp
		);
	`

	CreateAuthEventsTable = `
		CREATE TABLE IF NOT EXISTS auth_events (
			id bigserial,
			event_type text NOT NULL,
			event_data text NOT NULL,
			event_time timestamp NOT NULL
		);
	`

	CreateMessageEventsTable = `
		CREATE TABLE IF NOT EXISTS message_events (
			id bigserial,
			issuer text NOT NULL,
			content text NOT NULL,
			sent timestamp NOT NULL
		);
	`
)
