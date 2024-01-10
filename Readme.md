# Wireguard Gaming Server
This is a go application for instrumenting a wireguard server configured for joining several
machines in a vnet which mirrors a local network. It was written with the desire to have
a nicely dockerized setup for running a vpn server through which games can be played as if the
machines are on a local network.

## Config
The server is configured through environment variables:
 - INTERFACE_CONFIG_PATH: The path to the wireguard interfaces config. Defaults to /etc/wireguard/wg0.conf
 - SUBNET: The subnet the server used. The first ip is assigned to the server's interface. Defaults to 10.32.42.0/24


 ## Deploying
 The server is designed to be used inside of a docker container. There is an example docker container in the
 deploy folder, which combines this server with [wireguard-ui](https://github.com/ngoduykhanh/wireguard-ui/).
 The setup will need some customization:
  - Generate a random session secret for wireguard-ui and put it into the docker file
  - Either enable a portforwarding for port 5000 for wireguard-ui, or add it to your preferred reverse proxy (e.g. the excelent [traefik](https://traefik.io/traefik/)). Keep in mind that the networking for the container is shared with the wireguard container, so port forwardings have to be done on that container.
