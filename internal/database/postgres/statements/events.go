package statements

const (
	AddEvent = `
		INSERT INTO auth_events (event_type, event_data, event_time) 
		VALUES ($1, $2, $3) RETURNING id;
	`
)
