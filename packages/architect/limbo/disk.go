// Package limbo...(TODO)
package limbo

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type BlockDevice struct {
	Name        string        `json:"name"`
	Size        string        `json:"size"`
	Type        string        `json:"type"`
	Model       string        `json:"model"`
	Path        string        `json:"path"`
	Rm          bool          `json:"rm"`
	Ro          bool          `json:"ro"`
	Mountpoints []string      `json:"mountpoints"`
	Children    []BlockDevice `json:"children"`
}

func autoDevice(bds []BlockDevice) (string, error) {
	var best *BlockDevice
	var bestSize int64

	for i := range bds {
		bd := &bds[i]
		if bd.Type != "disk" || bd.Ro || bd.Rm {
			continue
		}

		mounted := false
		for _, child := range bd.Children {
			if len(child.Mountpoints) > 0 {
				mounted = true
				break
			}
		}
		if mounted {
			continue
		}

		size, err := strconv.ParseInt(bd.Size, 10, 64)
		if err != nil {
			continue
		}

		if size > bestSize {
			bestSize = size
			best = bd
		}
	}

	if best == nil {
		return "", fmt.Errorf("no suitable disk found (unmounted, non-removable, writable)")
	}

	return best.Path, nil
}

func promptDevice(bds []BlockDevice) (string, error) {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return "", fmt.Errorf("cannot open tty: %w", err)
	}
	defer tty.Close()

	var disks []BlockDevice
	for _, bd := range bds {
		if bd.Type != "disk" || bd.Ro {
			continue
		}
		disks = append(disks, bd)
	}

	if len(disks) == 0 {
		return "", fmt.Errorf("no disks available")
	}

	fmt.Fprintln(os.Stderr, "NUM\tNAME\tSIZE\tMODEL\tSTATUS")
	for i, bd := range disks {
		status := "available"
		if bd.Rm {
			status = "removable"
		}
		var mounts []string
		for _, child := range bd.Children {
			mounts = append(mounts, child.Mountpoints...)
		}
		if len(mounts) > 0 {
			status = "IN USE (" + strings.Join(mounts, ", ") + ")"
		}
		fmt.Fprintf(os.Stderr, "%d\t%s\t%s\t%s\t%s\n", i+1, bd.Name, bd.Size, bd.Model, status)
	}

	fmt.Fprintf(os.Stderr, "Select disk [1-%d]: ", len(disks))
	scanner := bufio.NewScanner(tty)
	scanner.Scan()
	input := scanner.Text()

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(disks) {
		fmt.Fprintln(os.Stderr, "invalid selection, try again")
		return promptDevice(bds)
	}

	return disks[choice-1].Path, nil
}

func selectDevice(bds []BlockDevice, mode string, placeholder string) (string, error) {
	switch mode {
	case "auto":
		return autoDevice(bds)
	case "prompt":
		return promptDevice(bds)
	case "specific":
		return placeholder, nil
	default:
		return "", fmt.Errorf("unknown disk selection mode: %s", mode)
	}
}
