package compose

import (
	"io/ioutil"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/docker/api/types/strslice"
	units "github.com/docker/go-units"
	yaml "gopkg.in/yaml.v2"
)

// Compose informations from docker compose file
type Compose struct {
	Version  string              `yaml:"version,omitempty"`
	Services map[string]*Service `yaml:"services"`
}

// Service informations from docker compose services
type Service struct {
	Command     strslice.StrSlice `yaml:"command,flow,omitempty"`
	Deploy      Deploy            `yaml:"deploy,omitempty"`
	DNS         strslice.StrSlice `yaml:"dns,omitempty"`
	DNSSearch   strslice.StrSlice `yaml:"dns_search,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	SecurityOpt []string          `yaml:"security_opt,omitempty"`
	Entrypoint  strslice.StrSlice `yaml:"entrypoint,flow,omitempty"`
	// Environment []string `yaml:"environment,omitempty"`
	Essential  bool     `yaml:"essential,omitempty"`
	ExtraHosts []string `yaml:"extra_hosts,omitempty"`
	Hostname   string   `yaml:"hostname,omitempty"`
	Image      string   `yaml:"image,omitempty"`
	Ports      []string `yaml:"ports,omitempty"`
	Privileged bool     `yaml:"privileged,omitempty"`
	Tty        bool     `yaml:"tty,omitempty"`
	User       string   `yaml:"user,omitempty"`
	WorkingDir string   `yaml:"working_dir,omitempty"`
}

// Deploy informations from compose service deploy
type Deploy struct {
	Resources Resouces `yaml:"resources,omitempty"`
	Replicas  int64    `yaml:"replicas,omitempty"`
}

// Resouces informations from compose service deploy
type Resouces struct {
	Limits       Limits       `yaml:"limits,omitempty"`
	Reservations Reservations `yaml:"reservations,omitempty"`
}

// Limits informations from compose service deploy
type Limits struct {
	Cpus   string `yaml:"limits,omitempty"`
	Memory string `yaml:"reservations,omitempty"`
}

// Reservations informations from compose service deploy
type Reservations struct {
	Cpus   string `yaml:"limits,omitempty"`
	Memory string `yaml:"reservations,omitempty"`
}

// Log driver information from compose service
type Log struct {
	Driver  string            `yaml:"driver,omitempty"`
	Options map[string]string `yaml:"options,omitempty"`
}

// ToAWSContainerDefinitions convert Compose Services to slice of AWS ContainerDefinition
func (c *Compose) ToAWSContainerDefinitions() (cds []*ecs.ContainerDefinition, err error) {
	for name, service := range c.Services {
		var cd *ecs.ContainerDefinition
		cd, err = service.ToAWSContainerDefinition(name)
		if err != nil {
			return
		}

		cds = append(cds, cd)
	}
	return
}

// ToAWSContainerDefinition convert Services to slice of AWS ContainerDefinition
func (s *Service) ToAWSContainerDefinition(name string) (*ecs.ContainerDefinition, error) {
	var err error
	cd := &ecs.ContainerDefinition{}

	if len(s.Command) > 0 {
		cd.Command = aws.StringSlice(s.Command)
	}

	if cpu, err := strconv.ParseInt(s.Deploy.Resources.Reservations.Cpus, 10, 64); err == nil {
		cd.Cpu = aws.Int64(cpu)
	}

	if len(s.DNSSearch) > 0 {
		cd.DnsSearchDomains = aws.StringSlice(s.DNSSearch)
	}

	if len(s.DNS) > 0 {
		cd.DnsServers = aws.StringSlice(s.DNS)
	}

	if len(s.Labels) > 0 {
		cd.DockerLabels = aws.StringMap(s.Labels)
	}

	if len(s.SecurityOpt) > 0 {
		cd.DockerSecurityOptions = aws.StringSlice(s.SecurityOpt)
	}

	if len(s.Entrypoint) > 0 {
		cd.EntryPoint = aws.StringSlice(s.Entrypoint)
	}

	cd.Essential = aws.Bool(s.Essential)

	// cd.ExtraHosts
	// cd.HealthCheck

	if s.Hostname != "" {
		cd.Hostname = aws.String(s.Hostname)
	}

	cd.Image = aws.String(s.Image)

	// cd.Interactive = aws.Bool()
	// cd.Links
	// cd.LinuxParameters
	// cd.LogConfiguration

	if memory, err := units.RAMInBytes(s.Deploy.Resources.Limits.Memory); err == nil {
		cd.Memory = aws.Int64(memory / 1024 / 1024)
	}

	if memoryReservation, err := units.RAMInBytes(s.Deploy.Resources.Reservations.Memory); err == nil {
		cd.MemoryReservation = aws.Int64(memoryReservation / 1024 / 1024)
	}

	// cd.Environment
	// cd.MountPoints

	cd.Name = aws.String(name)

	// cd.PortMappings

	cd.Privileged = aws.Bool(s.Privileged)

	cd.PseudoTerminal = aws.Bool(s.Tty)

	// cd.ReadonlyRootFilesystem
	// cd.RepositoryCredentials
	// cd.SystemControls
	// cd.Ulimits

	if s.User != "" {
		cd.User = aws.String(s.User)
	}

	// cd.VolumesFrom

	if s.User != "" {
		cd.WorkingDirectory = aws.String(s.User)
	}

	return cd, err
}

// FromFile open informed file as Compose object
func FromFile(filepath string) (compose *Compose, err error) {
	fileData, err := ioutil.ReadFile(filepath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(fileData, &compose)

	return
}
