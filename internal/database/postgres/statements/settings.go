package statements

const (
	GetGatekeeperSettings = `
		SELECT * FROM gatekeeper_settings ORDER BY id DESC LIMIT 1;
	`

	UpdateGatekeeperSettings = `
		INSERT INTO gatekeeper_settings (
			filter_chat, 
			filter_links, 
			ignore_mods, 
			ignore_subs, 
			symbols_max, 
			emotes_max, 
			bad_words, 
			set
		) VALUES (
			$1, 
			$2, 
			$3, 
			$4, 
			$5, 
			$6, 
			$7, 
			$8
		) RETURNING id;
	`
)
