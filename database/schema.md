### Schema Tables
Tables for a relational database

Incoming requests are mapped to a path uuid which then can be mapped to a one or more flows
| paths |
| --- |
| id* (serial) |
| route (string) |
| type (string) |
| listen (boolean) |

Flows are essentially events/requests to be made
| flows | 
| --- |
| id* (serial) |
| name (string) |
| description (string) |
| wait (boolean) |
| path (serial FK to paths.id) |

Flows can have dependents which must be done before the parent flow. There could be multiple dependent flows also which their order matters.
| flowDependency |
| --- |
| parentFlowID* (serial) |
| position* (int) |
| dependentFlowID* (serial) |

The following is almost like application cache

Temporary incoming paths. This is for multi-step Flows or wait Flows
| temporaryPaths |
| --- |
| path (string) |
| type (string) |
| identity (string) |
| tempFlowID (serial) |
| payload ([]byte) | // This stores the payload of each dependent as they come in. This will also 


| temporaryFlows |
| --- |
| id (serial) |
| flow ([]byte) |