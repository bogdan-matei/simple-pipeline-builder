package commons

import (
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type multiValueFlag []string

func (obj *multiValueFlag) String() string {
	return strings.Join(*obj, " ")
}

func (obj *multiValueFlag) Set(s string) error {
	*obj = append(*obj, s)
	return nil
}

type JobSpec struct {
	ImageName       string
	Command         string
	Env             multiValueFlag
	ContainerObject container.ContainerCreateCreatedBody
	Client          *client.Client
	Debug           bool
	Persistance     bool
}
