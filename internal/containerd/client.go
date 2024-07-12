package containerd

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Client struct {
	client *client.Client
}

type InteractiveCommand struct {
	Reader io.Reader
	Writer io.WriteCloser
}

func NewDefaultClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &Client{client: cli}, nil
}

func (c *Client) ExecInteractiveCommand(ctx context.Context, containerId, cmd, user string) (*InteractiveCommand, error) {
	execCreate, err := c.client.ContainerExecCreate(ctx, containerId, container.ExecOptions{
		User:         user,
		Tty:          true,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Detach:       true,
		Cmd:          []string{cmd},
	})
	if err != nil {
		return nil, err
	}

	hijackedResponse, err := c.client.ContainerExecAttach(ctx, execCreate.ID, container.ExecAttachOptions{
		Detach: false,
		Tty:    true,
	})
	if err != nil {
		return nil, err
	}

	return &InteractiveCommand{
		Reader: hijackedResponse.Reader,
		Writer: hijackedResponse.Conn,
	}, nil
}
