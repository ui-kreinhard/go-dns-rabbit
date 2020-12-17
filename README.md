What's this?
===
The idea is to provide a dns server which resolves dns queries by a hashmap which is updated by a heartbeat mechanism.
This means that client send heartbeats in a certain interval. These messages contain the hostname and the ip address. The server takes these heartbeats and updates the entries in it's hashmap.

When later a client sends a query to the server it'll resolve it by the filled hashmap. If no entry is found the server asks all clients if it's their hostname.

As a communication base rabbit is used for heartbeats and also for other communications like active querying

It's a hobby project and more as a proof of concept. It's not intended to be used in production.


How to use/run/whatever
===
Either compile it (never checked if it's working on another machine).

For using it you need 2 hosts which can connect to a rabbitmq instance

On server:
```
go run main.go amqp://guest:guest@localhost:5672/ server
``` 

On client
```
go run main.go amqp://guest:guest@localhost:5672/ client $interface
```

On the server side Try to resolve a dns name with dig
```
dig -p 5553 clientname.$interface @127.0.0.1
```

You can also run client and server on the same machine


