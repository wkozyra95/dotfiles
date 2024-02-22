package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	specs "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/fn"
)

type DockerRunOptions struct {
	Rebuild bool
}

var log = logger.NamedLogger("docker")

type DockerContext struct {
	cl client.Client
}

func NewDockerContext() *DockerContext {
	dockerClient, clientErr := client.NewClientWithOpts(client.WithVersion("1.43"))
	if clientErr != nil {
		panic(clientErr)
	}
	return &DockerContext{
		cl: *dockerClient,
	}
}

func Run(ctx *DockerContext, dockerEnv env.DockerEnvSpec, options DockerRunOptions) error {
	image, findImageErr := ctx.findImageByName(dockerEnv.ImageName)
	if findImageErr != nil {
		return findImageErr
	}
	didRebuildImage := false
	if image == nil || options.Rebuild { // or rebuild flag
		_, err := BuildImageWithoutContext(ctx, dockerEnv.ImageName, dockerEnv.DockerfilePath)
		if err != nil {
			return err
		}
		didRebuildImage = true
	}

	container, findContainerErr := ctx.findContainerByName(dockerEnv.ContainerName)
	if findContainerErr != nil {
		return findContainerErr
	}
	if container == nil || didRebuildImage {
		newContainer, err := BuildContainer(ctx, dockerEnv.ImageName, dockerEnv.ContainerName)
		if err != nil {
			return err
		}
		container = newContainer
	}
	if container.Status != "running" {
		log.Info("Starting docker container")
		if err := ctx.ContainerStart(container.ID); err != nil {
			return err
		}
	} else {
		log.Info("Container already started")
	}
	return exec.Command().WithStdio().Args("docker", "exec", "-it", container.ID, "zsh").Run()
}

func BuildImageWithoutContext(
	ctx *DockerContext,
	imageName string,
	dockerfilePath string,
) (*types.ImageSummary, error) {
	log.Info("Building image")
	userUID := fmt.Sprintf("%d", os.Getuid())
	userGID := fmt.Sprintf("%d", os.Getgid())

	buildOptions := types.ImageBuildOptions{
		Dockerfile: path.Base(dockerfilePath),
		Tags:       []string{imageName},
		BuildArgs: map[string]*string{
			"HOST_UID": &userUID,
			"HOST_GID": &userGID,
		},
		Remove:      true,
		ForceRemove: true,
	}
	buffer, contextError := createContext(path.Dir(dockerfilePath))
	if contextError != nil {
		log.Errorf("Preparing docker context failed with error %v", contextError)
		return nil, contextError
	}
	buildContext := bytes.NewReader(buffer.Bytes())

	buildResponse, buildErr := ctx.cl.ImageBuild(context.Background(), buildContext, buildOptions)
	if buildErr != nil {
		log.Fatal(buildErr)
	}
	defer buildResponse.Body.Close()

	output, readErr := io.ReadAll(buildResponse.Body)
	if readErr != nil {
		return nil, readErr
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
			return nil, err
		}
		if stream, ok := parsed["stream"].(string); ok {
			logBuf.WriteString(stream)
		} else if errorMsg, ok := parsed["error"].(string); ok {
			logBuf.WriteString(errorMsg)
		}
	}
	log.Info(logBuf.String())
	if buildErr != nil {
		return nil, buildErr
	}

	image, findImageErr := ctx.findImageByName(imageName)
	if findImageErr != nil {
		return nil, findImageErr
	}
	if image == nil {
		return nil, errors.New("Image not found")
	}
	return image, nil
}

func BuildContainer(
	ctx *DockerContext,
	imageName string,
	containerName string,
) (*types.Container, error) {
	if err := ctx.ensureContainerRemoved(containerName); err != nil {
		return nil, err
	}

	log.Info("Creating container")
	cwd := fn.Must(os.Getwd())
	_, createErr := ctx.cl.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:        imageName,
			Tty:          true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
		},
		&container.HostConfig{Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: cwd,
				Target: "/home/wojtek/project",
			},
		}},
		&network.NetworkingConfig{},
		&specs.Platform{},
		containerName,
	)
	if createErr != nil {
		return nil, createErr
	}
	container, findErr := ctx.findContainerByName(containerName)
	if findErr != nil {
		return nil, findErr
	}
	if container == nil {
		return nil, errors.New("Container not found")
	}
	return container, nil
}

func createContext(directory string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	reader, tarErr := archive.Tar(directory, archive.Uncompressed)
	if tarErr != nil {
		return nil, tarErr
	}
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, err
	}
	return buf, nil
}

func (ctx *DockerContext) findImageByName(name string) (*types.ImageSummary, error) {
	args := filters.NewArgs()
	args.Add("reference", name)
	list, listErr := ctx.cl.ImageList(context.Background(), types.ImageListOptions{Filters: args})
	if listErr != nil {
		return nil, listErr
	}
	if len(list) == 0 {
		return nil, nil
	}
	return &list[0], nil
}

func (ctx *DockerContext) findContainerByName(containerName string) (*types.Container, error) {
	containers, listErr := ctx.cl.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if listErr != nil {
		return nil, listErr
	}
	for _, container := range containers {
		for _, name := range container.Names {
			if name == fmt.Sprintf("/%s", containerName) {
				return &container, nil
			}
		}
	}
	return nil, nil
}

func (ctx *DockerContext) ensureContainerRemoved(containerName string) error {
	cont, containerErr := ctx.findContainerByName(containerName)
	if containerErr != nil {
		return containerErr
	}
	if cont != nil {
		if cont.State == "running" {
			log.Infof("Stopping %s container", containerName)
			if err := ctx.cl.ContainerStop(context.Background(), cont.ID, container.StopOptions{}); err != nil {
				return err
			}
		}
		log.Infof("Removing %s container", containerName)
		return ctx.ContainerRemove(cont.ID)
	}
	return nil
}

func (ctx *DockerContext) ListContainers() ([]types.Container, error) {
	containers, listErr := ctx.cl.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if listErr != nil {
		return nil, listErr
	}
	return containers, nil
}

func (ctx *DockerContext) ContainerStart(containerID string) error {
	startErr := ctx.cl.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
	if startErr != nil {
		return startErr
	}
	return nil
}

func (ctx *DockerContext) ContainerRemove(containerID string) error {
	removeErr := ctx.cl.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})
	if removeErr != nil {
		return removeErr
	}
	return nil
}
