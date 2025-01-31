// Global Variables
let devices = [];
let jobs = [];

// API Base URL
const API_BASE_URL = "http://localhost:8085/api/v1";

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
        const response = await fetch(`${API_BASE_URL}/users`);

        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const users = await response.json();
        console.log("Users fetched:", users);

        if (!Array.isArray(users) || users.length === 0) {
            console.warn("No users found.");
            return null;
        }

        return users[0]._id;
    } catch (error) {
        console.error("Error fetching users:", error);
        return null;
    }
};

const fetchDevices = async (userId) => {
  if (!userId) return [];

  try {
      console.log(`Fetching devices for user: ${userId}`);
      const response = await fetch(`${API_BASE_URL}/users/${userId}`);
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
      const response = await fetch(`${API_BASE_URL}/users/${userId}`);
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
        const response = await fetch(`${API_BASE_URL}/users/${userId}`);

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

const populateJobsTable = () => {
  const jobsTableBody = document.getElementById('jobs-table-body');
  jobsTableBody.innerHTML = '';

  if (jobs.length > 0) {
      jobs.forEach((job) => {
          const statusClass = job.status === "Active" ? "active" :
                              job.status === "In Progress" ? "inprogress" :
                              job.status === "Offline" ? "offline" : "error";

          const row = `
              <tr>
                  <td>${job.job_name || "Unknown"}</td>
                  <td>${job.job_type || "Unknown"}</td>
                  <td>
                      <span class="status-indicator ${statusClass}"></span>${job.status || "Unknown"}
                  </td>
                  <td>${job.last_active || "N/A"}</td>
              </tr>
          `;
          jobsTableBody.innerHTML += row;
      });
  } else {
      jobsTableBody.innerHTML = `
          <tr>
              <td colspan="5" class="text-center">No compute jobs</td>
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

  console.log("Devices Loaded:", devices);
  console.log("Jobs Loaded:", jobs);

  populateDevicesTable();
  populateJobsTable();
  updateStatsPanel(userId);
});

