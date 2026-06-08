# Mikrotik on Docker

This repository provides a simple way to run MikroTik CHR in Docker using QEMU on x86_64 containers. It is meant for learning, lab testing, and small network demonstrations.

The included `compose.yml` is only a sample startup example. You can change the service names, ports, storage paths, and environment variables to match your own setup.

## Requirements

Make sure your host has:

- Docker Engine
- Docker Compose v2
- Privileged container support
- Enough CPU and memory for QEMU-based virtualization
- Optional: KVM support on AMD64 hosts for better performance

## Installation

### 1. Clone the repository

```bash
git clone https://github.com/tkjskanesga/mikrotik-for-docker.git
cd mikrotik-for-docker
```

### 2. Build the image

```bash
docker build -t mikrotik-on-docker:latest .
```

### 3. Start the sample Compose file

```bash
docker compose up -d --build
```

This is only an example. You are free to edit `compose.yml` for your own environment.

## Environment Variables

The container entrypoint uses these variables:

- `RAM` — QEMU memory allocation, default: `1024M`
- `CPU` — number of vCPUs, default: `1`
- `DISK_PATH` — CHR disk path inside the container, default: `/storage/chr.img`
- `LINK_ETHER` — comma-separated labels for virtual links, for example `r1,r2`

Example:

```yaml
environment:
  - RAM=1024M
  - CPU=1
  - LINK_ETHER=r1,r2
```

## Other Ways to Run the Container

You can also run the image directly without Compose.

```bash
docker run -d --privileged \
  --name mikrotik-chr \
  -e RAM=1024M \
  -e CPU=1 \
  -e LINK_ETHER=r1 \
  -p 8291:8291 \
  -p 80:80 \
  -v "$PWD/router:/storage" \
  mikrotik-on-docker:latest
```

You can customize:

- `-p` for host ports
- `-v` for persistent storage
- `RAM` and `CPU` for resource limits
- `LINK_ETHER` for virtual link labels

## Access

After the container is running, access the MikroTik CHR instance through the published ports:

- Web interface: `http://localhost:80`
- WinBox: `localhost:8291`

If you use the sample Compose file, the ports are already defined there and can be changed as needed.

## Support

If you need help, want to report an issue, or would like to contribute improvements, you are welcome to open an issue or submit a pull request on GitHub.

Please make sure your report includes:

- what you expected to happen
- what actually happened
- the relevant command or configuration you used
- any error messages or logs

