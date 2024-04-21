package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const mountAction = "/mountAction"

const DOCKER_LINK = "https://get.docker.com"
const PODMAN_LINK = ""
const CONTAINERD_LINK = ""

func main() {
	fmt.Printf("CEXEC - Chroot Exec\n------------------------\n")

	// Parse the environment variables that are passed into the action
	blockDevice := os.Getenv("BLOCK_DEVICE")
	filesystemType := os.Getenv("FS_TYPE")
	chroot := os.Getenv("CHROOT")
	containerRuntime := strings.ToLower(os.Getenv("CONTAINER_RUNTIME"))

	var exitChroot func() error

	if blockDevice == "" {
		log.Fatalf("No Block Device speified with Environment Variable [BLOCK_DEVICE]")
	}

	// Create the /mountAction mountpoint (no folders exist previously in scratch container)
	err := os.Mkdir(mountAction, os.ModeDir)
	if err != nil {
		log.Fatalf("Error creating the action Mountpoint [%s]", mountAction)
	}

	// Mount the block device to the /mountAction point
	err = syscall.Mount(blockDevice, mountAction, filesystemType, 0, "")
	if err != nil {
		log.Fatalf("Mounting [%s] -> [%s] error [%v]", blockDevice, mountAction, err)
	}
	log.Infof("Mounted [%s] -> [%s]", blockDevice, mountAction)

	if chroot != "" {
		err = MountSpecialDirs()
		if err != nil {
			log.Fatal(err)
		}
		log.Infoln("Changing root before executing command")
		exitChroot, err = Chroot(mountAction)
		if err != nil {
			log.Fatalf("Error changing root to [%s]", mountAction)
		}
	}

	installContainerRuntime(containerRuntime)

	if chroot != "" {
		err = exitChroot()
		if err != nil {
			log.Errorf("Error exiting root from [%s], execution continuing", mountAction)
		}
	}
}

// Chroot handles changing the root, and returning a function to return back to the present directory.
func Chroot(path string) (func() error, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	root, err := os.Open(cwd)
	if err != nil {
		return nil, err
	}

	if err := syscall.Chroot(path); err != nil {
		root.Close()
		return nil, err
	}

	// set the working directory inside container
	if err := syscall.Chdir("/"); err != nil {
		root.Close()
		return nil, err
	}

	return func() error {
		defer root.Close()
		if err := root.Chdir(); err != nil {
			return err
		}
		return syscall.Chroot(".")
	}, nil
}

// MountSpecialDirs ensures that /dev /proc /sys exist in the chroot.
func MountSpecialDirs() error {
	// Mount dev
	dev := filepath.Join(mountAction, "dev")

	if err := syscall.Mount("none", dev, "devtmpfs", syscall.MS_RDONLY, ""); err != nil {
		return fmt.Errorf("couldn't mount /dev to %v: %w", dev, err)
	}

	// Mount proc
	proc := filepath.Join(mountAction, "proc")

	if err := syscall.Mount("none", proc, "proc", syscall.MS_RDONLY, ""); err != nil {
		return fmt.Errorf("couldn't mount /proc to %v: %w", proc, err)
	}

	// Mount sys
	sys := filepath.Join(mountAction, "sys")

	if err := syscall.Mount("none", sys, "sysfs", syscall.MS_RDONLY, ""); err != nil {
		return fmt.Errorf("couldn't mount /sys to %v: %w", sys, err)
	}

	return nil
}

func installContainerRuntime(runtime string) {
	runtimeURLs := map[string]string{
		"docker":     DOCKER_LINK,
		"podman":     PODMAN_LINK,
		"containerd": CONTAINERD_LINK,
	}

	scriptURL, ok := runtimeURLs[runtime]
	if !ok {
		log.Printf("Unsupported container runtime: '%s'", runtime)
		return
	}

	log.Infof("Installing %s...", runtime)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("curl -fsSL %s | sh", scriptURL))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to install %s: %v", runtime, err)
	}
}
