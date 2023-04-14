package statements

const (
	RegisterTwitchUser = `
		INSERT INTO twitch_users (twitchusername, firstseen, lastseen) VALUES ($1, $2, $3) 
		ON CONFLICT (twitchusername) 
		DO UPDATE SET lastseen = $3 
		WHERE twitch_users.twitchusername = $1 
		RETURNING id;
	`

	UpsertTwitchUserDC = `
		INSERT INTO twitch_users (twitchusername, firstseen, lastseen) VALUES ($1, $2, $3) 
		ON CONFLICT (twitchusername) 
		DO UPDATE SET twitchusername = $1, firstseen = $2, lastseen = $3 
		WHERE twitch_users.twitchusername = $1 
		RETURNING id;
	`

	UpsertTwitchUserBaseDetails = `
		INSERT INTO twitch_users (twitchid, twitchusername, displayname, ismod, firstseen, lastseen) VALUES ($2, $1, $3, $4, $5, $6) 
		ON CONFLICT (twitchusername) 
		DO UPDATE SET twitchid = $2, displayname = $3, ismod = $4, lastseen = $5 
		WHERE twitch_users.twitchusername = $1 
		RETURNING id;
	`

	UpsertTwitchUserBanOrTimeout = `
		INSERT INTO twitch_users (twitchid, twitchusername, lastseen, hasbeenbanned, lastban) VALUES ($1, $2, $3, $4, $5) 
		ON CONFLICT (twitchusername) 
		DO UPDATE SET twitchid = $1, lastseen = $3, hasbeenbanned = $4, lastban = $5 
		WHERE twitch_users.twitchusername = $2 
		RETURNING id;
	`
)
