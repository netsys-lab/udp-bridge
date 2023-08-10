# UDP-Bridge - Tunnel UDP traffic over TCP with fixed endpoints on both sites
This repo contains the code for a program that allows to tunnel UDP traffic over TCP between two applications on different hosts that
a) listen on fixed port numbers
b) expect the remote applications to listen on fixed port numbers

Our current use-case is to create a tunnel between two external Border Router interfaces in SCION, in case some firewall is only allowing TCP traffic. However, it can be used for other use-cases, too.

The name of this tool is intended to be different from the existing [udptunnel]() which provides similar functionalities, but did not fulfill our use-case.

## Idea


## Usage
Build it locally and then run it `go build && ./udp-bridge [OPTIONS]` or run it via `go run github.com/netsys-lab/udp-bridge [OPTIONS]`

```
Usage:
  udp-bridge [OPTIONS]

Application Options:
  -s, --server=  Configure the TCP Addr IP:Port to listen in Server Mode
  -c, --client=  Configure the TCP Addr IP:Port to connect to in Client Mode
  -u, --udp=     Configure the UDP Addr IP:Port to listen for incoming packets
  -d, --dest=    Configure the UDP Addr IP:Port to sent incoming packets to
  -v, --verbose  Show verbose debug information

Help Options:
  -h, --help     Show this help message
```