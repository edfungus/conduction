# conduction

#### Description
Enable MQTT/Rest endpoints and topics to interact with other MQTT/Rest endpoints and topics. A configurable interface(?)

#### Todo
* Make a MQTT connector (pub/sub)
* Make first mqtt e2e
* Make a REST connector
* Add EventWorker.go
* Refactor to combine kafka and kafkaDdistributor...

#### Testing
To run test, install ginkgo and in root of project, run:
```bash
ginkgo -r
```

#### Generating from proto file
```bash
./protoc -I=/home/edmund/Workspace/go/src/github.com/edfungus/conduction/model --go_out=/home/edmund/Workspace/go/src/github.com/edfungus/conduction/model/ /home/edmund/Workspace/go/src/github.com/edfungus/conduction/model/message.proto 
```
