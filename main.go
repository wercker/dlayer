package main

import (
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
)

var (
	logger      = log.New(os.Stdout, "", 0)
	dockerFlags = []cli.Flag{
		cli.StringFlag{Name: "docker-host", Value: "", Usage: "Docker api endpoint.", EnvVar: "DOCKER_HOST"},
		cli.StringFlag{Name: "docker-tls-verify", Value: "0", Usage: "Docker api tls verify.", EnvVar: "DOCKER_TLS_VERIFY"},
		cli.StringFlag{Name: "docker-cert-path", Value: "", Usage: "Docker api cert path.", EnvVar: "DOCKER_CERT_PATH"},
	}
	sizesCommand = cli.Command{
		Name:  "sizes",
		Usage: "output layer sizes",
		Action: func(c *cli.Context) {
			opts, err := NewDockerOptions(c)
			if err != nil {
				logger.Println("Invalid options\n", err)
				os.Exit(1)
			}
			err = cmdSizes(opts)
			if err != nil {
				os.Exit(1)
			}
		},
		Flags: dockerFlags,
	}
)

type ImageStats struct {
	Total     int
	ParentMap map[string]map[string]struct{}
	ImageMap  map[string]docker.APIImages
	TaggedMap map[string]docker.APIImages
}

func getStats(opts *DockerOptions) (*ImageStats, error) {
	client, err := NewDockerClient(opts)
	if err != nil {
		return nil, err
	}

	images, err := client.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		return nil, err
	}

	total := len(images)
	parentMap := map[string]map[string]struct{}{}
	imageMap := map[string]docker.APIImages{}
	taggedMap := map[string]docker.APIImages{}

	for _, image := range images {
		if image.ParentID != "" {
			children, ok := parentMap[image.ParentID]
			if !ok {
				children = map[string]struct{}{}
			}
			children[image.ID] = struct{}{}
			parentMap[image.ParentID] = children
		}
		imageMap[image.ID] = image
		if len(image.RepoTags) != 0 && image.RepoTags[0] != "<none>:<none>" {
			taggedMap[image.ID] = image
		}
	}

	return &ImageStats{
		Total:     total,
		ParentMap: parentMap,
		ImageMap:  imageMap,
		TaggedMap: taggedMap,
	}, nil

}

// trace the layers through the parent map
func trace(stats *ImageStats, imageID string) map[string]docker.APIImages {
	layers := map[string]docker.APIImages{}

	currentImage := stats.ImageMap[imageID]
	layers[currentImage.ID] = currentImage

	for {
		parent := currentImage.ParentID
		if parent == "" {
			break
		}
		currentImage = stats.ImageMap[parent]
		layers[currentImage.ID] = currentImage
	}

	return layers
}

func cmdSizes(opts *DockerOptions) error {
	stats, err := getStats(opts)
	if err != nil {
		return err
	}

	allFoundLayers := map[string]docker.APIImages{}
	var allFoundSize int64
	allSharedLayers := map[string]docker.APIImages{}
	var allSharedSize int64
	var virtualSize int64

	// calculate shared information between tagged images
	for _, image := range stats.TaggedMap {
		layers := trace(stats, image.ID)

		fmt.Printf(
			"Tag %-40v: %4d layers - %4dMB (virtual)\n",
			image.RepoTags[0],
			len(layers),
			image.VirtualSize/(1024*1024),
		)
		virtualSize += image.VirtualSize

		for _, layer := range layers {
			if _, ok := allFoundLayers[layer.ID]; ok {
				// add it if it isn't already there
				if _, ok := allSharedLayers[layer.ID]; !ok {
					allSharedLayers[layer.ID] = layer
					allSharedSize += layer.Size
				}
			} else {
				allFoundLayers[layer.ID] = layer
				allFoundSize += layer.Size
			}
		}
	}

	var totalSize int64
	for _, image := range stats.ImageMap {
		totalSize += image.Size
	}

	fmt.Printf(
		"Total    : %4d layers - %6dMB (actual)\n",
		len(stats.ImageMap),
		totalSize/(1024*1024),
	)

	fmt.Printf(
		"Reachable: %4d layers - %6dMB (actual)\n",
		len(allFoundLayers),
		allFoundSize/(1024*1024),
	)
	fmt.Printf(
		"                         %6dMB (virtual)\n",
		virtualSize/(1024*1024),
	)

	fmt.Printf(
		"Shared   : %4d layers - %6dMB (actual)\n",
		len(allSharedLayers),
		allSharedSize/(1024*1024),
	)

	return nil
}

func main() {
	app := cli.NewApp()
	app.Author = "Team wercker"
	app.Name = "dlayer"
	app.Usage = ""
	app.Commands = []cli.Command{
		sizesCommand,
	}
	app.Run(os.Args)
}
