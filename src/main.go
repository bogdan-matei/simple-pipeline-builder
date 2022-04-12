package main

import (
	// local package
	"flag"
	"spb-job/commons"

	"context"
	"fmt"
	"io"
	"strings"

	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type GoJobSpec struct {
	JobSpec commons.JobSpec
	Flags   *flag.FlagSet
	Error   error
}

func (cjob *GoJobSpec) Run() error {
	ctx := context.Background()

	cjob.JobSpec.Client, cjob.Error = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	cjob.JobError()

	cjob.PullImage(ctx)

	if cjob.JobSpec.Debug {
		cjob.CreateJob(ctx)

		// for loop with execs
		for _, exec := range strings.Split(cjob.JobSpec.Command, "&&") {
			cjob.ExecJob(ctx, strings.TrimRight(strings.TrimLeft(exec, " "), " "))
		}

		// stop container
		if !cjob.JobSpec.Persistance {
			cjob.DeleteJob(ctx)
		}
	} else {
		cjob.CreateJobWithInput(ctx)
	}
	return nil
}

func (cjob *GoJobSpec) ParseFlags() error {
	cjob.Flags.StringVar(&cjob.JobSpec.ImageName, "image", "golang:1.18.0-alpine3.15", "name of image used for build command")
	cjob.Flags.BoolVar(&cjob.JobSpec.Debug, "d", false, "create and exec is used over run")
	cjob.Flags.Var(&cjob.JobSpec.Env, "env", "environment variables to be passed")
	cjob.Flags.Var(&cjob.JobSpec.Env, "e", "environment variables to be passed")
	cjob.Flags.BoolVar(&cjob.JobSpec.Persistance, "persistance", false, "prevents deletion of the job while in debug mode")

	cjob.Flags.Parse(os.Args[2:])

	cjob.JobSpec.Command = strings.Join(cjob.Flags.Args(), " ")

	fmt.Printf("Command: %s\n", cjob.JobSpec.ImageName)
	fmt.Println(cjob.JobSpec.Env.String())

	return nil
}

func (cjob GoJobSpec) JobError() {
	if cjob.Error != nil {
		panic(cjob.Error)
	}

}

func (cjob GoJobSpec) PullImage(ctx context.Context) {
	reader, err := cjob.JobSpec.Client.ImagePull(ctx, cjob.JobSpec.ImageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)
}

func (cjob *GoJobSpec) CreateJob(ctx context.Context) {
	cjob.JobSpec.ContainerObject, cjob.Error = cjob.JobSpec.Client.ContainerCreate(ctx, &container.Config{
		Image: cjob.JobSpec.ImageName,
		Tty:   false,
		Cmd:   []string{"tail", "-f", "/dev/null"},
		Env:   cjob.JobSpec.Env,
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/go", os.Getenv("PWD")),
		},
	}, nil, nil, "")

	(*cjob).JobError()

	cjob.Error = cjob.JobSpec.Client.ContainerStart(ctx, cjob.JobSpec.ContainerObject.ID, types.ContainerStartOptions{})

	(*cjob).JobError()

	fmt.Printf("[%s] Job with image '%s' started successfully\n", cjob.JobSpec.ContainerObject.ID[0:12], cjob.JobSpec.ImageName)

}

func (cjob *GoJobSpec) CreateJobWithInput(ctx context.Context) {

	cjob.JobSpec.ContainerObject, cjob.Error = cjob.JobSpec.Client.ContainerCreate(ctx, &container.Config{
		AttachStdout: true,
		AttachStderr: true,
		Image:        cjob.JobSpec.ImageName,
		Tty:          false,
		Cmd:          []string{"/bin/sh", "-c", cjob.JobSpec.Command},
		Env:          cjob.JobSpec.Env,
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/go", os.Getenv("PWD")),
		},
	}, nil, nil, "")

	cjob.JobError()

	cjob.Error = cjob.JobSpec.Client.ContainerStart(ctx, cjob.JobSpec.ContainerObject.ID, types.ContainerStartOptions{})

	cjob.JobError()

	fmt.Printf("[%s] Job with image '%s' ran successfully\n", cjob.JobSpec.ContainerObject.ID[0:12], cjob.JobSpec.ImageName)

	reader, err := cjob.JobSpec.Client.ContainerLogs(ctx, cjob.JobSpec.ContainerObject.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})

	cjob.Error = err
	cjob.JobError()

	defer reader.Close()
	io.Copy(os.Stdout, reader)

	fmt.Printf("[%s] Job with image '%s' finished successfully\n", cjob.JobSpec.ContainerObject.ID[0:12], cjob.JobSpec.ImageName)

}

func (cjob GoJobSpec) DeleteJob(ctx context.Context) {
	fmt.Printf("[%s] Deleting job\n", cjob.JobSpec.ContainerObject.ID[0:12])
	cjob.Error = cjob.JobSpec.Client.ContainerRemove(ctx, cjob.JobSpec.ContainerObject.ID, types.ContainerRemoveOptions{
		Force: true,
	})

	cjob.JobError()
	fmt.Printf("[%s] Job deleted succesfully\n", cjob.JobSpec.ContainerObject.ID[0:12])

}

func (cjob GoJobSpec) ExecJob(ctx context.Context, cmd string) {
	fmt.Printf("[%s] Executing '%s' action\n", cjob.JobSpec.ContainerObject.ID[0:12], cmd)
	exec, err := cjob.JobSpec.Client.ContainerExecCreate(ctx, cjob.JobSpec.ContainerObject.ID, types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Tty:          true,
		Cmd:          []string{"/bin/sh", "-c", cmd},
	})

	cjob.Error = err
	cjob.JobError()

	attach, err := cjob.JobSpec.Client.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{})

	cjob.Error = err
	cjob.JobError()

	defer attach.Close()
	io.Copy(os.Stdout, attach.Reader)

	fmt.Printf("[%s] Action '%s' executed succesfully\n", cjob.JobSpec.ContainerObject.ID[0:12], cmd)
}

func Run(jobRun commons.JobRun) {
	jobRun.ParseFlags()
	jobRun.Run()
}

func main() {

	// load plugins from source

	// refresh state

	// run action
	flag := flag.NewFlagSet(os.Args[1], flag.PanicOnError)
	switch os.Args[1] {
	case "run-go":
		goJob := GoJobSpec{Flags: flag}
		Run(&goJob)
	default:
		fmt.Println("Plugin doesn't exist")
		os.Exit(1)
	}
}
