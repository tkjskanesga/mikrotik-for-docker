package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	ram := getEnv("RAM", "1024M")
	cpu := getEnv("CPU", "1")
	diskPath := getEnv("DISK_PATH", "/storage/chr.img")
	templatePath := "/images/chr.img"

	// Read virtual cable labels, for example: "r1,r2"
	linkEtherStr := getEnv("LINK_ETHER", "")

	// fmt.Printf("==================================================\n")
	arch := runtime.GOARCH
	fmt.Printf("Detecting container architecture: %s\n", arch)
	// fmt.Printf("==================================================\n")

	isNewDisk := false
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		fmt.Println("[INFO] New disk detected, copying CHR template...")
		err := copyFile(templatePath, diskPath)
		if err != nil {
			fmt.Printf("[ERROR] Failed to copy disk template: %v\n", err)
			os.Exit(1)
		}
		isNewDisk = true
	}

	os.MkdirAll("/cfg", 0755)

	if isNewDisk {
		fmt.Println("[INFO] Creating automatic configuration package (branding.tgz)...")
		rscContent := "/ip dhcp-client add interface=ether1 disabled=no comment=\"Winbox-NAT\"\n" +
			"/ip address add address=10.0.2.15/24 interface=ether1\n" +
			"/ip route add gateway=10.0.2.2\n" +
			"/user set admin password=\"admin123\"\n"

		createBrandingTarGz("/cfg/branding.tgz", rscContent)
	}

	args := []string{
		"-nographic",
		"-serial", "mon:stdio",
		"-smp", cpu,
		"-m", ram,
		"-drive", fmt.Sprintf("file=%s,if=none,id=drive-disk0,format=raw", diskPath),
		"-device", "virtio-blk-pci,drive=drive-disk0,id=virtio-disk0,bootindex=1",
		"-drive", "file=fat:rw:/cfg,if=none,id=drive-disk1,format=raw",
		"-device", "virtio-blk-pci,drive=drive-disk1,id=virtio-disk1,bootindex=2",
	}

	if arch == "amd64" {
		_, kvmErr := os.Stat("/dev/kvm")
		if kvmErr == nil {
			args = append(args, "-cpu", "host", "-machine", "type=q35,accel=kvm")
		} else {
			args = append(args, "-cpu", "qemu64", "-machine", "type=q35")
		}
	} else {
		args = append(args, "-cpu", "qemu64", "-machine", "type=q35")
	}

	// INTERFACE 1 (ether1): Base internet / Winbox management path
	netParam0 := "user,id=net0,hostfwd=tcp::80-:80,hostfwd=tcp::8291-:8291"
	args = append(args, "-netdev", netParam0, "-device", "virtio-net-pci,netdev=net0,mac=52:54:00:12:34:56")

	// AUTOMATIC CABLE LOGIC BASED ON LABEL STRING
	if linkEtherStr != "" {
		labels := strings.Split(linkEtherStr, ",")

		for i, label := range labels {
			label = strings.TrimSpace(label)
			if label == "" {
				continue
			}

			// Generate a unique port (range 10000 - 60000) from the label string using FNV hash
			port := generatePortFromLabel(label)
			ifId := fmt.Sprintf("net%d", i+1) // net1 jadi ether2, net2 jadi ether3, dst.

			fmt.Printf("[NET] Connecting %s to virtual cable/bus [%s] via multicast port: %d\n", ifId, label, port)

			// Use standard QEMU multicast IP with a dynamic port derived from the label hash
			netParam := fmt.Sprintf("socket,id=%s,mcast=230.0.0.1:%d", ifId, port)
			macAddr := generateRandomMac()

			args = append(args, "-netdev", netParam, "-device", "virtio-net-pci,netdev="+ifId+",mac="+macAddr)
		}
	}

	fmt.Printf("\nStarting MikroTik CHR v7 (x86_64)...\n\n")

	cmd := exec.Command("qemu-system-x86_64", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("[ERROR] QEMU stopped unexpectedly: %v\n", err)
		os.Exit(1)
	}
}

// Convert a label string (for example "r1") into a consistent integer port (10000 - 60000)
func generatePortFromLabel(label string) int {
	h := fnv.New32a()
	h.Write([]byte(label))
	hashValue := h.Sum32()
	return 10000 + int(hashValue%50000)
}

func generateRandomMac() string {
	buf := make([]byte, 4)
	_, err := rand.Read(buf)
	if err != nil {
		return "52:54:00:99:99:99"
	}
	return fmt.Sprintf("52:54:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3])
}

func createBrandingTarGz(targetPath, rscContent string) error {
	file, err := os.Create(targetPath); if err != nil { return err }; defer file.Close()
	gw := gzip.NewWriter(file); defer gw.Close()
	tw := tar.NewWriter(gw); defer tw.Close()
	body := []byte(rscContent)
	header := &tar.Header{Name: "defconf.rsc", Mode: 0644, Size: int64(len(body))}
	tw.WriteHeader(header)
	tw.Write(body)
	return nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists { return value }
	return fallback
}

func copyFile(src, dst string) error {
	in, err := os.Open(src); if err != nil { return err }; defer in.Close()
	out, err := os.Create(dst); if err != nil { return err }; defer out.Close()
	_, err = io.Copy(out, in); return out.Sync()
}
