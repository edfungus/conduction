# conduction

#### Description
Event orchestrator

#### Todo
* Figure out how acks will work
* Replace main.go with transport interface
* Database and outbound interfaces
* Add EventWorker.go

#### Testing
To run test, install ginkgo and in root of project, run:
```bash
ginkgo -r
```

#### Generating from proto file
```bash
./protoc -I=/home/edmund/Workspace/go/src/github.com/edfungus/conduction/models/ --go_out=/home/edmund/Workspace/go/src/github.com/edfungus/conduction/models /home/edmund/Workspace/go/src/github.com/edfungus/conduction/models/message.proto 
```
