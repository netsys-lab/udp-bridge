# UDP-Bridge - Tunnel UDP traffic over TCP with fixed endpoints on both sites
This repo contains the code for a program that allows to tunnel UDP traffic over TCP between two applications on different hosts that
a) listen on fixed port numbers
b) expect the remote applications to listen on fixed port numbers

Our current use-case is to create a tunnel between two external Border Router interfaces in SCION, in case some firewall is only allowing TCP traffic. However, it can be used for other use-cases, too.

The name of this tool is intended to be different from the existing [udptunnel](https://manpages.ubuntu.com/manpages/jammy/man1/udptunnel.1.html) which provides similar functionalities, but did not fulfill our use-case.

## Idea
![udp-bridge](https://github.com/netsys-lab/udp-bridge/assets/32448709/83e1d489-1f7b-4a8d-9983-101d7c83d112)

### Configuration for this example
We now build the setup depicted in the figure above with `udp-bridge`. Assuming Application1 listens on `10.0.0.1:30001` and Application2 on `10.0.0.2:40001`, and there is a firewall in between filtering UDP traffic. We now configure udp-bridge1 to listen on `10.0.1.1:30000` on UDP and send incoming traffic (from the TCP connection) to `10.0.1.1:30001` over UDP. Udp-bridge2 will listen on `10.0.0.2:40000` on UDP and send incoming traffic (from the TCP connection) to `10.0.0.2:40001` over UDP. Udp-bridge1 will now connect to udp-bridge2 via TCP to `10.0.0.2:40000` and both brdges are ready to tunnel traffic bidirectionally. 

The commandlines for both sites look like the following:
- Host1: `udp-bridge -c 10.0.0.2:40000 -d 10.0.0.1:30001 -u 10.0.0.1:30000` (Listen on UDP `10.0.0.1:30000`, send UDP to `10.0.0.1:30001`, connect to TCP `10.0.0.2:40000`)
- Host2: `udp-bridge -s 10.0.0.2:40000 -d 10.0.0.2:40001 -u 10.0.0.2:40000` (Listen on UDP `10.0.0.2:40000`, send UDP to `10.0.0.2:40001`, listen on TCP `10.0.0.2:40000`)

Application1 now only needs to send UDP traffic to `10.0.0.1:30000` to reach Application2 and Application2 needs to send UDP traffic to `10.0.0.2:40000` to reach Application1.

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
