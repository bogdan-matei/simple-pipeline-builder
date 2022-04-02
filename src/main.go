package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	// "github.com/docker/docker/pkg/stdcopy"

	"os"
)

type ContainerJob struct {
	ImageName       *string
	SourcePath      *string
	Commands        []string
	ContainerObject container.ContainerCreateCreatedBody
	Client          *client.Client
	Error           error // keep track of latest error for container client
}

type actionrequest struct {
	Flag flag.FlagSet
}

func (cjob ContainerJob) PullImage(ctx context.Context) {
	reader, err := cjob.Client.ImagePull(ctx, *cjob.ImageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer reader.Close()
	// io.Copy(os.Stdout, reader)
}

func (cjob *ContainerJob) CreateJob(ctx context.Context) {
	obj, err := cjob.Client.ContainerCreate(ctx, &container.Config{
		Image: *cjob.ImageName,
		Tty:   false,
		Cmd:   []string{"tail", "-f", "/dev/null"},
		Env:   []string{"GOPATH", "GO111MODULE=auto"},
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s/projects/spb/src:/go", os.Getenv("HOME")),
		},
	}, nil, nil, "")

	cjob.ContainerObject = obj
	cjob.Error = err
	cjob.ContainerError()

	cjob.Error = cjob.Client.ContainerStart(ctx, cjob.ContainerObject.ID, types.ContainerStartOptions{})

	cjob.ContainerError()
	fmt.Printf("[%s] Job with image '%s' started successfully\n", cjob.ContainerObject.ID[0:12], *cjob.ImageName)

}

func (cjob ContainerJob) ContainerError() {
	if cjob.Error != nil {
		panic(cjob.Error)
	}
}

func (cjob ContainerJob) DeleteJob(ctx context.Context) {
	fmt.Printf("[%s] Deleting job\n", cjob.ContainerObject.ID[0:12])
	cjob.Error = cjob.Client.ContainerRemove(ctx, cjob.ContainerObject.ID, types.ContainerRemoveOptions{
		Force: true,
	})

	cjob.ContainerError()
	fmt.Printf("[%s] Job deleted succesfully\n", cjob.ContainerObject.ID[0:12])

}

func (cjob ContainerJob) ExecJob(ctx context.Context, action string) {
	fmt.Printf("[%s] Executing '%s' action\n", cjob.ContainerObject.ID[0:12], action)
	exec, err := cjob.Client.ContainerExecCreate(ctx, cjob.ContainerObject.ID, types.ExecConfig{
		Detach: true,
		Cmd:    []string{"/bin/sh", "-c", action},
	})

	cjob.Error = err
	cjob.ContainerError()

	err = cjob.Client.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{
		Detach: true,
		Tty:    true,
	})

	cjob.Error = err
	cjob.ContainerError()

	fmt.Printf("[%s] Action '%s' executed succesfully\n", cjob.ContainerObject.ID[0:12], action)
}

func (cjob *ContainerJob) SetFields(a actionrequest) {
	cjob.ImageName = a.Flag.String("image", "golang:1.18.0-alpine3.15", "name of image used for build command")
	cjob.SourcePath = a.Flag.String("source-path", "", "path of the directory or file you want to run the jobs on")
	a.Flag.Parse(os.Args[3:])
	cjob.Commands = strings.Split(strings.Join(a.Flag.Args(), " "), ",")
	fmt.Println("Commands: %v", cjob.Commands)
}

func (a actionrequest) BuildSubCommandExecute() {

	cjob := ContainerJob{}
	cjob.SetFields(a)

	ctx := context.Background()

	cjob.Client, cjob.Error = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	cjob.ContainerError()

	cjob.PullImage(ctx)
	cjob.CreateJob(ctx)

	// for loop with execs
	// cjob.ExecJob(ctx, cjob.Commands)
	for _, exec := range cjob.Commands {
		cjob.ExecJob(ctx, exec)
	}

	reader, err := cjob.Client.ContainerLogs(ctx, cjob.ContainerObject.ID, types.ContainerLogsOptions{ShowStdout: true})

	cjob.Error = err
	cjob.ContainerError()

	defer reader.Close()
	io.Copy(os.Stdout, reader)

	// cjob.DeleteJob(ctx)
}

func (a actionrequest) Execute() error {

	if len(os.Args) < 3 {
		return errors.New("missing subcommand")
	}

	switch os.Args[2] {
	case "build":
		build := actionrequest{Flag: *flag.NewFlagSet(os.Args[2], flag.PanicOnError)}
		build.BuildSubCommandExecute()
	}

	return nil
}

func main() {

	// load plugins from source

	// refresh state

	// run action
	tmp := actionrequest{Flag: *flag.NewFlagSet("run", flag.PanicOnError)}

	switch os.Args[1] {
	case "run":
		err := tmp.Execute()

		if err != nil {
			panic(err)
		}

	}

}
