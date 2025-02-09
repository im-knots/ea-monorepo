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

// Delay helper function
const delay = (ms) => new Promise(resolve => setTimeout(resolve, ms));

// Track active refresh intervals for each agent
const jobRefreshIntervals = {};

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
    console.log("Fetching agents for userId:", userId); // Add this line
    try {
        const response = await fetch(`${AGENT_MANAGER_URL}/agents?creator_id=${userId}`);
        console.log("Request URL:", response.url); // Log the actual URL being called
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        const agents = await response.json();
        console.log("Fetched agents:", agents); // Log the response for inspection
        return agents;
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

// Helper function to create the JSON editor
const createJsonWindow = (agentDetails) => {
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

// Delete agent function
const deleteAgent = async (agentId) => {
    try {
        const response = await fetch(`${AGENT_MANAGER_URL}/agents/${agentId}`, {
            method: "DELETE"
        });

        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        console.log(`Agent ${agentId} deleted successfully`);
    } catch (error) {
        console.error(`Error deleting agent ${agentId}:`, error);
    }
};

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

function saveJson() {
    const jsonEditor = document.getElementById("jsonEditor");
    try {
        const jsonData = JSON.parse(jsonEditor.value);
        console.log("JSON saved:", jsonData);
        alert("JSON saved successfully!");
    } catch (error) {
        alert("Invalid JSON format!");
        console.error("Invalid JSON:", error);
    }
}

// Fetch all nodes created by the user
async function getNodes(userId) {
    try {
        const response = await fetch(`${AGENT_MANAGER_URL}/nodes?creator_id=${userId}`);
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        
        const nodes = await response.json();
        console.log("Fetched nodes:", nodes);

        // Fetch detailed info for each node
        const detailedNodes = await fetchNodeDetails(nodes);
        populateNodeGrid(detailedNodes);
    } catch (error) {
        console.error("Error fetching nodes:", error);
    }
}

// Fetch detailed node information for each node
async function fetchNodeDetails(nodes) {
    const detailedNodes = await Promise.all(
        nodes.map(async (node) => {
            try {
                const response = await fetch(`${AGENT_MANAGER_URL}/nodes/${node.id}`);
                if (!response.ok) throw new Error(`Failed to fetch details for node ID: ${node.id}`);
                const details = await response.json();
                return details;
            } catch (error) {
                console.error(`Error fetching node details for ID ${node.id}:`, error);
                return node;  // Fallback to the basic node data if details fail
            }
        })
    );

    return detailedNodes;
}

// Populate the Node Grid with detailed node information
function populateNodeGrid(nodes) {
    const nodeGrid = document.getElementById("nodeGrid");
    nodeGrid.innerHTML = "";  // Clear previous nodes

    nodes.forEach(node => {
        const nodeTile = document.createElement("div");
        nodeTile.className = "col-md-3 node-tile";
        nodeTile.innerHTML = `
            <h5>${node.name || "Unnamed Node"}</h5>
            <p>${node.type || "No node type address found."}</p>
            <p>${node.metadata.description || "No description available."}</p>
        `;

        nodeTile.onclick = () => {
            alert(`Node Selected: ${node.name || node.type}`);
            // You can trigger additional functionality here
        };

        nodeGrid.appendChild(nodeTile);
    });
}

document.addEventListener('DOMContentLoaded', async () => {
    const userId = await fetchFirstUserId();
    if (userId) {
        getNodes(userId);
    }
});

// Select the SVG canvas
const svgCanvas = document.getElementById('svgCanvas');

// Function to add a node to the SVG canvas
function addNodeToCanvas(node) {
    const nodeGroup = document.createElementNS("http://www.w3.org/2000/svg", "g");
    const rect = document.createElementNS("http://www.w3.org/2000/svg", "rect");
    const nameText = document.createElementNS("http://www.w3.org/2000/svg", "text");
    const typeText = document.createElementNS("http://www.w3.org/2000/svg", "text");

    let posX = 50 + Math.random() * 200;
    let posY = 50 + Math.random() * 200;

    const baseHeight = 80;
    const paramHeight = (node.parameters?.length || 0) * 70;  // Increased parameter spacing
    const totalHeight = baseHeight + paramHeight + 30;        // Additional padding

    rect.setAttribute("x", posX);
    rect.setAttribute("y", posY);
    rect.setAttribute("width", 240);                          // Slightly wider for content
    rect.setAttribute("height", totalHeight);
    rect.setAttribute("rx", 10);
    rect.setAttribute("ry", 10);
    rect.setAttribute("fill", "#2d2d2d");
    rect.setAttribute("stroke", "#ffffff");                   // Changed to white border
    rect.setAttribute("stroke-width", "2");
    rect.style.cursor = "move";

    nameText.setAttribute("x", posX + 120);
    nameText.setAttribute("y", posY + 25);
    nameText.setAttribute("fill", "#ffffff");
    nameText.setAttribute("font-size", "14px");
    nameText.setAttribute("text-anchor", "middle");
    nameText.textContent = node.name || "Unnamed Node";

    typeText.setAttribute("x", posX + 120);
    typeText.setAttribute("y", posY + 45);
    typeText.setAttribute("fill", "#aaaaaa");
    typeText.setAttribute("font-size", "12px");
    typeText.setAttribute("text-anchor", "middle");
    typeText.textContent = node.type || "Unknown Type";

    nodeGroup.appendChild(rect);
    nodeGroup.appendChild(nameText);
    nodeGroup.appendChild(typeText);

    if (node.parameters) {
        node.parameters.forEach((param, index) => {
            const label = document.createElementNS("http://www.w3.org/2000/svg", "text");
            label.setAttribute("x", posX + 10);
            label.setAttribute("y", posY + 75 + (index * 70));
            label.setAttribute("fill", "#ffffff");
            label.setAttribute("font-size", "12px");
            label.textContent = param.key;
            nodeGroup.appendChild(label);

            const foreignObject = document.createElementNS("http://www.w3.org/2000/svg", "foreignObject");
            foreignObject.setAttribute("x", posX + 10);
            foreignObject.setAttribute("y", posY + 85 + (index * 70));  // Increased spacing
            foreignObject.setAttribute("width", 220);
            foreignObject.setAttribute("height", 40);

            if (param.type === "bool") {
                const toggleContainer = document.createElement("label");
                toggleContainer.classList.add("switch");

                const input = document.createElement("input");
                input.setAttribute("type", "checkbox");
                input.checked = param.default || false;

                const slider = document.createElement("span");
                slider.classList.add("slider");

                toggleContainer.appendChild(input);
                toggleContainer.appendChild(slider);
                foreignObject.appendChild(toggleContainer);
            } else if (param.enum) {
                const select = document.createElement("select");
                select.classList.add("svg-select");

                param.enum.forEach(optionValue => {
                    const option = document.createElement("option");
                    option.value = optionValue;
                    option.textContent = optionValue;
                    if (optionValue === param.default) {
                        option.selected = true;
                    }
                    select.appendChild(option);
                });

                foreignObject.appendChild(select);
            } else {
                const input = document.createElement("input");
                input.setAttribute("type", "text");
                input.setAttribute("value", param.default || "");
                input.classList.add("svg-input");
                foreignObject.appendChild(input);
            }

            nodeGroup.appendChild(foreignObject);
        });
    }

    svgCanvas.appendChild(nodeGroup);

    // Dragging functionality
    let isDragging = false;
    let offsetX, offsetY;

    nodeGroup.addEventListener("mousedown", (e) => {
        if (e.target.tagName !== "INPUT" && e.target.tagName !== "SELECT" && e.target.tagName !== "LABEL") {
            isDragging = true;
            const rectBounds = nodeGroup.getBoundingClientRect();
            offsetX = e.clientX - rectBounds.left;
            offsetY = e.clientY - rectBounds.top;
        }
    });

    window.addEventListener("mousemove", (e) => {
        if (isDragging) {
            const mouseX = e.clientX - svgCanvas.getBoundingClientRect().left;
            const mouseY = e.clientY - svgCanvas.getBoundingClientRect().top;

            rect.setAttribute("x", mouseX - offsetX);
            rect.setAttribute("y", mouseY - offsetY);
            nameText.setAttribute("x", mouseX - offsetX + 120);
            nameText.setAttribute("y", mouseY - offsetY + 25);
            typeText.setAttribute("x", mouseX - offsetX + 120);
            typeText.setAttribute("y", mouseY - offsetY + 45);

            nodeGroup.querySelectorAll("foreignObject").forEach((fo, idx) => {
                fo.setAttribute("x", mouseX - offsetX + 10);
                fo.setAttribute("y", mouseY - offsetY + 85 + (idx * 70));
            });
            nodeGroup.querySelectorAll("text").forEach((textEl, idx) => {
                if (idx > 1) {
                    textEl.setAttribute("x", mouseX - offsetX + 10);
                    textEl.setAttribute("y", mouseY - offsetY + 75 + ((idx - 2) * 70));
                }
            });
        }
    });

    window.addEventListener("mouseup", () => {
        isDragging = false;
    });
}



// Populate the Node Grid with detailed node information
function populateNodeGrid(nodes) {
    const nodeGrid = document.getElementById("nodeGrid");
    nodeGrid.innerHTML = "";  // Clear previous nodes

    nodes.forEach(node => {
        const nodeTile = document.createElement("div");
        nodeTile.className = "col-md-3 node-tile";
        nodeTile.innerHTML = `
            <h5>${node.name || "Unnamed Node"}</h5>
            <p>${node.type || "No node type found."}</p>
            <p>${node.metadata?.description || "No description available."}</p>
        `;

        // Attach click event to add the node to the canvas
        nodeTile.onclick = () => {
            addNodeToCanvas(node);
        };

        nodeGrid.appendChild(nodeTile);
    });
}

// Draw the SVG grid
function drawSVGGrid() {
    const gridSize = 50;
    const width = window.innerWidth;
    const height = window.innerHeight;

    svgCanvas.setAttribute('width', width);
    svgCanvas.setAttribute('height', height);
    svgCanvas.innerHTML = ''; // Clear the canvas before redrawing

    for (let x = 0; x <= width; x += gridSize) {
        const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        line.setAttribute('x1', x);
        line.setAttribute('y1', 0);
        line.setAttribute('x2', x);
        line.setAttribute('y2', height);
        line.setAttribute('stroke', '#444');
        line.setAttribute('stroke-width', '0.5');
        svgCanvas.appendChild(line);
    }

    for (let y = 0; y <= height; y += gridSize) {
        const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        line.setAttribute('x1', 0);
        line.setAttribute('y1', y);
        line.setAttribute('x2', width);
        line.setAttribute('y2', y);
        line.setAttribute('stroke', '#444');
        line.setAttribute('stroke-width', '0.5');
        svgCanvas.appendChild(line);
    }
}

// Redraw grid on resize
window.addEventListener('resize', drawSVGGrid);
drawSVGGrid();
