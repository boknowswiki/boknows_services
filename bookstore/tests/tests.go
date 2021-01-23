package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Container tracks information about a docker container started for tests.
type Container struct {
	ID   string
	Host string // IP:Port
}

// StartContainer runs a postgres container to execute commands.
func StartContainer(t *testing.T) *Container {
	t.Helper()

	cmd := exec.Command("docker", "run", "-P", "-d", "mongo")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not start container: %v", err)
	}

	id := out.String()[:12]
	t.Log("DB ContainerID:", id)

	cmd = exec.Command("docker", "inspect", id)
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not inspect container %s: %v", id, err)
	}

	var doc []struct {
		NetworkSettings struct {
			Ports struct {
				TCP27017 []struct {
					HostIP   string `json:"HostIp"`
					HostPort string `json:"HostPort"`
				} `json:"27017/tcp"`
			} `json:"Ports"`
		} `json:"NetworkSettings"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("could not decode json: %v", err)
	}

	network := doc[0].NetworkSettings.Ports.TCP27017[0]

	c := Container{
		ID:   id,
		Host: network.HostIP + ":" + network.HostPort,
	}

	t.Log("DB Host:", c.Host)

	return &c
}

// StopContainer stops and removes the specified container.
func StopContainer(t *testing.T, c *Container) {
	t.Helper()

	if err := exec.Command("docker", "stop", c.ID).Run(); err != nil {
		t.Fatalf("could not stop container: %v", err)
	}
	t.Log("Stopped:", c.ID)

	if err := exec.Command("docker", "rm", c.ID, "-v").Run(); err != nil {
		t.Fatalf("could not remove container: %v", err)
	}
	t.Log("Removed:", c.ID)
}

// DumpContainerLogs runs "docker logs" against the container and send it to t.Log
func DumpContainerLogs(t *testing.T, c *Container) {
	t.Helper()

	out, err := exec.Command("docker", "logs", c.ID).CombinedOutput()
	if err != nil {
		t.Fatalf("could not log container: %v", err)
	}
	t.Logf("Logs for %s\n%s:", c.ID, out)
}

// NewUnit creates a test database inside a Docker container. It creates the
// required table structure but the database is otherwise empty.
//
// It does not return errors as this intended for testing only. Instead it will
// call Fatal on the provided testing.T if anything goes wrong.
//
// It returns the database to use as well as a function to call at the end of
// the test.
func NewUnit(t *testing.T) (*mongo.Client, func()) {
	t.Helper()

	c := StartContainer(t)

	var url string

	url = fmt.Sprintf("mongodb://%s", c.Host)

	log.Printf("getting new mongo client with url %v", url)
	mclient, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		log.Fatalf("error: failed to create client: %s", err)
	}

	log.Println("got new mongo client")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = mclient.Connect(ctx)

	if err != nil {
		log.Fatal("error: failed to connect to mongodb docker image %v", err)
	}
	/*
		defer func() {
			if err = mclient.Disconnect(ctx); err != nil {
				panic(err)
			}
		}()
	*/
	log.Println("client connet")

	t.Log("waiting for database to be ready")

	// Wait for the database to be ready. Wait 100ms longer between each attempt.
	// Do not try more than 20 times.
	var pingError error
	maxAttempts := 20
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		pingError := mclient.Ping(ctx, nil)
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
	}

	if pingError != nil {
		DumpContainerLogs(t, c)
		StopContainer(t, c)
		t.Fatalf("waiting for database to be ready: %v", pingError)
	}

	/*
		if err := schema.Migrate(db); err != nil {
			databasetest.StopContainer(t, c)
			t.Fatalf("migrating: %s", err)
		}
	*/

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		t.Helper()
		mclient.Disconnect(ctx)
		StopContainer(t, c)
	}

	return mclient, teardown
}
