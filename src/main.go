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

	"os"
)

type ContainerJob struct {
	ImageName       string
	Command         string
	Debug           bool
	Env             multiValueFlag
	ContainerObject container.ContainerCreateCreatedBody
	Client          *client.Client
	Error           error // keep track of latest error for container client
}

type multiValueFlag []string

func (obj *multiValueFlag) String() string {
	return strings.Join(*obj, " ")
}

func (obj *multiValueFlag) Set(s string) error {
	*obj = append(*obj, s)
	return nil
}

type actionrequest struct {
	Flag flag.FlagSet
}

func (cjob ContainerJob) PullImage(ctx context.Context) {
	reader, err := cjob.Client.ImagePull(ctx, cjob.ImageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)
}

func (cjob *ContainerJob) CreateJob(ctx context.Context) {
	obj, err := cjob.Client.ContainerCreate(ctx, &container.Config{
		Image: cjob.ImageName,
		Tty:   false,
		Cmd:   []string{"tail", "-f", "/dev/null"},
		Env:   cjob.Env,
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/go", os.Getenv("PWD")),
		},
	}, nil, nil, "")

	cjob.ContainerObject = obj
	cjob.Error = err
	cjob.ContainerError()

	cjob.Error = cjob.Client.ContainerStart(ctx, cjob.ContainerObject.ID, types.ContainerStartOptions{})

	cjob.ContainerError()
	fmt.Printf("[%s] Job with image '%s' started successfully\n", cjob.ContainerObject.ID[0:12], cjob.ImageName)

}

func (cjob *ContainerJob) CreateJobWithInput(ctx context.Context) {

	obj, err := cjob.Client.ContainerCreate(ctx, &container.Config{
		AttachStdout: true,
		AttachStderr: true,
		Image:        cjob.ImageName,
		Tty:          false,
		Cmd:          []string{"/bin/sh", "-c", cjob.Command},
		Env:          cjob.Env,
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/go", os.Getenv("PWD")),
		},
	}, nil, nil, "")

	cjob.ContainerObject = obj
	cjob.Error = err
	cjob.ContainerError()

	cjob.Error = cjob.Client.ContainerStart(ctx, cjob.ContainerObject.ID, types.ContainerStartOptions{})

	cjob.ContainerError()

	fmt.Printf("[%s] Job with image '%s' ran successfully\n", cjob.ContainerObject.ID[0:12], cjob.ImageName)

	reader, err := cjob.Client.ContainerLogs(ctx, cjob.ContainerObject.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})

	cjob.Error = err
	cjob.ContainerError()

	defer reader.Close()
	io.Copy(os.Stdout, reader)

	fmt.Printf("[%s] Job with image '%s' finished successfully\n", cjob.ContainerObject.ID[0:12], cjob.ImageName)

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

func (cjob ContainerJob) ExecJob(ctx context.Context, cmd string) {
	fmt.Printf("[%s] Executing '%s' action\n", cjob.ContainerObject.ID[0:12], cmd)
	exec, err := cjob.Client.ContainerExecCreate(ctx, cjob.ContainerObject.ID, types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Tty:          true,
		Cmd:          []string{"/bin/sh", "-c", cmd},
	})

	cjob.Error = err
	cjob.ContainerError()

	attach, err := cjob.Client.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{})

	cjob.Error = err
	cjob.ContainerError()

	defer attach.Close()
	io.Copy(os.Stdout, attach.Reader)

	fmt.Printf("[%s] Action '%s' executed succesfully\n", cjob.ContainerObject.ID[0:12], cmd)
}

func (cjob *ContainerJob) SetFields(a actionrequest) {
	a.Flag.StringVar(&cjob.ImageName, "image", "golang:1.18.0-alpine3.15", "name of image used for build command")
	a.Flag.BoolVar(&cjob.Debug, "d", false, "create and exec is used over run")
	a.Flag.Var(&cjob.Env, "env", "environment variables to be passed")
	a.Flag.Var(&cjob.Env, "e", "environment variables to be passed")

	a.Flag.Parse(os.Args[3:])
	cjob.Command = strings.Join(a.Flag.Args(), " ")

	fmt.Printf("Command: %s\n", cjob.Command)
	fmt.Println(cjob.Env.String())
}

func (a actionrequest) BuildSubCommandExecute() {

	cjob := ContainerJob{}
	cjob.SetFields(a)

	ctx := context.Background()

	cjob.Client, cjob.Error = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	cjob.ContainerError()

	cjob.PullImage(ctx)

	if cjob.Debug {
		cjob.CreateJob(ctx)

		// for loop with execs
		for _, exec := range strings.Split(cjob.Command, "&&") {
			cjob.ExecJob(ctx, strings.TrimRight(strings.TrimLeft(exec, " "), " "))
		}

		// stop container
		cjob.DeleteJob(ctx)
	} else {
		cjob.CreateJobWithInput(ctx)
	}
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
