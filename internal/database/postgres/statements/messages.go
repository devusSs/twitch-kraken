package statements

const (
	AddMessage = `
		INSERT INTO message_events (
			issuer,
			content,
			sent
		) VALUES (
			$1,
			$2,
			$3
		) RETURNING id;
	`
)
