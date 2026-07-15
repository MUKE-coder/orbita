package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/orbita-sh/orbita/internal/models"
)

// DumpDatabase produces a logical dump of a managed database by exec-ing the
// engine's dump tool inside its running container. Returns the raw
// (uncompressed) dump bytes.
func (o *Orchestrator) DumpDatabase(ctx context.Context, mdb *models.ManagedDatabase, connString string) ([]byte, error) {
	containerID, err := o.containerForDB(ctx, mdb)
	if err != nil {
		return nil, err
	}

	user, pass := credentialsFromConnString(mdb.Engine, mdb.Name, connString)

	var cmd, env []string
	switch mdb.Engine {
	case models.EnginePostgres:
		// --clean --if-exists makes the dump drop objects on restore
		cmd = []string{"pg_dump", "--clean", "--if-exists", "-U", user, mdb.Name}
	case models.EngineMySQL, models.EngineMariaDB:
		cmd = []string{"sh", "-c", "exec mysqldump -u root " + shellQuote(mdb.Name)}
		env = []string{"MYSQL_PWD=" + pass}
	case models.EngineMongoDB:
		cmd = []string{"mongodump", "--archive", "-u", user, "-p", pass, "--authenticationDatabase", "admin"}
	case models.EngineRedis:
		cmd = []string{"sh", "-c", "redis-cli SAVE > /dev/null && cat /data/dump.rdb"}
	default:
		return nil, fmt.Errorf("DumpDatabase: unsupported engine %s", mdb.Engine)
	}

	res, err := o.dockerClient.ExecInContainer(ctx, containerID, cmd, env, nil)
	if err != nil {
		return nil, fmt.Errorf("DumpDatabase: %w", err)
	}
	if res.ExitCode != 0 {
		return nil, fmt.Errorf("DumpDatabase: dump tool exited %d: %s", res.ExitCode, truncate(string(res.Stderr), 500))
	}
	return res.Stdout, nil
}

// RestoreDatabase loads a previously produced dump back into the database.
func (o *Orchestrator) RestoreDatabase(ctx context.Context, mdb *models.ManagedDatabase, connString string, dump []byte) error {
	containerID, err := o.containerForDB(ctx, mdb)
	if err != nil {
		return err
	}

	user, pass := credentialsFromConnString(mdb.Engine, mdb.Name, connString)

	var cmd, env []string
	switch mdb.Engine {
	case models.EnginePostgres:
		cmd = []string{"psql", "-U", user, "-d", mdb.Name, "-v", "ON_ERROR_STOP=1"}
	case models.EngineMySQL, models.EngineMariaDB:
		cmd = []string{"sh", "-c", "exec mysql -u root " + shellQuote(mdb.Name)}
		env = []string{"MYSQL_PWD=" + pass}
	case models.EngineMongoDB:
		cmd = []string{"mongorestore", "--archive", "--drop", "-u", user, "-p", pass, "--authenticationDatabase", "admin"}
	case models.EngineRedis:
		// Write the RDB file and bounce the service so Redis loads it at boot.
		cmd = []string{"sh", "-c", "cat > /data/dump.rdb"}
	default:
		return fmt.Errorf("RestoreDatabase: unsupported engine %s", mdb.Engine)
	}

	res, err := o.dockerClient.ExecInContainer(ctx, containerID, cmd, env, bytes.NewReader(dump))
	if err != nil {
		return fmt.Errorf("RestoreDatabase: %w", err)
	}
	if res.ExitCode != 0 {
		return fmt.Errorf("RestoreDatabase: restore tool exited %d: %s", res.ExitCode, truncate(string(res.Stderr), 500))
	}

	if mdb.Engine == models.EngineRedis {
		if err := o.RestartDatabase(ctx, mdb); err != nil {
			return fmt.Errorf("RestoreDatabase: redis restart: %w", err)
		}
	}
	return nil
}

func (o *Orchestrator) containerForDB(ctx context.Context, mdb *models.ManagedDatabase) (string, error) {
	if mdb.DockerServiceID == nil || *mdb.DockerServiceID == "" {
		return "", fmt.Errorf("database %s has no running service", mdb.Name)
	}
	containerID, err := o.dockerClient.FindContainerIDForService(ctx, *mdb.DockerServiceID)
	if err != nil {
		return "", fmt.Errorf("database %s: %w", mdb.Name, err)
	}
	return containerID, nil
}

// credentialsFromConnString extracts user/password from the stored connection
// string. Falls back to the db name as user (how provisioning creates them).
func credentialsFromConnString(engine, dbName, connString string) (user, pass string) {
	user = dbName
	switch engine {
	case models.EnginePostgres, models.EngineMongoDB:
		if u, err := url.Parse(connString); err == nil && u.User != nil {
			user = u.User.Username()
			pass, _ = u.User.Password()
		}
	case models.EngineMySQL, models.EngineMariaDB:
		// user:pass@tcp(host:port)/db
		if at := strings.Index(connString, "@"); at > 0 {
			creds := connString[:at]
			if colon := strings.Index(creds, ":"); colon > 0 {
				user = creds[:colon]
				pass = creds[colon+1:]
			}
		}
	}
	return user, pass
}

// shellQuote single-quotes a value for safe interpolation into sh -c.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
