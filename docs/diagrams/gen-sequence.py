import os
from plantuml import PlantUML

# Define your sequence diagram in PlantUML format
uml_code = """
@startuml
actor User
participant "Ea Frontend" as EA
participant "Ea API Gateway" as GATEWAY
participant "Job API" as JOB_API
participant "Agent Manager API" as AGENT_API
participant "Agent Operator" as AGENT_OPERATOR
participant "Ainu Manager API" as AINU_API
participant "Job Operator" as JOB_OPERATOR
participant "Job Executor" as JOB_EXECUTOR
database "ETCD" as ETCD
database "Agent Manager DB" as AGENT_DB
database "Ainu Engine DB" as AINU_DB

User -> EA : Requests job execution
EA -> GATEWAY : API job request passed into ea-platform
GATEWAY -> JOB_API : API job request to Job Engine API
JOB_API -> AGENT_API: Job API request with AgentID
AGENT_API -> AGENT_DB: Agent API looks up AgentID in DB
AGENT_DB -> AGENT_API: Response with AgentID details
AGENT_API -> JOB_API: Agent Manager response with details for AgentID

JOB_API -> ETCD: Create new AgentJob CR
ETCD -> JOB_OPERATOR: Operator sees new AgentJob with blank status
JOB_OPERATOR -> ETCD: Marks AgentJob as inactive
ETCD -> AGENT_OPERATOR: Operator sees inactive AgentJob
AGENT_OPERATOR -> AGENT_DB: Operator adds AgentJob to user's Job array as Inactive
ETCD -> JOB_OPERATOR: Operator sees inactive AgentJob
JOB_OPERATOR -> JOB_EXECUTOR : Execute AgentJob CR
ETCD -> AGENT_OPERATOR: Operator sees executing AgentJob
AGENT_OPERATOR -> AGENT_DB: Operator adds AgentJob to user's Job array as Active
JOB_OPERATOR -> ETCD: Marks AgentJob as executing
ETCD -> JOB_OPERATOR: Operator sees executor job is complete
JOB_OPERATOR -> ETCD: Marks AgentJob as complete
ETCD -> AGENT_OPERATOR: Operator sees complete AgentJob
AGENT_OPERATOR -> AGENT_DB: Operator adds AgentJob to user's Job array as Complete
ETCD -> JOB_OPERATOR: Operator sees completed AgentJob
JOB_OPERATOR -> ETCD: Cleans up completed AgentJob

' Periodic job status polling from frontend
loop Every few seconds while job is active
    EA -> GATEWAY: Request job status update
    GATEWAY -> AINU_API: Send job status udpate request to Ainu Manager
    AINU_API -> AINU_DB: Fetch latest job statuses
    AINU_DB -> AINU_API: Return updated job list
    AINU_API -> GATEWAY: Send updated job statuses
    GATEWAY -> EA: Sund updated job statuses
    EA -> User: Refresh job status in UI
end
@enduml
"""

# Save the PlantUML script
uml_file = "request_response.puml"
with open(uml_file, "w") as f:
    f.write(uml_code)

# Generate the diagram
plantuml = PlantUML(url="http://www.plantuml.com/plantuml/png/")
image_path = plantuml.processes_file(uml_file)

# Delete the .puml file after generation
if os.path.exists(uml_file):
    os.remove(uml_file)
    print(f"Deleted temporary file: {uml_file}")

print(f"Generated sequence diagram saved at: {image_path}")

