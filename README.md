# conduction

### Description
Event based Kafka router; second times the charm!

[Last commit of previous attempt](https://github.com/edfungus/conduction/tree/892cf01f4c0c2b669f69b1d6aa1077ce7e7bf66f)

### Examples ... aka thinking out loud
This is for me to get my thoughts together to make a better architecture. There are a fair number of features and requirements I want to support and it is more complicated than I anticipated.

#### mqtt to mqtt (pass through, one step)
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

    Using <origin.path>_<origin.type>, get list of Flows to execute:
    ```
    []Flows = [
        Flow {
            Path path = {path=out, type=1}
        }
    ]
    ```
4. conduction --> mqtt connector

    Using the Path given, create a new Message to pass back to the mqtt connector
     ```
    Message {
        Path destination = {path=out, type=1}
        bytes payload
    }
    ```
5. mqtt connector --> mqtt

#### rest to rest (return call, two step)
1. rest --> rest connector

2. rest connector --> conduction

    Coverts call to Message. REST will always have a return Path to identify the controller:
    ```
    Message {
        Path origin = {path=GET_/home, type=0}
        Path return = {type=0, metadata.controllerID = 1}
    }
    ```
3. conduction <--> database

    Get the list of Flow to execute:
    ```
    []Flows = [
        Flow {
            Path path = {path=GET_/catpics, type=0}
        }
    ]
    ```

3. conduction --> rest connector

    Making the call to get a cat picture. Notice return get passed to the connector because returns are handled by the connector (perhaps not...might just insert into `incomingTempPaths` table):
    ```
    Message {
        Path destination = {path=GET_/catpics, type=1}
        Path return = {type=0, metadata.controllerID = 1}
    }
    ```
4. rest connector --> conduction

    Afer getting the cat piture, send back to conduction. If there was a return, the connector will place it in destination when completely its task:
    ```
    Message {
        Path origin = {path=GET_/catpics, type=1}
        Path destination = {type=0, metadata.controllerID = 1}
        bytes payload = `catpic`
    }
    ```

5. conduction <--> database

    Because there is an origin, conduction will check to see if this will launch a flow. Yes, infinite loops may be possible :(. In this case there are none
    ```
    []Flows = nil
    ```

6. conduction --> rest connector

    When are no Flows, conduction would usually stop, but since there is a destination, it will send the Message to the destination
    ```
    Message {
        Path destination = {type=0, metadata.controllerID = 1}
        bytes payload = `catpic`
    }

#### mqtt to rest to mqtt (two step, return call, two types)
1. mqtt --> mqtt connector
2. mqtt connector --> conduction
3. conduction <--> database
4. conduction --> rest connector
5. rest connector <--> rest
6. rest connector --> conduction 
7. conduction <--> database
8. conduction --> mqtt connector
9. mqtt connector --> mqtt

#### mqtt to 2 mqtt (multistep, dependents, two types)
This one is very similar to jst mqtt to mqtt
1. mqtt --> mqtt connector
2. mqtt connector --> conduction 
3. conduction <--> database
4. conduction --> mqtt connector (x2) 
5. mqtt connector -- mqtt (x2)

#### rest to mqtt to rest (two step, return call, waiting)
1. rest --> rest connector

2. rest connector --> conduction 
    Creates:
    ```
    Message {
        Path origin = {path=GET_/home, type=0}
        Path return = {type=0, metadata.controllerID = 1}
        bytes payload = 255        
    }
    ```
3. conduction <--> database
    ```
    []Flows = [
        Flow {
            Path path = {path=lightOn, type=1, 
                dependentFlows=Flow {
                    Path path = {path=turnOnLight, type=1}
                }, wait=true
            }
        },
        Flow {
            Path path = {type=0, metadata.controllerID = 1}
        }    
    ]
    ```
4. conduction --> database
    Because there is more than one Flow, find Flow we are going to do first and remove from the object. Check now that the next Flow's `wait` If it is true, put an entry to database with the rest of the Flow
    ```
    For `tempFLows`, insert `Flow` (the rest of it)
    For `incomingTempPaths`, insert `path`, `type`, `identity`(created by conduction), `tempFlow` (uuid)
    ```
5. conduction --> mqtt connector
    Note that identity is set.. but MQTT doesn't support identity...so what ever next message matches the expect return will continue to Flow
    ```
    Message {
        Path destination = {path=turnOnLight, type=0, identity=uuid}
        bytes payload = 255                
    }
    ```
6. mqtt connector --> mqtt
    Light sets to 255
7. mqtt --> mqtt connector
    Light sends back "DONE"
8. mqtt connector --> conduction 
    ```
    Message {
        Path origin = {path=lightOn, type=0}
        bytes payload = "DONE"                
    }
    ```
9. conduction <--> database
    Note, there is no path to flow relationship for this incoming mqtt message. The reason it will get routed is because conduction put in a temporary entry to wait for response. Because we can tell that there are no more entries with emtpy payload for the `tempFlow` uuid in `incomingTempPaths`, we know we can consolidate all the payload and pass them together to the parent Flow. I guess the combination will be JSON key-value
    ```
    incomingPaths --> []Flow = nil
    incomingTempPaths --> tempFlows --> []Flow = [
        Flow {
            Path path = {path=turnOnLight, type=1, wait=true}
        },
        Flow {
            Path path = {type=0, metadata.controllerID = 1}
        }
    ]
    ```
10. conduction --> rest controller
    Because the next Flows is `wait`=true, we don't actually execute. We usually consolidate all the payloads and send to next Flow which is REST but since one is `wait`=true, we onl take the payload of the `wait` Flow. Note, if the parent was a Lambda, we would send the consolidated JSON to the lambda and then send the output of the Lambda to REST. 
    `` 
    Message {
        Path destination = {type=0, metadata.controllerID = 1}
        bytes payload = "DONE"                        
    }
    ```
11. rest controller --> rest

#### mqtt to rest (x2) to mqtt
This triggers two rest calls which then merges the results together and sends to mqtt. Ideally, the parent flow of the two rest calls would be a lamdba call which can merge the results. This is a far off feature...

#### mqtt to mqtt or mqtt
The []Flows has no logic and will jsut run the next one, but what if you want logic? Should that be imbedded in the Flow object or fired via lambda??

#### What happens if we get errors?? :O
