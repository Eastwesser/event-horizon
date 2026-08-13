package config

// ConfigProvider exposes typed config views for DI and tests.
type ConfigProvider interface {
	GRPCConfig() GRPCConfig
	MetricsConfig() MetricsConfig
	PostgresConfig() PostgresConfig
	RedisConfig() RedisConfig
	NATSConfig() NATSConfig
	LoggerConfig() LoggerConfig
}

type GRPCConfig interface {
	Port() string
}

type MetricsConfig interface {
	Port() string
}

type PostgresConfig interface {
	Host() string
	Port() string
	User() string
	Password() string
	Database() string
	DSN() string
}

type RedisConfig interface {
	Addr() string
	DB() int
}

type NATSConfig interface {
	URL() string
}

type LoggerConfig interface {
	Level() string
	Format() string
}

type grpcView struct{ port string }

func (g grpcView) Port() string { return g.port }

type metricsView struct{ port string }

func (m metricsView) Port() string { return m.port }

type postgresView struct {
	host, port, user, password, database string
}

func (p postgresView) Host() string     { return p.host }
func (p postgresView) Port() string     { return p.port }
func (p postgresView) User() string     { return p.user }
func (p postgresView) Password() string { return p.password }
func (p postgresView) Database() string { return p.database }
func (p postgresView) DSN() string {
	return "postgres://" + p.user + ":" + p.password + "@" + p.host + ":" + p.port + "/" + p.database
}

type redisView struct {
	addr string
	db   int
}

func (r redisView) Addr() string { return r.addr }
func (r redisView) DB() int      { return r.db }

type natsView struct{ url string }

func (n natsView) URL() string { return n.url }

type loggerView struct {
	level  string
	format string
}

func (l loggerView) Level() string  { return l.level }
func (l loggerView) Format() string { return l.format }

func (c *Config) GRPCConfig() GRPCConfig {
	return grpcView{port: c.GRPCPort}
}

func (c *Config) MetricsConfig() MetricsConfig {
	return metricsView{port: c.MetricsPort}
}

func (c *Config) PostgresConfig() PostgresConfig {
	return postgresView{
		host: c.DBHost, port: c.DBPort, user: c.DBUser,
		password: c.DBPassword, database: c.DBName,
	}
}

func (c *Config) RedisConfig() RedisConfig {
	return redisView{addr: c.RedisAddr, db: c.RedisDB}
}

func (c *Config) NATSConfig() NATSConfig {
	return natsView{url: c.NATSUrl}
}

func (c *Config) LoggerConfig() LoggerConfig {
	return loggerView{level: getEnv("LOG_LEVEL", "info"), format: getEnv("LOG_FORMAT", "text")}
}
