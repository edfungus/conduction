# conduction

#### Description
Event orchestrator

#### Todo
* Kafka cannot connect tests?
* Test RedisStorage
* Replace main.go with transport interface
* Add EventWorker.go
* Refactor to combine kafka and kafkaDdistributor...

#### Testing
To run test, install ginkgo and in root of project, run:
```bash
ginkgo -r
```

#### Generating from proto file
```bash
./protoc -I=/home/edmund/Workspace/go/src/github.com/edfungus/conduction/storage/redis --go_out=/home/edmund/Workspace/go/src/github.com/edfungus/conduction/storage/redis/ /home/edmund/Workspace/go/src/github.com/edfungus/conduction/storage/redis/value.proto 
```
