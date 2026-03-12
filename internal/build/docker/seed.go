package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/moby/moby/client"
	"github.com/previewctl/previewctl-core/types"
)

// RunPrestartSeeds copies seed files/directories into a created (but not started)
// container. Prestart seeds have no cmd — they are copy-only.
func RunPrestartSeeds(ctx context.Context, cli *client.Client, containerID string, seeds []types.SeedEntry, workingDir string) error {
	for i, entry := range seeds {
		fmt.Printf("    prestart seed[%d]: copying %s -> %s\n", i, entry.Source, entry.Destination)
		if err := copyToContainer(ctx, cli, containerID, entry.Source, entry.Destination, workingDir); err != nil {
			return fmt.Errorf("prestart seed[%d]: %w", i, err)
		}
	}
	return nil
}

// RunPoststartSeeds copies seed files/directories into a running container and
// optionally executes a command after each copy.
func RunPoststartSeeds(ctx context.Context, cli *client.Client, containerID string, seeds []types.SeedEntry, workingDir string) error {
	for i, entry := range seeds {
		fmt.Printf("    poststart seed[%d]: copying %s -> %s\n", i, entry.Source, entry.Destination)
		if err := copyToContainer(ctx, cli, containerID, entry.Source, entry.Destination, workingDir); err != nil {
			return fmt.Errorf("poststart seed[%d]: %w", i, err)
		}
		if entry.Cmd != "" {
			fmt.Printf("    poststart seed[%d]: running cmd: %s\n", i, entry.Cmd)
			if err := ExecInContainer(ctx, cli, containerID, entry.Cmd); err != nil {
				return fmt.Errorf("poststart seed[%d] cmd: %w", i, err)
			}
		}
	}
	return nil
}

// WaitHealthy polls the container state until it is running (and healthy, if a
// HEALTHCHECK is defined) or the timeout elapses.
func WaitHealthy(ctx context.Context, cli *client.Client, containerID string, timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("container %s did not become ready within %s", containerID[:12], timeout)
		case <-ticker.C:
			inspect, err := cli.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
			if err != nil {
				return fmt.Errorf("failed to inspect container %s: %w", containerID[:12], err)
			}
			state := inspect.Container.State
			if state == nil {
				continue
			}
			if state.Running {
				if state.Health != nil {
					if state.Health.Status == "healthy" {
						return nil
					}
					continue
				}
				return nil
			}
			if !state.Running && state.ExitCode != 0 {
				return fmt.Errorf("container %s exited with code %d", containerID[:12], state.ExitCode)
			}
		}
	}
}

// ExecInContainer runs a command inside a running container via sh -c.
// Returns an error if the command exits with a non-zero code.
func ExecInContainer(ctx context.Context, cli *client.Client, containerID string, cmd string) error {
	execResp, err := cli.ExecCreate(ctx, containerID, client.ExecCreateOptions{
		Cmd:          []string{"sh", "-c", cmd},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	attachResp, err := cli.ExecAttach(ctx, execResp.ID, client.ExecAttachOptions{})
	if err != nil {
		return fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer attachResp.Close()

	_, _ = io.Copy(os.Stdout, attachResp.Reader)

	inspect, err := cli.ExecInspect(ctx, execResp.ID, client.ExecInspectOptions{})
	if err != nil {
		return fmt.Errorf("failed to inspect exec result: %w", err)
	}
	if inspect.ExitCode != 0 {
		return fmt.Errorf("seed command exited with code %d", inspect.ExitCode)
	}
	return nil
}

// copyToContainer copies a file or directory from the host into a container.
// source is resolved relative to workingDir. destination is the absolute path
// inside the container.
func copyToContainer(ctx context.Context, cli *client.Client, containerID, source, destination, workingDir string) error {
	hostPath := filepath.Join(workingDir, source)

	info, err := os.Stat(hostPath)
	if err != nil {
		return fmt.Errorf("failed to stat source %q: %w", hostPath, err)
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	if info.IsDir() {
		if err := archiveDirectory(tw, hostPath, ""); err != nil {
			return err
		}
	} else {
		if err := archiveFile(tw, hostPath, filepath.Base(destination)); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// For files: extract into the parent directory so the file ends up at destination.
	// For directories: extract into the destination directory itself.
	extractDir := destination
	if !info.IsDir() {
		extractDir = filepath.ToSlash(filepath.Dir(destination))
	}

	_, err = cli.CopyToContainer(ctx, containerID, client.CopyToContainerOptions{
		DestinationPath:           extractDir,
		Content:                   &buf,
		AllowOverwriteDirWithFile: true,
	})
	if err != nil {
		return fmt.Errorf("failed to copy to container: %w", err)
	}
	return nil
}

// archiveFile adds a single file to the tar writer with the given name.
func archiveFile(tw *tar.Writer, hostPath, name string) error {
	data, err := os.ReadFile(hostPath)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", hostPath, err)
	}

	if err := tw.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0o644,
		Size: int64(len(data)),
	}); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}
	if _, err := tw.Write(data); err != nil {
		return fmt.Errorf("failed to write tar content: %w", err)
	}
	return nil
}

// archiveDirectory recursively adds all files and subdirectories to the tar writer.
// prefix is the path prefix inside the tar archive.
func archiveDirectory(tw *tar.Writer, dirPath, prefix string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %q: %w", dirPath, err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		archiveName := entry.Name()
		if prefix != "" {
			archiveName = prefix + "/" + entry.Name()
		}

		if entry.IsDir() {
			if err := tw.WriteHeader(&tar.Header{
				Name:     archiveName + "/",
				Typeflag: tar.TypeDir,
				Mode:     0o755,
			}); err != nil {
				return fmt.Errorf("failed to write dir header: %w", err)
			}
			if err := archiveDirectory(tw, fullPath, archiveName); err != nil {
				return err
			}
		} else {
			info, err := entry.Info()
			if err != nil {
				return fmt.Errorf("failed to stat %q: %w", fullPath, err)
			}
			data, err := os.ReadFile(fullPath)
			if err != nil {
				return fmt.Errorf("failed to read %q: %w", fullPath, err)
			}
			if err := tw.WriteHeader(&tar.Header{
				Name: archiveName,
				Mode: int64(info.Mode().Perm()),
				Size: int64(len(data)),
			}); err != nil {
				return fmt.Errorf("failed to write tar header: %w", err)
			}
			if _, err := tw.Write(data); err != nil {
				return fmt.Errorf("failed to write tar content: %w", err)
			}
		}
	}
	return nil
}
