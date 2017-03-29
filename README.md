# conduction

### Description
Event based Kafka router; second times the charm!

[Last commit of previous attempt](https://github.com/edfungus/conduction/tree/892cf01f4c0c2b669f69b1d6aa1077ce7e7bf66f)

### Examples ... aka thinking out loud
This is for me to get my thoughts together to make a better architecture. There are a fair number of features and requirements I want to support and it is more complicated than I anticipated.

#### mqtt topic to mqtt topic (trival)
1. mqtt --> mqtt connector

2. mqtt connector --> conduction

    MQTT connector gets message and makes this:
    ```
    Message {
        Path origin = {path=in, type=1}
        bytes payload
    }
    ```
3. conduction <--> database

    Using <origin.path>_<origin.type> as key in db, get []FlowIds:
    ```
    []FlowIds = ["some Flow uuid"]
    ```
    
    Using the Flow UUID, get a Flow which one may be:
    ```
    Flow {
        Path path = {path=out, type=1}
    }
    ```
4. conduction --> mqtt connector
    Using the Path given, create a new Message to pass back to the mqtt connector
     ```
    Message {
        Path origin = {path=out, type=1}
        bytes payload
    }
    ```
5. mqtt connector --> mqtt

#### rest call to rest call     
