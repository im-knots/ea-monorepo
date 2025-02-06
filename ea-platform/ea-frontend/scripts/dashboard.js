// Global Variables
let devices = [];
let jobs = [];
let agents = [];

// API URLs
const AINU_MANAGER_URL = "http://localhost:8085/api/v1";
const AGENT_MANAGER_URL = "http://localhost:8083/api/v1";
const JOB_API_URL = "http://localhost:8084/api/v1";

// Load Sidebar
fetch('../html/sidebar.html')
    .then(response => response.text())
    .then(html => {
        document.getElementById('sidebar-container').innerHTML = html;

        // Dynamically load sidebar.css
        const sidebarStylesheet = document.createElement('link');
        sidebarStylesheet.rel = 'stylesheet';
        sidebarStylesheet.href = '../styles/sidebar.css';
        document.head.appendChild(sidebarStylesheet);

        // Load sidebar.js after sidebar is added
        const script = document.createElement('script');
        script.src = '../scripts/sidebar.js';
        document.body.appendChild(script);
    });

// Fetch First User ID
const fetchFirstUserId = async () => {
    try {
        console.log("Fetching users...");
        const response = await fetch(`${AINU_MANAGER_URL}/users`);

        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const users = await response.json();
        console.log("Users fetched:", users);

        if (!Array.isArray(users) || users.length === 0) {
            console.warn("No users found.");
            return null;
        }

        return users[0].id;
    } catch (error) {
        console.error("Error fetching users:", error);
        return null;
    }
};

const fetchAgents = async (userId) => {
    if (!userId) return [];
  
    try {
        console.log(`Fetching agents for user: ${userId}`);
        const response = await fetch(`${AGENT_MANAGER_URL}/agents`);
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
  
        const userData = await response.json();
        console.log("Full user data:", userData);  // Debugging log
  
        return userData.filter(agent => agent.creator === userId) || [];  // Filter agents by userId
    } catch (error) {
        console.error("Error fetching agents:", error);
        return [];
    }
};

const fetchAgentName = async (agentId) => {
    if (!agentId) return "Unknown";

    try {
        const response = await fetch(`${AGENT_MANAGER_URL}/agents/${agentId}`);
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

        const agentData = await response.json();
        return agentData.name || "Unknown"; // Return the agent's name
    } catch (error) {
        console.error(`Error fetching agent with ID ${agentId}:`, error);
        return "Unknown";
    }
};


const fetchDevices = async (userId) => {
  if (!userId) return [];

  try {
      console.log(`Fetching devices for user: ${userId}`);
      const response = await fetch(`${AINU_MANAGER_URL}/users/${userId}`);
      if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

      const userData = await response.json();
      console.log("Full user data:", userData);  // Debugging log

      return userData.compute_devices || [];  // Ensure correct field name
  } catch (error) {
      console.error("Error fetching devices:", error);
      return [];
  }
};

const fetchJobs = async (userId) => {
  if (!userId) return [];

  try {
      console.log(`Fetching jobs for user: ${userId}`);
      const response = await fetch(`${AINU_MANAGER_URL}/users/${userId}`);
      if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

      const userData = await response.json();
      console.log("Full user jobs:", userData);  // Debugging log

      return userData.jobs || [];  // Ensure correct field name
  } catch (error) {
      console.error("Error fetching jobs:", error);
      return [];
  }
};

// Fetch Stats and Graph Data
const fetchStatsAndGraphData = async (userId) => {
    if (!userId) return {};

    try {
        console.log(`Fetching stats for user: ${userId}`);
        const response = await fetch(`${AINU_MANAGER_URL}/users/${userId}`);

        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const userData = await response.json();
        console.log("User stats received:", userData);

        const totalComputeRate = userData.compute_devices?.reduce((sum, device) => sum + device.compute_rate, 0) || 0;

        return {
            totalComputeCredits: userData.compute_credits || 0,
            computeUsage: {
                labels: ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul"],
                data: [5, 10, 7, 12, 8, 150, 200]
            },
            computeRate: {
                labels: ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul"],
                data: [0, 0, 100, 40, 80, 600, totalComputeRate]
            }
        };
    } catch (error) {
        console.error("Error fetching stats and graph data:", error);
        return {};
    }
};

const deleteJob = async (jobId, userId) => {
    if (!jobId || !userId) {
        console.error("Missing jobId or userId for deletion");
        return;
    }

    try {
        console.log(`Deleting job with ID: ${jobId}`);
        const response = await fetch(`${AINU_MANAGER_URL}/users/${userId}/jobs/${jobId}`, {
            method: "DELETE",
        });

        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

        console.log(`Job ${jobId} deleted successfully`);
        
        // Refresh job table after deletion
        jobs = jobs.filter(job => job.id !== jobId);
        populateJobsTable();
    } catch (error) {
        console.error(`Error deleting job ${jobId}:`, error);
    }
};

const startAgent = async (agentId, userId) => {
    if (!agentId || !userId) {
        console.error("Missing agentId or userId for starting agent");
        return;
    }

    const requestBody = {
        agent_id: agentId,
        user_id: userId,
    };

    try {
        console.log(`Starting agent with ID: ${agentId}`);
        console.log("POST Request Body:", JSON.stringify(requestBody, null, 2));  // Pretty-printed request body

        const response = await fetch(`${JOB_API_URL}/jobs`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(requestBody),
        });

        const responseData = await response.json();  // Parse the JSON response

        console.log("API Response Status:", response.status);
        console.log("API Response Body:", JSON.stringify(responseData, null, 2));  // Pretty-printed response

        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

        console.log(`Agent ${agentId} started successfully`);
    } catch (error) {
        console.error(`Error starting agent ${agentId}:`, error);
    }
};


// Update Stats Panel
const updateStatsPanel = async (userId) => {
    const statsData = await fetchStatsAndGraphData(userId);

    document.getElementById("totalComputeCredits").textContent = statsData.totalComputeCredits || 0;

    const totalComputeRate = devices.reduce((sum, device) => sum + device.compute_rate, 0);
    document.getElementById("totalComputeRate").textContent = totalComputeRate;

    const inferenceJobsInProgress = jobs.filter(job => job.job_type === "Inference" && job.status === "Active").length;
    document.getElementById("inferenceJobsInProgress").textContent = inferenceJobsInProgress;

    const trainingJobsInProgress = jobs.filter(job => job.job_type === "Training" && job.status === "Active").length;
    document.getElementById("trainingJobsInProgress").textContent = trainingJobsInProgress;

    const agentJobsInProgress = jobs.filter(job => job.job_type === "Agent" && job.status === "Active").length;
    document.getElementById("agentJobsInProgress").textContent = agentJobsInProgress;

    const ctx = document.getElementById("computeUsageChart").getContext("2d");

    new Chart(ctx, {
        type: "line",
        data: {
            labels: statsData.computeUsage?.labels || [],
            datasets: [
                {
                    label: "Compute Usage",
                    data: statsData.computeUsage?.data || [],
                    borderColor: "rgba(255, 99, 132, 1)",
                    backgroundColor: "rgba(255, 99, 132, 0.2)",
                    borderWidth: 2
                },
                {
                    label: "Compute Rate",
                    data: statsData.computeRate?.data || [],
                    borderColor: "rgba(75, 192, 192, 1)",
                    backgroundColor: "rgba(75, 192, 192, 0.2)",
                    borderWidth: 2
                }
            ]
        },
        options: {
            responsive: true,
            plugins: {
                legend: {
                    display: true
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: "TFLOPS",
                        color: "#ffffff",
                        font: { size: 14 }
                    }
                }
            }
        }
    });
};

const populateDevicesTable = () => {
  const devicesTableBody = document.getElementById('devices-table-body');
  devicesTableBody.innerHTML = '';

  if (devices.length > 0) {
      devices.forEach((device) => {
          const statusClass = device.status === "Active" ? "active" :
                              device.status === "Offline" ? "offline" : "error";

          const row = `
              <tr>
                  <td>${device.device_name || "Unknown"}</td>
                  <td>${device.device_os || "Unknown"}</td>
                  <td>${device.compute_type || "Unknown"}</td>
                  <td>
                      <span class="status-indicator ${statusClass}"></span>${device.status || "Unknown"}
                  </td>
                  <td>${device.compute_rate || 0}</td>
                  <td>${device.id || "N/A"}</td>
                  <td>${device.last_active || "N/A"}</td>
              </tr>
          `;
          devicesTableBody.innerHTML += row;
      });
  } else {
      devicesTableBody.innerHTML = `
          <tr>
              <td colspan="5" class="text-center">No devices available. Please install Ainulindale on supported devices.</td>
          </tr>
      `;
  }
};


const populateJobsTable = async (userId) => {
    const jobsTableBody = document.getElementById('jobs-table-body');
    jobsTableBody.innerHTML = '';

    if (jobs.length > 0) {
        for (const job of jobs) {
            const statusClass = job.status === "New" ? "new" :
                                job.status === "Pending" ? "pending" :
                                job.status === "Error"  ? "error" :
                                job.status === "Complete" ? "complete" : "error";

            // Fetch the agent name asynchronously
            const agentName = await fetchAgentName(job.agent_id);

            // Determine if delete button should be enabled
            const isDeletable = job.status === "Complete" ? "" : "text-muted disabled";
            
            const row = `
                <tr>
                    <td>${job.job_name || "Unknown"}</td>
                    <td>${job.job_type || "Unknown"}</td>
                    <td>${job.agent_id || "Unknown"}</td>
                    <td>${agentName || "Unknown"}</td> 
                    <td><span class="status-indicator ${statusClass}"></span>${job.status || "Unknown"}</td>
                    <td>${job.last_active || "N/A"}</td>
                    <td>
                        <i class="bi bi-trash ${isDeletable}" 
                           role="button" 
                           data-job-id="${job.id}" 
                           title="Delete Job"></i>
                    </td>
                </tr>
            `;
            jobsTableBody.innerHTML += row;
        }

        // Attach click events after rendering table rows
        document.querySelectorAll(".bi-trash").forEach(icon => {
            if (!icon.classList.contains("text-muted")) {
                icon.addEventListener("click", () => {
                    const jobId = icon.getAttribute("data-job-id");
                    deleteJob(jobId, userId);
                });
            }
        });
    } else {
        jobsTableBody.innerHTML = `
            <tr>
                <td colspan="7" class="text-center">No compute jobs</td>
            </tr>
        `;
    }
};


const populateAgentsTable = async (userId) => {
    const agentsTableBody = document.getElementById("agents-table-body");
    agentsTableBody.innerHTML = "";

    if (agents.length > 0) {
        agents.forEach((agent) => {
            const row = `
                <tr>
                    <td>${agent.name || "Unknown"}</td>
                    <td>${agent.id || "Unknown"}</td>
                    <td>${agent.creator || "Unknown"}</td>
                    <td>
                        <button class="btn btn-dark btn-sm start-agent-btn" 
                            data-agent-id="${agent.id}" 
                            data-user-id="${userId}">
                            <i class="bi bi-play-circle"></i> Start
                        </button>
                    </td>
                </tr>
            `;
            agentsTableBody.innerHTML += row;
        });

        // Attach event listener correctly
        document.querySelectorAll(".start-agent-btn").forEach((button) => {
            button.addEventListener("click", (event) => {
                const agentId = event.currentTarget.getAttribute("data-agent-id");
                const userId = event.currentTarget.getAttribute("data-user-id");  // Ensure this exists
                startAgent(agentId, userId);
            });
        });
    } else {
        agentsTableBody.innerHTML = `
            <tr>
                <td colspan="4" class="text-center">No Agents Found. Visit the builder or marketplace to add some!</td>
            </tr>
        `;
    }
};

  


document.addEventListener('DOMContentLoaded', async () => {
  const userId = await fetchFirstUserId();
  if (!userId) {
      console.warn("No valid user ID found.");
      return;
  }

  devices = await fetchDevices(userId);
  jobs = await fetchJobs(userId);
  agents = await fetchAgents(userId);

  console.log("Devices Loaded:", devices);
  console.log("Jobs Loaded:", jobs);
  console.log("Agents Loaded:", agents);

  populateDevicesTable();
  await populateJobsTable(userId);
  populateAgentsTable(userId);
  updateStatsPanel(userId);
});

