// API URLs
const AINU_MANAGER_URL = "http://ainu-manager.ea.erulabs.local/api/v1";
const AGENT_MANAGER_URL = "http://agent-manager.ea.erulabs.local/api/v1";
const JOB_API_URL = "http://job-api.ea.erulabs.local/api/v1";

// Load Sidebar
fetch('../html/sidebar.html')
    .then(response => response.text())
    .then(html => {
        document.getElementById('sidebar-container').innerHTML = html;
        const sidebarStylesheet = document.createElement('link');
        sidebarStylesheet.rel = 'stylesheet';
        sidebarStylesheet.href = '../styles/sidebar.css';
        document.head.appendChild(sidebarStylesheet);
        const script = document.createElement('script');
        script.src = '../scripts/sidebar.js';
        document.body.appendChild(script);
    });

// Fetch first user ID
const fetchFirstUserId = async () => {
    try {
        const response = await fetch(`${AINU_MANAGER_URL}/users`);
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        const users = await response.json();
        return users.length > 0 ? users[0].id : null;
    } catch (error) {
        console.error("Error fetching user ID:", error);
        return null;
    }
};

// Fetch agents belonging to the user
const fetchAgents = async (userId) => {
    try {
        const response = await fetch(`${AGENT_MANAGER_URL}/agents`);
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        const agents = await response.json();
        return agents.filter(agent => agent.creator === userId);
    } catch (error) {
        console.error("Error fetching agents:", error);
        return [];
    }
};

// Fetch detailed agent data
const fetchAgentDetails = async (agentId) => {
    try {
        const response = await fetch(`${AGENT_MANAGER_URL}/agents/${agentId}`);
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        return await response.json();
    } catch (error) {
        console.error(`Error fetching details for agent ${agentId}:`, error);
        return null;
    }
};

// Delete agent function
const deleteAgent = async (agentId) => {
    try {
        const response = await fetch(`${AGENT_MANAGER_URL}/agents/${agentId}`, {
            method: "DELETE"
        });

        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        console.log(`Agent ${agentId} deleted successfully`);

        // Refresh the agent list after deletion
        populateAgentsTable();
    } catch (error) {
        console.error(`Error deleting agent ${agentId}:`, error);
    }
};

const fetchJobs = async (agentId, userId) => {
    if (!userId) return [];
    try {
        console.log(`Fetching jobs for user: ${userId}`);
        const response = await fetch(`${AINU_MANAGER_URL}/users/${userId}`);
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

        const userData = await response.json();
        console.log("Full user jobs:", userData);  // Debugging log

        // Filter jobs based on the provided agentId
        const agentJobs = userData.jobs?.filter(job => job.agent_id === agentId) || [];

        return agentJobs;
    } catch (error) {
        console.error("Error fetching jobs:", error);
        return [];
    }
};

// Helper function to create the JSON editor
const createAgentEditor = (agentDetails) => {
    return `
        <pre class="json-editor" style="
            background-color:rgb(0, 0, 0); 
            color:rgb(85, 255, 0); 
            padding: 10px; 
            border-radius: 5px; 
            font-family: 'Courier New', monospace; 
            white-space: pre-wrap; 
            word-wrap: break-word; 
            overflow-y: auto;
            margin: 0;
        ">
${JSON.stringify(agentDetails, null, 4)}
        </pre>
    `;
};


// Start agent function
const startAgent = async (agentId, userId) => {
    if (!agentId || !userId) {
        console.error("Missing agentId or userId");
        return;
    }

    const requestBody = { agent_id: agentId, user_id: userId };

    try {
        const response = await fetch(`${JOB_API_URL}/jobs`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(requestBody),
        });

        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

        console.log(`Agent ${agentId} started successfully`);
    } catch (error) {
        console.error(`Error starting agent ${agentId}:`, error);
    }
};
  

// Populate Agents in Grid View
const populateAgentsGrid = async () => {
    const userId = await fetchFirstUserId();
    if (!userId) {
        console.warn("No valid user ID found.");
        return;
    }

    const agents = await fetchAgents(userId);
    const agentsContainer = document.getElementById("agents-container");
    agentsContainer.innerHTML = "";

    if (agents.length === 0) {
        agentsContainer.innerHTML = `<p class="text-center">No agents found. Visit the marketplace to add some.</p>`;
        return;
    }

    for (const agent of agents) {
        const details = await fetchAgentDetails(agent.id);
        const jobDetails = await fetchJobs(agent.id, userId)
        const nodeList = details.nodes.map(node => `<li>${node.type}</li>`).join('');

        const tile = document.createElement("div");
        tile.classList.add("col-lg-4", "col-md-6", "d-flex");

        tile.innerHTML = `
            <div class="agent-tile p-3 d-flex flex-column justify-content-between w-100">
                <h5 class="text-center">${agent.name}</h5>
                <p class="text-muted">${details?.description || "No description available."}</p>
                <p><strong>Nodes:</strong> ${details.nodes.length}</p>
                <p><strong>Jobs:</strong> ${jobDetails.length}</p>
                <ul class="node-list">${nodeList}</ul>
                <div class="d-flex gap-2">
                    <button class="btn btn-sm btn-success w-50 start-agent-btn" data-agent-id="${agent.id}" data-user-id="${userId}">
                        <i class="bi bi-play-circle"></i> Start
                    </button>
                    <button class="btn btn-sm btn-primary w-50 modify-agent-btn" data-agent-id="${agent.id}" data-user-id="${userId}">
                        <i class="bi bi-pencil-square"></i> Modify
                    </button>
                    <button class="btn btn-sm btn-danger w-50 delete-agent-btn" data-agent-id="${agent.id}">
                        <i class="bi bi-trash"></i> Delete
                    </button>
                </div>
            </div>
        `;

        agentsContainer.appendChild(tile);
    }

    // Attach event listeners for "Start Agent" buttons
    document.querySelectorAll(".start-agent-btn").forEach(button => {
        button.addEventListener("click", (event) => {
            const agentId = event.currentTarget.getAttribute("data-agent-id");
            const userId = event.currentTarget.getAttribute("data-user-id");
            startAgent(agentId, userId);
        });
    });

    // Attach event listeners for "Modify Agent" buttons
    document.querySelectorAll(".modify-agent-btn").forEach(button => {
        button.addEventListener("click", (event) => {
            const agentId = event.currentTarget.getAttribute("data-agent-id");
            const userId = event.currentTarget.getAttribute("data-user-id");
            modifyAgent(agentId, userId);
        });
    });

    // Attach event listeners for "Delete Agent" buttons
    document.querySelectorAll(".delete-agent-btn").forEach(button => {
        button.addEventListener("click", (event) => {
            const agentId = event.currentTarget.getAttribute("data-agent-id");
            const userId = event.currentTarget.getAttribute("data-user-id");
            deleteAgent(agentId, userId);
        });
    });
    


};

// Populate Agents in List View (Collapsible Rows)
const populateAgentsTable = async () => {
    const userId = await fetchFirstUserId();
    const agents = await fetchAgents(userId);
    const agentsContainer = document.getElementById("agents-container");
    agentsContainer.innerHTML = `
        <table class="table table-dark table-hover">
            <thead>
                <tr>
                    <th>Agent Name</th>
                    <th>Nodes</th>
                    <th>Jobs</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody id="agents-table-body"></tbody>
        </table>
    `;

    const tableBody = document.getElementById("agents-table-body");

    for (const agent of agents) {
        const details = await fetchAgentDetails(agent.id);
        const nodeList = details.nodes.map(node => `<li>${node.type}</li>`).join('');
        const jobDetails = await fetchJobs(agent.id, userId);

        const row = document.createElement("tr");
        row.classList.add("clickable-row");
        row.innerHTML = `
            <td>${agent.name}</td>
            <td>${details.nodes.length}</td>
            <td>${jobDetails.length}</td>
            <td>
                <div class="d-flex gap-2">
                    <button class="btn btn-sm btn-success w-50 start-agent-btn" data-agent-id="${agent.id}" data-user-id="${userId}">
                        <i class="bi bi-play-circle"></i> Start
                    </button>
                    <button class="btn btn-sm btn-primary w-50 modify-agent-btn" data-agent-id="${agent.id}" data-user-id="${userId}">
                        <i class="bi bi-pencil-square"></i> Modify
                    </button>
                    <button class="btn btn-sm btn-danger w-50 delete-agent-btn" data-agent-id="${agent.id}">
                        <i class="bi bi-trash"></i> Delete
                    </button>
                </div>
            </td>
        `;
        
        const jobRows = jobDetails.map(job => {
            const statusClass = job.status === "New" ? "new" :
                                job.status === "Pending" ? "pending" :
                                job.status === "Error" ? "error" :
                                job.status === "Complete" ? "complete" : "unknown";
        
            return `
                <tr>
                    <td>${job.agent_id}</td>
                    <td>${job.created_time}</td>
                    <td>${job.last_active || "N/A"}</td>
                    <td>
                        <span class="status-indicator ${statusClass}"></span>
                        ${job.status || "Unknown"}
                    </td>
                </tr>
            `;
        }).join('');

        const detailsRow = document.createElement("tr");
        detailsRow.classList.add("details-row");
        detailsRow.style.display = "none";
        detailsRow.innerHTML = `
            <td colspan="2">
                <strong>Description:</strong> ${details?.description || "No description available."}<br>
                <strong>Nodes:</strong> ${details.nodes.length}
                <ul class="node-list">${nodeList}</ul>

                <h6 class="mt-3">Jobs</h6>
                <table class="table table-sm table-bordered table-dark">
                    <thead>
                        <tr>
                            <th>Agent ID</th>
                            <th>Created Time</th>
                            <th>Last Active</th>
                            <th>Status</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${jobRows || '<tr><td colspan="3" class="text-center">No jobs found for this agent.</td></tr>'}
                    </tbody>
                </table>
            </td>
            <td colspan="2">
                ${createAgentEditor(details)}
            </td>
        `;

        tableBody.appendChild(row);
        tableBody.appendChild(detailsRow);

        // Toggle details on row click
        row.addEventListener("click", (event) => {
            if (
                !event.target.classList.contains("start-agent-btn") &&
                !event.target.classList.contains("modify-agent-btn") &&
                !event.target.classList.contains("delete-agent-btn")
            ) {
                const isVisible = detailsRow.style.display === "table-row";
                detailsRow.style.display = isVisible ? "none" : "table-row";

                // Adjust JSON editor height dynamically
                if (!isVisible) {
                    const jsonEditor = detailsRow.querySelector(".json-editor");
                    const parentHeight = detailsRow.clientHeight;
                    jsonEditor.style.maxHeight = `${parentHeight}px`;
                }
            }
        });

        // Start agent button event listener
        row.querySelector(".start-agent-btn").addEventListener("click", (event) => {
            event.stopPropagation(); // Prevent triggering the row click event
            startAgent(agent.id, userId);
        });

        // Modify agent button event listener
        row.querySelector(".modify-agent-btn").addEventListener("click", (event) => {
            event.stopPropagation(); // Prevent triggering the row click event
            modifyAgent(agent.id, userId);
        });

        // Delete agent button event listener
        row.querySelector(".delete-agent-btn").addEventListener("click", (event) => {
            event.stopPropagation(); // Prevent triggering the row click event
            deleteAgent(agent.id, userId);
        });
    }
};



// Initialize on page load
document.addEventListener('DOMContentLoaded', populateAgentsTable);

// Toggle between Grid and List Views
document.addEventListener('DOMContentLoaded', () => {
    const agentsContainer = document.getElementById('agents-container');

    document.getElementById('grid-view').addEventListener('click', () => {
        agentsContainer.classList.remove('list-group');
        agentsContainer.classList.add('carousel-container');
        populateAgentsGrid();
    });

    document.getElementById('list-view').addEventListener('click', () => {
        agentsContainer.classList.remove('carousel-container');
        agentsContainer.classList.add('list-group');
        populateAgentsTable();
    });
});
