package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tailored-agentic-units/container"
	"github.com/tailored-agentic-units/container/docker"
)

func main() {
	docker.Register()

	rt, err := container.Create("docker")
	if err != nil {
		log.Fatalf("create runtime: %v", err)
	}

	ctx := context.Background()

	c, err := rt.Create(ctx, container.CreateOptions{
		Image:  "alpine:3.21",
		Name:   "tau-docker-hello",
		Cmd:    []string{"sh", "-c", "trap exit TERM; sleep infinity & wait"},
		Labels: map[string]string{"example": "docker-hello"},
	})
	if err != nil {
		log.Fatalf("create container: %v", err)
	}

	if err := rt.Start(ctx, c.ID); err != nil {
		log.Fatalf("start: %v", err)
	}

	res, err := rt.Exec(ctx, c.ID, container.ExecOptions{
		Cmd:          []string{"echo", "hello from", c.Name},
		AttachStdout: true,
	})
	if err != nil {
		log.Fatalf("exec: %v", err)
	}
	fmt.Print(string(res.Stdout))

	info, err := rt.Inspect(ctx, c.ID)
	if err != nil {
		log.Fatalf("inspect: %v", err)
	}
	fmt.Printf("manifest: %v (alpine carries no tau manifest)\n", info.Manifest)

	m := info.Manifest
	if m == nil {
		m = container.Fallback()
	}
	fmt.Printf("fallback shell: %s\n", m.Shell)

	if err := rt.Stop(ctx, c.ID, 5*time.Second); err != nil {
		log.Fatalf("stop: %v", err)
	}
	if err := rt.Remove(ctx, c.ID, false); err != nil {
		log.Fatalf("remove: %v", err)
	}
}
