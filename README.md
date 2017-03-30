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

    Making the call to get a cat picture. Notice return get passed to the connector because returns are handled by the connector.:
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

#### mqtt to 2 mqtt (multistep, dependents, two types)

#### rest to mqtt to rest (two step, return call, waiting)

#### What happens if we get errors?? :O
