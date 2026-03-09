package store

// PreviewEnvironment represents an active preview environment.
type PreviewEnvironment struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Workspace string `db:"workspace"`
	Branch    string `db:"branch"`
	Status    string `db:"status"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}
