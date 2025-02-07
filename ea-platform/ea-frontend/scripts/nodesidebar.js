const AGENT_MANAGER_URL = "http://agent-manager.ea.erulabs.local/api/v1";

// Toggle Sidebar
document.addEventListener("DOMContentLoaded", () => {
    const sidebar = document.getElementById("node-sidebar");
    const toggleButton = document.getElementById("toggle-node-sidebar");

    if (sidebar && toggleButton) {
        toggleButton.addEventListener("click", (event) => {
            event.stopPropagation(); // Prevent interference with other click events
            sidebar.classList.toggle("collapsed");
        });

        renderNodeTiles();
    } else {
        console.error("Sidebar or toggle button not found!");
    }
});

// Fetch All Nodes
const fetchNodes = async () => {
    try {
        const response = await fetch(`${AGENT_MANAGER_URL}/nodes`);
        if (!response.ok) throw new Error(`Error: ${response.status}`);
        return await response.json();
    } catch (error) {
        console.error("Error fetching nodes:", error);
        return [];
    }
};

// Fetch Node Details
const fetchNodeDetails = async (nodeId) => {
    try {
        const response = await fetch(`${AGENT_MANAGER_URL}/nodes/${nodeId}`);
        if (!response.ok) throw new Error(`Error: ${response.status}`);
        return await response.json();
    } catch (error) {
        console.error(`Error fetching details for node ${nodeId}:`, error);
        return null;
    }
};

// Render Node Tiles
const renderNodeTiles = async () => {
    const nodes = await fetchNodes();
    const container = document.getElementById("node-tiles-container");
    container.innerHTML = "";

    for (const node of nodes) {
        const nodeDetails = await fetchNodeDetails(node.id);
        if (!nodeDetails) continue;

        const tile = document.createElement("div");
        tile.classList.add("node-tile");

        tile.innerHTML = `
            <h6>${nodeDetails.name}</h6>
            <p>Type: ${nodeDetails.type}</p>
            <p>Creator: ${nodeDetails.creator}</p>
        `;

        // Expand Details on Click
        tile.addEventListener("click", () => {
            alert(JSON.stringify(nodeDetails, null, 2)); // Replace with modal if needed
        });

        container.appendChild(tile);
    }
};
