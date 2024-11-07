# IP Address Changer

This project allows users to change their internet IP address by reconnecting to a dynamic IP provider. The client application listens for a hotkey (Ctrl + Shift + S) and sends a request to the server to initiate the reconnection process. The server performs the necessary commands to disconnect and reconnect the network interface, then retrieves the new IP address.

## Features

- Automatically reconnects the specified network interface.
- Displays notifications about the old and new IP addresses.
- Logs the process to a file.
- Supports dynamic IP changes from a variety of external IP services.

## Server-Side Instructions

### Building the Server
To build the server application, use the following command in the terminal:
```bash
GOOS=linux GOARCH=arm64 go build -o ip_changer_server /server/main.go
```

### Copy the executable file
Move the executable file of your ip_changer_server application to the /opt/ip_changer/ directory on your OpenWRT router. Make sure the directory exists, and if not, create it:
```bash
mkdir -p /opt/ip_changer/
scp /path/to/ip_changer_server root@<192.168.1.1>:/opt/ip_changer/

```

### Create the service
Create a service file /etc/init.d/ip_changer and add the following code to it:

```sh
#!/bin/sh /etc/rc.common

START=99  
STOP=10   

PROG_NAME="ip_changer"
PROG_PATH="/opt/ip_changer/ip_changer_server" 

start() {
    echo "Start $PROG_NAME..."
    $PROG_PATH &
}

stop() {
    echo "Stop $PROG_NAME..."
    kill $(pgrep -f $PROG_NAME)
}

restart() {
    stop
    start
}

boot() {
    start
}
```
### Make the file executable
Make sure the service file has execute permissions:
```bash
chmod +x /etc/init.d/ip_changer
```
And also the application itself
```bash
chmod +x /opt/ip_changer/ip_changer_server
```
### Add the service to autorun
Enable autorun for your service:

```bash
/etc/init.d/ip_changer enable
```

If the error `': No such file or directory.common` occurs, here is [solution](https://stackoverflow.com/questions/73799370/no-such-file-or-directory-common)
### Start the service
Start the service manually:

```bash
/etc/init.d/ip_changer start
```
Now your application will automatically start when the router boots up and you can control it with the `start`, `stop`, and `restart` commands.

## Client-Side Instructions

### Dependencies

Make sure you have the following Go packages installed:

```bash
go get github.com/gen2brain/beeep
go get golang.design/x/hotkey
```

### Building the Client
To build the client application, use the following command in the terminal:
```bash
go build -buildmode=exe -o ip-changer-client.exe /client/main.go 
```

### Use 
The hotkey to change the IP address is Ctrl + Shift + S.



