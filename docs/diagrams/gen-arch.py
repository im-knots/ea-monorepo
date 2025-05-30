from diagrams import Diagram, Cluster
from diagrams.k8s.compute import Pod, Job, Cronjob
from diagrams.k8s.infra import ETCD
from diagrams.k8s.others import CRD
from diagrams.generic.storage import Storage
from diagrams.generic.device import Mobile
from diagrams.gcp.network import LoadBalancing, DNS
from diagrams.gcp.security import Iam
from diagrams.gcp.storage import GCS
from diagrams.onprem.monitoring import Grafana

with Diagram("Eru Labs", show=False):
    dns = DNS("Public DNS")
    
    ainuClients = Mobile("Ainulindale Client Compute")
    ollama = Pod("Ollama Pods")

    # Brand www
    with Cluster("Brand WWW"):
        brandLB = LoadBalancing("Brand WWW LB")
        brandFrontend = Pod("Brand Frontend")
        brandBackend = Pod("Brand Backend")
        brandDB = Storage("Brand WWW DB")
        dns >> brandLB >> brandFrontend >> brandBackend >> brandDB

    # Ea www
    with Cluster("Ea WWW"):
        eaLB = LoadBalancing("Ea WWW LB")
        eaFrontend = Pod("Ea Frontend")
        eaAPIGateway = Pod("Ea API Gateway")
        dns >> eaLB >> eaFrontend >> eaAPIGateway

    # Ea Agent Engine
    with Cluster("Agent Engine"):
        eaAgentManager = Pod("Agent Manager API")
        eaAgentDB = Storage("Agent Manager DB")
        eaAPIGateway >> eaAgentManager >> eaAgentDB
    
    # Ea Job Engine
    with Cluster("Job Engine"):
        eaJobAPI = Pod("Job API")
        eaJobOperator = Pod("Job Operator")
        eaJobUtils = Pod("Ea Job Utilities")
        eaAgentJobCRD = CRD("AgentJob CRD")
        eaAgentJobETCD = ETCD("AgentJob CRs")
        eaJobExecutor = Job("Ea Job Executor")
        eaAPIGateway >> eaJobAPI
        eaAgentJobCRD >> eaJobAPI >> eaAgentJobETCD 
        eaJobOperator >> eaAgentJobETCD
        eaJobOperator >> eaJobExecutor >> ollama
        eaJobExecutor >> eaJobUtils
        eaJobExecutor >> eaAgentManager
        ollama >> eaJobExecutor
        eaJobAPI >> eaAgentManager

        # Future state
        # eaFrontend >> eaJobOrchestrator >> [eaJobInf, eaJobTrn, eaJobAgt] >> ainuClients



    # Ea Ainu Engine
    with Cluster("Ainulindale Engine"):
        eaAinuUserManager = Pod("Ainu Engine User Manager")
        eaAinuDB = Storage("Ainu Engine DB")
        eaAinuOperator = Pod("Ainu Operator")
        eaAPIGateway >> eaAinuUserManager >> eaAinuDB
        eaAinuOperator >> eaAgentJobETCD
        eaAinuOperator >> eaAinuDB
    
    # Ea User Engine
    # with Cluster("User Engine"):
    #     eaUser = Pod("User API")
    #     eaUserDB = Storage("User DB")
    #     authProvider = Iam("Some Auth Provider?")
    #     eaFrontend >> eaUser >> eaUserDB
    #     eaUser >> authProvider

    # Ea Game Engine
    # with Cluster("Game Engine"):
    #     eaGame = Pod("Game API")
    #     eaGameDB = Storage("Game DB")
    #     eaFrontend >> eaGame >> eaGameDB

    # Ea User Data Engine
    # with Cluster("Data Engine"):
    #     eaDataManager = Pod("User Data Manager API")
    #     userDataBuckets = GCS("User Data Buckets")
    #     eaFrontend - eaDataManager - userDataBuckets
    
    # Ea Commerce Engine
    # with Cluster("Commerce Engine"):
    #     eaMarketplace = Pod("Marketplace API")
    #     eaMarketplaceDB = Storage("Marketplace DB")
    #     eaCredits = Pod("Compute Credit API")
    #     eaFrontend >> eaMarketplace >> eaMarketplaceDB
    #     eaCredits - eaUser
    
    # Ea Analytics Engine
    # with Cluster("Analytics Engine"):
    #     eaDataAggregator = Pod("Data Aggregator API")
    #     eaDataGrafana = Grafana("Database Dashboards")
    #     eaDataGrafana << eaDataAggregator << [brandDB, eaUserDB, eaGameDB, eaMarketplaceDB, eaAgentDB]


        
        


