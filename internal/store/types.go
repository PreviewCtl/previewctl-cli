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

// PortMapping stores the host port assigned to a service in a preview environment.
type PortMapping struct {
	ID            string `db:"id"`
	PreviewEnvID  string `db:"preview_env_id"`
	ServiceName   string `db:"service_name"`
	ContainerPort int    `db:"container_port"`
	HostPort      int    `db:"host_port"`
	CreatedAt     int64  `db:"created_at"`
}
