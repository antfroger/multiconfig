package multiconfig

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

type (
	Server struct {
		Name       string `required:"true"`
		Port       int    `default:"6060"`
		ID         int64
		Labels     []int
		Enabled    bool
		Users      []string
		Postgres   Postgres
		unexported string
		Interval   time.Duration
	}

	// Postgres holds Postgresql database related configuration
	Postgres struct {
		Enabled           bool
		Port              int      `required:"true" customRequired:"yes"`
		Hosts             []string `required:"true"`
		DBName            string   `default:"configdb"`
		AvailabilityRatio float64
		unexported        string
	}

	TaggedServer struct {
		Name     string `required:"true"`
		Postgres `structs:",flatten"`
	}

	FlattenedServer struct {
		Postgres Postgres
	}

	CamelCaseServer struct {
		AccessKey         string
		Normal            string
		DBName            string `default:"configdb"`
		AvailabilityRatio float64
	}

	// App holds services including an API, a Database and a Server
	App struct {
		API      API
		Service1 AppServer
		Service2 AppServer
		Mongo    Database
	}

	AppServer struct {
		Scheme   string `default:"https"`
		Host     string
		Port     int
		Username string
		Password string
	}

	API struct {
		AppServer
		Test bool
	}

	Database struct {
		AppServer
		DBName string
	}
)

var (
	testTOML = "testdata/config.toml"
	testJSON = "testdata/config.json"
	testYAML = "testdata/config.yaml"
)

func getDefaultServer() *Server {
	return &Server{
		Name:     "koding",
		Port:     6060,
		Enabled:  true,
		ID:       1234567890,
		Labels:   []int{123, 456},
		Users:    []string{"ankara", "istanbul"},
		Interval: 10 * time.Second,
		Postgres: Postgres{
			Enabled:           true,
			Port:              5432,
			Hosts:             []string{"192.168.2.1", "192.168.2.2", "192.168.2.3"},
			DBName:            "configdb",
			AvailabilityRatio: 8.23,
		},
	}
}

func getDefaultCamelCaseServer() *CamelCaseServer {
	return &CamelCaseServer{
		AccessKey:         "123456",
		Normal:            "normal",
		DBName:            "configdb",
		AvailabilityRatio: 8.23,
	}
}

func getDefaultApp() *App {
	return &App{
		API: API{
			AppServer: AppServer{
				Scheme: "https",
				Host:   "api.myapp.com",
				Port:   81,
			},
			Test: false,
		},
		Service1: AppServer{
			Scheme: "https",
			Host:   "service1.myapp.com",
			Port:   82,
		},
		Service2: AppServer{
			Scheme: "https",
			Host:   "service2.myapp.com",
			Port:   83,
		},
		Mongo: Database{
			DBName: "myDatabase",
			AppServer: AppServer{
				Scheme:   "mongodb",
				Host:     "localhost",
				Port:     27017,
				Username: "admin",
				Password: "admin",
			},
		},
	}
}

func TestNewWithPath(t *testing.T) {
	var _ Loader = NewWithPath(testTOML)
}

func TestLoad(t *testing.T) {
	m := NewWithPath(testTOML)

	s := new(Server)
	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	testStruct(t, s, getDefaultServer())
}

func TestDefaultLoader(t *testing.T) {
	m := New()

	s := new(Server)
	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	if err := m.Validate(s); err != nil {
		t.Error(err)
	}
	testStruct(t, s, getDefaultServer())

	s.Name = ""
	if err := m.Validate(s); err == nil {
		t.Error("Name should be required")
	}
}

func testStruct(t *testing.T, s *Server, d *Server) {
	if s.Name != d.Name {
		t.Errorf("Name value is wrong: %s, want: %s", s.Name, d.Name)
	}

	if s.Port != d.Port {
		t.Errorf("Port value is wrong: %d, want: %d", s.Port, d.Port)
	}

	if s.Enabled != d.Enabled {
		t.Errorf("Enabled value is wrong: %t, want: %t", s.Enabled, d.Enabled)
	}

	if s.Interval != d.Interval {
		t.Errorf("Interval value is wrong: %v, want: %v", s.Interval, d.Interval)
	}

	if s.ID != d.ID {
		t.Errorf("ID value is wrong: %v, want: %v", s.ID, d.ID)
	}

	if len(s.Labels) != len(d.Labels) {
		t.Errorf("Labels value is wrong: %d, want: %d", len(s.Labels), len(d.Labels))
	} else {
		for i, label := range d.Labels {
			if s.Labels[i] != label {
				t.Errorf("Label is wrong for index: %d, label: %d, want: %d", i, s.Labels[i], label)
			}
		}
	}

	if len(s.Users) != len(d.Users) {
		t.Errorf("Users value is wrong: %d, want: %d", len(s.Users), len(d.Users))
	} else {
		for i, user := range d.Users {
			if s.Users[i] != user {
				t.Errorf("User is wrong for index: %d, user: %s, want: %s", i, s.Users[i], user)
			}
		}
	}

	testPostgres(t, s.Postgres, d.Postgres)
}

func testFlattenedStruct(t *testing.T, s *FlattenedServer, d *Server) {
	// Explicitly state that Enabled should be true, no need to check
	// `x == true` infact.
	testPostgres(t, s.Postgres, d.Postgres)
}

func testPostgres(t *testing.T, s Postgres, d Postgres) {
	if s.Enabled != d.Enabled {
		t.Errorf("Postgres enabled is wrong %t, want: %t", s.Enabled, d.Enabled)
	}

	if s.Port != d.Port {
		t.Errorf("Postgres Port value is wrong: %d, want: %d", s.Port, d.Port)
	}

	if s.DBName != d.DBName {
		t.Errorf("DBName is wrong: %s, want: %s", s.DBName, d.DBName)
	}

	if s.AvailabilityRatio != d.AvailabilityRatio {
		t.Errorf("AvailabilityRatio is wrong: %f, want: %f", s.AvailabilityRatio, d.AvailabilityRatio)
	}

	if len(s.Hosts) != len(d.Hosts) {
		// do not continue testing if this fails, because others is depending on this test
		t.Fatalf("Hosts len is wrong: %v, want: %v", s.Hosts, d.Hosts)
	}

	for i, host := range d.Hosts {
		if s.Hosts[i] != host {
			t.Fatalf("Hosts number %d is wrong: %v, want: %v", i, s.Hosts[i], host)
		}
	}
}

func testCamelcaseStruct(t *testing.T, s *CamelCaseServer, d *CamelCaseServer) {
	if s.AccessKey != d.AccessKey {
		t.Errorf("AccessKey is wrong: %s, want: %s", s.AccessKey, d.AccessKey)
	}

	if s.Normal != d.Normal {
		t.Errorf("Normal is wrong: %s, want: %s", s.Normal, d.Normal)
	}

	if s.DBName != d.DBName {
		t.Errorf("DBName is wrong: %s, want: %s", s.DBName, d.DBName)
	}

	if s.AvailabilityRatio != d.AvailabilityRatio {
		t.Errorf("AvailabilityRatio is wrong: %f, want: %f", s.AvailabilityRatio, d.AvailabilityRatio)
	}
}

func TestLoadApp(t *testing.T) {
	m := NewWithPath(testTOML)

	app := new(App)
	if err := m.Load(app); err != nil {
		t.Error(err)
	}

	opts := cmp.AllowUnexported(Server{}, Postgres{})
	if diff := cmp.Diff(getDefaultApp(), app, opts); diff != "" {
		t.Errorf("diff = %s", diff)
	}
}
