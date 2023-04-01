package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type dockerInstance struct {
	dockerfile    string
	imageName     string
	containerName string
	client        *client.Client
	cwd           string
	mounts        []mount.Mount
}

func (d *dockerInstance) init() {
	if d.client != nil {
		panic("already initialised")
	}
	if d.dockerfile == "" {
		d.dockerfile = "Dockerfile"
	}
	if d.containerName == "" {
		d.containerName = "system_setup_test"
	}
	if d.imageName == "" {
		d.imageName = "system_setup_image"
	}
	cli, cliErr := client.NewClientWithOpts(client.WithVersion("1.39"))
	if cliErr != nil {
		panic(cliErr)
	}
	d.client = cli
}

func (d *dockerInstance) start() error {
	log.Info("Preparing docker ...")
	d.init()
	image, findImageErr := d.findImage()
	if findImageErr != nil {
		return findImageErr
	}
	if image == nil {
		if err := d.buildImage(); err != nil {
			return err
		}
	} else {
		log.Info("Image already exists")
	}
	container, findContainerErr := d.findContainer()
	if findContainerErr != nil {
		newContainer, buildErr := d.buildContainer()
		if buildErr != nil {
			return buildErr
		}
		container = newContainer
	} else {
		log.Info("Container already exists")
	}

	if container.Status != "running" {
		log.Info("Starting docker container")
		if err := d.client.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}
	} else {
		log.Info("Container already started")
	}

	return nil
}

func (d *dockerInstance) stop() error {
	log.Info("Stopping docker container")

	cont, findErr := d.findContainer()
	if findErr != nil {
		return findErr
	}

	if err := d.client.ContainerStop(context.Background(), cont.ID, container.StopOptions{}); err != nil {
		return err
	}
	return nil
}

func (d *dockerInstance) remove() error {
	container, findErr := d.findContainer()
	if findErr != nil {
		return findErr
	}
	log.Info("Removing container")
	if err := d.client.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}
	return nil
}

func (d *dockerInstance) buildContainer() (*types.Container, error) {
	log.Info("Creating new docker container")
	_, createErr := d.client.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:        d.imageName,
			Tty:          true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
		},
		&container.HostConfig{Mounts: d.mounts},
		&network.NetworkingConfig{},
		&specs.Platform{},
		d.containerName,
	)
	if createErr != nil {
		return nil, createErr
	}
	container, findErr := d.findContainer()
	if findErr != nil {
		return nil, findErr
	}
	return container, nil
}

func (d *dockerInstance) createDockerContext() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	reader, tarErr := archive.Tar(d.cwd, archive.Uncompressed)
	if tarErr != nil {
		return nil, tarErr
	}
	buf.ReadFrom(reader)
	return buf, nil
}

func (d *dockerInstance) buildImage() error {
	log.Info("Creating docker image from dockerfile")
	buffer, contextError := d.createDockerContext()
	if contextError != nil {
		log.Errorf("Preparing docker context failed with error %v", contextError)
		return contextError
	}
	buildContext := bytes.NewReader(buffer.Bytes())

	buildOptions := types.ImageBuildOptions{
		Dockerfile:  d.dockerfile,
		Tags:        []string{d.imageName},
		Remove:      true,
		ForceRemove: true,
	}

	buildResponse, buildErr := d.client.ImageBuild(context.Background(), buildContext, buildOptions)
	if buildErr != nil {
		log.Fatal(buildErr)
	}
	defer buildResponse.Body.Close()

	output, readErr := ioutil.ReadAll(buildResponse.Body)
	if readErr != nil {
		return readErr
	}

	var logBuf bytes.Buffer
	logBuf.WriteString("Docker output\n")
	rawJsons := strings.Split(string(output), "\n")
	for _, rawJson := range rawJsons {
		if rawJson == "" {
			continue
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(rawJson), &parsed); err != nil {
			log.Errorf("Parsing of docker output failed %v", rawJson)
			return err
		}
		if stream, ok := parsed["stream"].(string); ok {
			logBuf.WriteString(stream)
		} else if errorMsg, ok := parsed["error"].(string); ok {
			logBuf.WriteString(errorMsg)
		}
	}
	log.Info(logBuf.String())

	return buildErr
}

func (d *dockerInstance) findImage() (*types.ImageSummary, error) {
	log.Info("Looking for existing images locally")
	args := filters.NewArgs()
	args.Add("reference", d.imageName)
	list, listErr := d.client.ImageList(context.Background(), types.ImageListOptions{Filters: args})
	if listErr != nil {
		return nil, listErr
	}
	if len(list) == 0 {
		return nil, nil
	}
	return &list[0], nil
}

func (d *dockerInstance) findContainer() (*types.Container, error) {
	containers, listErr := d.client.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if listErr != nil {
		return nil, listErr
	}
	for _, container := range containers {
		for _, name := range container.Names {
			if name == fmt.Sprintf("/%s", d.containerName) {
				return &container, nil
			}
		}
	}
	return nil, fmt.Errorf("not found")
}

func (d *dockerInstance) exec(cmd string) error {
	//	container, findErr := d.findContainer()
	//	if findErr != nil {
	//		return findErr
	//	}

	return nil
}
