package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
)

// DockerClient is our wrapper for docker.Client
type DockerClient struct {
	*docker.Client
}

// NewDockerClient based on options and env
func NewDockerClient(options *DockerOptions) (*DockerClient, error) {
	dockerHost := options.DockerHost
	tlsVerify := options.DockerTLSVerify

	var (
		client *docker.Client
		err    error
	)

	if tlsVerify == "1" {
		// We're using TLS, let's locate our certs and such
		// boot2docker puts its certs at...
		dockerCertPath := options.DockerCertPath

		// TODO(termie): maybe fast-fail if these don't exist?
		cert := path.Join(dockerCertPath, fmt.Sprintf("cert.pem"))
		ca := path.Join(dockerCertPath, fmt.Sprintf("ca.pem"))
		key := path.Join(dockerCertPath, fmt.Sprintf("key.pem"))
		client, err = docker.NewVersionnedTLSClient(dockerHost, cert, key, ca, "")
		if err != nil {
			return nil, err
		}
	} else {
		client, err = docker.NewClient(dockerHost)
		if err != nil {
			return nil, err
		}
	}
	return &DockerClient{Client: client}, nil
}

// DockerOptions for our docker client
type DockerOptions struct {
	DockerHost      string
	DockerTLSVerify string
	DockerCertPath  string
}

// NewDockerOptions constructor
func NewDockerOptions(c *cli.Context) (*DockerOptions, error) {
	dockerHost := c.String("docker-host")
	dockerTLSVerify := c.String("docker-tls-verify")
	dockerCertPath := c.String("docker-cert-path")

	speculativeOptions := &DockerOptions{
		DockerHost:      dockerHost,
		DockerTLSVerify: dockerTLSVerify,
		DockerCertPath:  dockerCertPath,
	}

	// We're going to try out a few settings and set DockerHost if
	// one of them works
	guessAndUpdateDockerOptions(speculativeOptions)
	return speculativeOptions, nil
}

func guessAndUpdateDockerOptions(opts *DockerOptions) {
	if opts.DockerHost != "" {
		return
	}

	// Check the unix socket, default on linux
	// This will fail instantly so don't bother with the goroutine
	if runtime.GOOS == "linux" {
		unixSocket := "unix:///var/run/docker.sock"
		logger.Println("No Docker host specified, checking", unixSocket)
		client, err := NewDockerClient(&DockerOptions{
			DockerHost: unixSocket,
		})
		if err == nil {
			_, err = client.Version()
			if err == nil {
				opts.DockerHost = unixSocket
				return
			}
		}
	}

	// Check the boot2docker port with default cert paths and such
	b2dCertPath := filepath.Join(os.Getenv("HOME"), ".boot2docker/certs/boot2docker-vm")
	b2dHost := "tcp://192.168.59.103:2376"

	logger.Printf("No Docker host specified, checking for boot2docker", b2dHost)
	client, err := NewDockerClient(&DockerOptions{
		DockerHost:      b2dHost,
		DockerCertPath:  b2dCertPath,
		DockerTLSVerify: "1",
	})
	if err == nil {
		// This can take a long time if it isn't up, so toss it in a
		// goroutine so we can time it out
		result := make(chan bool)
		go func() {
			_, err = client.Version()
			if err == nil {
				result <- true
			} else {
				result <- false
			}
		}()
		select {
		case success := <-result:
			if success {
				opts.DockerHost = b2dHost
				opts.DockerCertPath = b2dCertPath
				opts.DockerTLSVerify = "1"
				return
			}
		case <-time.After(1 * time.Second):
		}
	}

	// Pick a default localhost port and hope for the best :/
	opts.DockerHost = "tcp://127.0.0.1:2375"
	logger.Println("No Docker host found, falling back to default", opts.DockerHost)
}
