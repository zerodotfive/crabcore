package osrelease

import (
	"bufio"
	"os"
	"strings"
)

type OSRelease struct {
	ID              string
	Name            string
	Version         string
	VersionID       string
	VersionCodename string
	PrettyName      string
}

func GetOSRelease() (*OSRelease, error) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rel := &OSRelease{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := strings.Trim(parts[1], `"`)

		switch key {
		case "ID":
			rel.ID = value
		case "NAME":
			rel.Name = value
		case "VERSION":
			rel.Version = value
		case "VERSION_ID":
			rel.VersionID = value
		case "UBUNTU_CODENAME":
			rel.VersionCodename = value
		case "PRETTY_NAME":
			rel.PrettyName = value
		}
	}
	return rel, scanner.Err()
}
