### Schema Tables
Tables for a relational database

Flows are essentially events/requests to be made
| Flows | 
| --- |
| flow* (uuid) |
| name (string) |
| description (string) |
| path (Path) |

Flows can have dependents which must be done before the parent flow. There could be multiple dependent flows also which their order matters.
| FlowDependents |
| --- |
| parentFlow* (uuid) |
| order* (int) |
| dependentFlow* (uuid) |

This maps the incoming requests (path) to a flow
| incomingPath2Flow |
| --- |
| path* (uuid) |
| flow (uuid) |

Incoming requests are mapped to a path uuid which then can be mapped to a one or more flows
| incomingPaths |
| --- |
| path* (uuid) |
| path (string) |
| type (string?) |

