package barkup

import (
	"fmt"
	"os"
	"os/exec"
)

var (
	// PGDumpCmd is the path to the `pg_dump` executable
	PGDumpCmd    = "pg_dump"
	PGRestoreCmd = "pg_restore"
)

// Postgres is an `Exporter` interface that backs up a Postgres database via the `pg_dump` command
type Postgres struct {
	// DB Host (e.g. 127.0.0.1)
	Host string
	// DB Port (e.g. 5432)
	Port string
	// DB Name
	DB string
	// Connection Username
	Username string
	// Filename of the file to save
	Filename string
	// Extra pg_dump options
	// e.g []string{"--inserts"}
	Options []string
}

// Export produces a `pg_dump` of the specified database, and creates a gzip compressed tarball archive.
func (x Postgres) Export() *ExportResult {
	result := &ExportResult{MIME: "application/gzip"}
	result.Path = fmt.Sprintf(x.Filename)
	out, err := exec.Command(PGDumpCmd, "-Fc", "-U", x.Username, x.DB).Output()
	if err != nil {
		result.Error = makeErr(err, string(out))
		return result
	}

	result.Path = x.Filename
	_, err = exec.Command(GzipCmd, "-9", ">", result.Path).Output()
	if err != nil {
		result.Error = makeErr(err, string(out))
		return result
	}
	os.Remove(x.Filename)

	return result
}

func (x Postgres) Import() error {
	options := x.dumpOptions()
	command := fmt.Sprintf("gunzip < %s | %s %s", x.Filename, PGRestoreCmd, x.DB)
	_, err := exec.Command(command, options...).Output()
	if err != nil {
		return err
	}
	return nil
}

func (x Postgres) dumpOptions() []string {
	options := x.Options

	if x.DB != "" {
		options = append(options, fmt.Sprintf(`-d %v`, x.DB))
	}

	if x.Host != "" {
		options = append(options, fmt.Sprintf(`-h %v`, x.Host))
	}

	if x.Port != "" {
		options = append(options, fmt.Sprintf(`-p %v`, x.Port))
	}

	if x.Username != "" {
		options = append(options, fmt.Sprintf(`-U %v`, x.Username))
	}

	return options
}
