export interface Job {
    id: string;
    agent_id: string;
    created_time: string;
    last_active?: string;
    status: string;
  }
  
  export interface Agent {
    id: string;
    name: string;
    description?: string;
    nodes: { type: string }[];
    jobs: Job[];
  }
  