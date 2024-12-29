// Global Devices Array
let devices = [];

// Load Sidebar
fetch('../html/sidebar.html')
    .then(response => response.text())
    .then(html => {
    document.getElementById('sidebar-container').innerHTML = html;

    // Dynamically load sidebar.css
    const sidebarStylesheet = document.createElement('link');
    sidebarStylesheet.rel = 'stylesheet';
    sidebarStylesheet.href = '../styles/sidebar.css'; // Path to your sidebar.css
    document.head.appendChild(sidebarStylesheet);

    // Load the sidebar.js script after the sidebar is added
    const script = document.createElement('script');
    script.src = '../scripts/sidebar.js'; // Path to your sidebar.js
    document.body.appendChild(script);
});

// Mock Call for Devices
const fetchDevices = async () => {
  return [
    { name: "Knot's Samsung Galaxy S21", os: "Linux (Android)", type: "CPU", status: "Active", computeRate: 5, id: "07627d25-a263-4c93-8a13-30252f7f19a6", lastActive: "2024-12-29 10:00:00 ET" },
    { name: "Athena", os: "Linux (Ubuntu)", type: "CPU + GPU", status: "Active", computeRate: 85, id: "e6f4bcae-4f4a-4459-a154-0af50a045f99", lastActive: "2024-12-29 10:00:00 ET" },
    { name: "SPR-1135UH", os: "MacOs", type: "GPU", status: "Error", computeRate: 0, id: "3d6e3664-6dad-456d-ba8b-4ca58a6ae49d", lastActive: "2024-12-29 10:00:00 ET" },
    { name: "pi-lab", os: "Linux (Raspberry Pi OS)", type: "CPU", status: "Offline", computeRate: 0, id: "bcf5283d-3c2e-43e7-8943-269724549d8c", lastActive: "2024-12-29 10:00:00 ET" },
    { name: "Knot's IPhone 15 Pro", os: "iOS", type: "CPU", status: "Active", computeRate: 10, id: "356f4cdc-37e3-4efd-90b5-4ab2b327f040", lastActive: "2024-12-29 10:00:00 ET" },
    { name: "Knot's Samsung Galaxy Tab S10 Ultra", os: "Linux (Android)", type: "CPU", status: "Offline", computeRate: 0, id: "57d63e95-3198-4b3f-9ab7-e40d60f4b72c", lastActive: "2024-12-29 10:00:00 ET" },
    { name: "work-laptop-001", os: "Windows", type: "CPU", status: "Error", computeRate: 0, id: "a31bd4d3-c5ad-48d0-a51b-260b3c779de7", lastActive: "2024-12-29 10:00:00 ET" },
    { name: "gcp-instance-10", os: "Linux (Fedora)", type: "TPU", status: "Active", computeRate: 30, id: "a31bd4d3-c5ad-48d0-a51b-260b3c779de7", lastActive: "2024-12-29 10:00:00 ET" }
  ];
};

// Mock API Call for Stats and Graph Data
const fetchStatsAndGraphData = async () => {
  // Calculate the total compute rate dynamically
  const totalComputeRate = devices.reduce((sum, device) => sum + device.computeRate, 0);

  // Simulated API response
  return {
    totalComputeCredits: 1250, // Total credits
    computeUsage: {
      labels: ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul"], // Example months
      data: [5, 10, 7, 12, 8, 15, 20], // Example usage data
    },
    computeRate: {
      labels: ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul"], // Example months
      data: [0, 0, 20, 40, 80, 20, totalComputeRate], // Inject the calculated rate
    },
  };
};

// Update Stats Panel
const updateStatsPanel = async () => {
  const statsData = await fetchStatsAndGraphData();

  // Update the total compute credits stat
  const totalCreditsElement = document.getElementById("totalComputeCredits");
  totalCreditsElement.textContent = statsData.totalComputeCredits;

  // Calculate and update the total compute rate dynamically
  const totalComputeRate = devices.reduce((sum, device) => sum + device.computeRate, 0);
  const totalComputeElement = document.getElementById("totalComputeRate");
  totalComputeElement.textContent = totalComputeRate;

  // Update the graph
  const ctx = document.getElementById("computeUsageChart").getContext("2d");

  new Chart(ctx, {
    type: "line",
    data: {
      labels: statsData.computeUsage.labels,
      datasets: [
        {
          label: "Compute Usage",
          data: statsData.computeUsage.data,
          borderColor: "rgba(255, 99, 132, 1)", // Red line
          backgroundColor: "rgba(255, 99, 132, 0.2)",
          borderWidth: 2,
        },
        {
          label: "Compute Rate",
          data: statsData.computeRate.data,
          borderColor: "rgba(75, 192, 192, 1)", // Green line
          backgroundColor: "rgba(75, 192, 192, 0.2)",
          borderWidth: 2,
        },
      ],
    },
    options: {
      responsive: true,
      plugins: {
        legend: {
          display: true, // Show legend for both lines
        },
      },
      scales: {
        y: {
            beginAtZero: true,
            title: {
                display: true,
                text: "TFLOPS", // Label for the Y-axis
                color: "#ffffff", // Optional: Set the label color (white for dark theme)
                font: {
                size: 14, // Optional: Adjust font size
                },
            },
        },
      },
    },
  });
};

// Initialize Stats and Devices on Page Load
document.addEventListener('DOMContentLoaded', async () => {
  devices = await fetchDevices();
  populateDevicesTable();
  updateStatsPanel();
});

// Populate Devices Table
const populateDevicesTable = () => {
  const devicesTableBody = document.getElementById('devices-table-body');
  devicesTableBody.innerHTML = ''; // Clear existing rows

  if (devices.length > 0) {
    devices.forEach((device) => {
      const statusClass =
        device.status === "Active"
          ? "active"
          : device.status === "Offline"
          ? "offline"
          : "error";

      const row = `
        <tr>
          <td>${device.name}</td>
          <td>${device.os}</td>
          <td>${device.type}</td>
          <td>
            <span class="status-indicator ${statusClass}"></span>${device.status}
          </td>
          <td>${device.id}</td>
          <td>${device.lastActive}</td>
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
