# TinyChat

A simple toy telnet chat server built in 5 hours

## Run Server

```git clone github.com/jaredfolkins/tinychat``` 

```$ ./run.sh```

## Env Variables

Set the log file location

```export TCLog="./"```

Set the port

```export TCPort="8091"```

Set the host

```export TCHost="localhost"```

See examples in ```run.sh```

## Connect Client

```telnet localhost 8091```

## Help

```
--|Help|-----------------------------------------------------------------------------------------

{no flag needed}
send a message to the room you are in
(example: hi freeze, i'm batman)

/help
prints this banner
(example: /help) 

/quit
quits the application
(example: /quit) 

/nick
sets your nickname
(example: /nick batman)

/room 
change chat room, only 1 room may be joined
(example: /room gotham)

/blast
blast a message to all connected clients 
(example: /blast the ice man cometh)

-------------------------------------------------------------------------------------------------
```

## Features

- [x] Connect a client to the server
- [x] Send a message to the server
- [x] The server relays messages to all connected clients, including a timestamp and name of the client sending the message

## Extra Features

- [x] Log messages to a file
- [x] Read in config from a config file for port, IP, and log file location
- [x] Support multiple channels or rooms
- [x] Support changing clients changing their names
- [ ] An HTTP API to post messages
- [ ] An HTTP API to query for messages

## Test Converage

coverage: 28.3% of statements
