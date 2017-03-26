# conduction

#### Description
Enable MQTT/Rest endpoints and topics to interact with other MQTT/Rest endpoints and topics. A configurable interface(?)

#### Todo
* Make a MQTT connector (pub/sub)
* Make first mqtt e2e
* Make a REST connector
* Add EventWorker.go
* Refactor to combine kafka and kafkaDdistributor...

#### Future potential problems..
* one MQTT or Rest entry point only (? ... should do multiple but that gets complex)
* message pb should have nextPoint and finalPoint instead of just endPoint
* we might have to make another "storage" package for storing more than one entry point
* if we have a distributed system, we need to fine a way to get messages back to the correct instance so the rest handler can return the message to user
* what lambda engine?
* none linear traveling event (aka event spwans two events which are expected to finish before firing final event)

#### Testing
To run test, install ginkgo and in root of project, run:
```bash
ginkgo -r
```

#### Generating from proto file
```bash
./protoc -I=/home/edmund/Workspace/go/src/github.com/edfungus/conduction/model --go_out=/home/edmund/Workspace/go/src/github.com/edfungus/conduction/model/ /home/edmund/Workspace/go/src/github.com/edfungus/conduction/model/message.proto 
```
