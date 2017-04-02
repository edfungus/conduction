CREATE DATABASE IF NOT EXISTS conduction;

SET DATABASE = conduction;

GRANT ALL ON DATABASE conduction TO conductor;

CREATE TABLE IF NOT EXISTS paths (
    id SERIAL PRIMARY KEY, 
    route STRING NOT NULL, 
    type STRING NOT NULL,
    listen BOOL NOT NULL,
    UNIQUE (route, type)
); 

CREATE TABLE IF NOT EXISTS flows (
    "path" INT,    
    id SERIAL, 
    name STRING NOT NULL, 
    description STRING, 
    wait BOOL NOT NULL,
    PRIMARY KEY ("path", id),
    CONSTRAINT fk_pathId FOREIGN KEY ("path") REFERENCES paths
    ) INTERLEAVE IN PARENT paths ("path")
;

CREATE TABLE IF NOT EXISTS flow_dependency (
    parent_path INT,
    parent_flow INT,
    position INT NOT NULL,
    dependent_path INT NOT NULL,
    dependent_flow INT NOT NULL,
    PRIMARY KEY (parent_path, parent_flow, position),
    CONSTRAINT fk_parent_flow FOREIGN KEY (parent_path, parent_flow) REFERENCES conduction.flows,
    CONSTRAINT fk_dependent_flow FOREIGN KEY (dependent_path, dependent_flow) REFERENCES conduction.flows
    ) INTERLEAVE IN PARENT flows (parent_path, parent_flow)
;
