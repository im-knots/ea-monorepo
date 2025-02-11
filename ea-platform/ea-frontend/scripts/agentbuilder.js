// API URLs
const AINU_MANAGER_URL = "http://ainu-manager.ea.erulabs.local/api/v1";
const AGENT_MANAGER_URL = "http://agent-manager.ea.erulabs.local/api/v1";
const JOB_API_URL = "http://job-api.ea.erulabs.local/api/v1";

let userId = null;

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

function showModal(isSuccess, message = "") {
    if (isSuccess) {
        const successModal = new bootstrap.Modal(document.getElementById('successModal'));
        successModal.show();
    } else {
        document.getElementById('errorMessage').textContent = message || "An unexpected error occurred.";
        const errorModal = new bootstrap.Modal(document.getElementById('errorModal'));
        errorModal.show();
    }
}


function saveJson() {
    const jsonEditor = document.getElementById("jsonEditor");
    const agentName = document.getElementById("agentNameInput").value;

    try {
        const jsonData = JSON.parse(jsonEditor.value);

        if (!jsonData.name) {
            jsonData.name = agentName || "Untitled Agent";
        }

        fetch(`${AGENT_MANAGER_URL}/agents`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(jsonData)
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to save agent: ${response.statusText}`);
            }
            return response.json();
        })
        .then(data => {
            console.log("Agent saved successfully:", data);
            showModal(true); // Show success modal
        })
        .catch(error => {
            console.error("Error saving agent:", error);
            showModal(false, `❌ Error saving agent: ${error.message}`); // Show error modal with message
        });

    } catch (error) {
        showModal(false, "❌ Invalid JSON format!");
        console.error("Invalid JSON:", error);
    }
}

document.getElementById("goToAgentManagerBtn").addEventListener("click", function() {
    window.location.href = "/html/agentmanager.html"; // Update this path as needed
});

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

function deleteNode(nodeGroup) {
    // Remove all connections linked to this node
    connections
        .filter(conn => conn.source === nodeGroup || conn.target === nodeGroup)
        .forEach(conn => {
            svgCanvas.removeChild(conn.arrow);
        });

    // Filter out the deleted connections from the connections array
    const remainingConnections = connections.filter(
        conn => conn.source !== nodeGroup && conn.target !== nodeGroup
    );
    connections.length = 0;
    connections.push(...remainingConnections);

    // Remove the node itself
    svgCanvas.removeChild(nodeGroup);
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
    userId = await fetchFirstUserId();
    if (userId) {
        getNodes(userId);
    }
});

// Select the SVG canvas
const svgCanvas = document.getElementById('svgCanvas');

let selectedSourceNode = null;
const connections = []; 

// Function to draw an arrow between two nodes
function drawArrow(source, target) {
    const arrow = document.createElementNS("http://www.w3.org/2000/svg", "line");
    arrow.setAttribute("stroke", "#ffffff");
    arrow.setAttribute("stroke-width", "2");
    arrow.setAttribute("marker-end", "url(#arrowhead)");

    svgCanvas.insertBefore(arrow, svgCanvas.firstChild); // Draw below nodes

    connections.push({ source, target, arrow });  // New: Track the connection

    updateArrowPosition(source, target, arrow);   // New: Initial positioning
}

function updateArrowPosition(source, target, arrow) {
    const sourceRect = source.querySelector("rect");
    const targetRect = target.querySelector("rect");

    // Get position from the "transform" attribute
    const [sourceX, sourceY] = source
        .getAttribute("transform")
        .match(/-?\d+(\.\d+)?/g)
        .map(Number);

    const [targetX, targetY] = target
        .getAttribute("transform")
        .match(/-?\d+(\.\d+)?/g)
        .map(Number);

    const startX = sourceX + parseFloat(sourceRect.getAttribute("width"));
    const startY = sourceY + parseFloat(sourceRect.getAttribute("height")) / 2;

    const endX = targetX;
    const endY = targetY + parseFloat(targetRect.getAttribute("height")) / 2;

    arrow.setAttribute("x1", startX);
    arrow.setAttribute("y1", startY);
    arrow.setAttribute("x2", endX);
    arrow.setAttribute("y2", endY);
}
// Add marker for arrowheads
const marker = document.createElementNS("http://www.w3.org/2000/svg", "marker");
marker.setAttribute("id", "arrowhead");
marker.setAttribute("markerWidth", "10");
marker.setAttribute("markerHeight", "7");
marker.setAttribute("refX", "10");
marker.setAttribute("refY", "3.5");
marker.setAttribute("orient", "auto");
marker.setAttribute("markerUnits", "strokeWidth");

const arrowPath = document.createElementNS("http://www.w3.org/2000/svg", "path");
arrowPath.setAttribute("d", "M0,0 L10,3.5 L0,7 Z");
arrowPath.setAttribute("fill", "#00ff00");
marker.appendChild(arrowPath);

const defs = document.createElementNS("http://www.w3.org/2000/svg", "defs");
defs.appendChild(marker);
svgCanvas.appendChild(defs);

// Add a node to the canvas
function addNodeToCanvas(node) {
    const nodeGroup = document.createElementNS("http://www.w3.org/2000/svg", "g");
    const rect = document.createElementNS("http://www.w3.org/2000/svg", "rect");
    const nameText = document.createElementNS("http://www.w3.org/2000/svg", "text");
    const typeText = document.createElementNS("http://www.w3.org/2000/svg", "text");

    let posX = 50 + Math.random() * 200;
    let posY = 50 + Math.random() * 200;

    const nodeWidth = 240;
    const padding = 15;
    let currentY = padding;

    nodeGroup.setAttribute("transform", `translate(${posX}, ${posY})`);

    rect.setAttribute("width", nodeWidth);
    rect.setAttribute("rx", 10);
    rect.setAttribute("ry", 10);
    rect.setAttribute("fill", "#2d2d2d");
    rect.setAttribute("stroke", "#ffffff");
    rect.setAttribute("stroke-width", "2");
    rect.style.cursor = "move";
    nodeGroup.appendChild(rect);

    nameText.setAttribute("x", nodeWidth / 2);
    nameText.setAttribute("y", currentY + 20);
    nameText.setAttribute("fill", "#ffffff");
    nameText.setAttribute("font-size", "16px");
    nameText.setAttribute("text-anchor", "middle");
    nameText.textContent = node.name || "Unnamed Node";
    nodeGroup.appendChild(nameText);
    currentY += 35;

    typeText.setAttribute("x", nodeWidth / 2);
    typeText.setAttribute("y", currentY);
    typeText.setAttribute("fill", "#aaaaaa");
    typeText.setAttribute("font-size", "13px");
    typeText.setAttribute("text-anchor", "middle");
    typeText.textContent = node.type || "Unknown Type";
    nodeGroup.appendChild(typeText);
    currentY += 30;

    const aliasFO = document.createElementNS("http://www.w3.org/2000/svg", "foreignObject");
    aliasFO.setAttribute("x", padding);
    aliasFO.setAttribute("y", currentY);
    aliasFO.setAttribute("width", nodeWidth - 2 * padding);
    aliasFO.setAttribute("height", 40);

    const aliasInput = document.createElement("input");
    aliasInput.setAttribute("type", "text");
    aliasInput.classList.add("svg-input");
    aliasInput.style.padding = "8px";
    aliasInput.style.width = "100%";
    aliasInput.value = node.alias || `node${Date.now()}`;

    aliasInput.addEventListener("input", () => {
        nodeGroup.dataset.alias = aliasInput.value;
        updateJsonSidebar();
    });

    aliasFO.appendChild(aliasInput);
    nodeGroup.appendChild(aliasFO);
    nodeGroup.dataset.alias = aliasInput.value;

    currentY += 50;

    if (node.parameters) {
        node.parameters.forEach((param) => {
            const label = document.createElementNS("http://www.w3.org/2000/svg", "text");
            label.setAttribute("x", padding);
            label.setAttribute("y", currentY);
            label.setAttribute("fill", "#ffffff");
            label.setAttribute("font-size", "13px");
            label.textContent = param.key;
            nodeGroup.appendChild(label);

            currentY += 20;

            const foreignObject = document.createElementNS("http://www.w3.org/2000/svg", "foreignObject");
            foreignObject.setAttribute("x", padding);
            foreignObject.setAttribute("y", currentY);
            foreignObject.setAttribute("width", nodeWidth - 2 * padding);
            foreignObject.setAttribute("height", 45);

            if (param.type === "bool") {
                const toggleContainer = document.createElement("label");
                toggleContainer.classList.add("switch");

                const input = document.createElement("input");
                input.setAttribute("type", "checkbox");
                input.checked = param.default || false;

                input.addEventListener("change", updateJsonSidebar);

                const slider = document.createElement("span");
                slider.classList.add("slider");

                toggleContainer.appendChild(input);
                toggleContainer.appendChild(slider);
                foreignObject.appendChild(toggleContainer);
            } else if (param.enum) {
                const select = document.createElement("select");
                select.classList.add("svg-select");
                select.style.padding = "6px";

                param.enum.forEach(optionValue => {
                    const option = document.createElement("option");
                    option.value = optionValue;
                    option.textContent = optionValue;
                    if (optionValue === param.default) {
                        option.selected = true;
                    }
                    select.appendChild(option);
                });

                select.addEventListener("change", updateJsonSidebar);
                foreignObject.appendChild(select);
            } else {
                const input = document.createElement("input");
                input.setAttribute("type", "text");
                input.setAttribute("value", param.default || "");
                input.classList.add("svg-input");
                input.style.padding = "8px";
                input.style.width = "100%";

                input.addEventListener("input", updateJsonSidebar);
                foreignObject.appendChild(input);
            }

            nodeGroup.appendChild(foreignObject);
            currentY += 55;
        });
    }

    rect.setAttribute("height", currentY + padding);

    svgCanvas.appendChild(nodeGroup);

    // Dragging functionality
    let isDragging = false;
    let startX, startY;

    nodeGroup.addEventListener("mousedown", (e) => {
        if (e.target.tagName !== "INPUT" && e.target.tagName !== "SELECT" && e.target.tagName !== "LABEL") {
            isDragging = true;
            startX = e.clientX;
            startY = e.clientY;
        }
    });

    window.addEventListener("mousemove", (e) => {
        if (isDragging) {
            const dx = e.clientX - startX;
            const dy = e.clientY - startY;

            posX += dx;
            posY += dy;

            nodeGroup.setAttribute("transform", `translate(${posX}, ${posY})`);

            connections.forEach(({ source, target, arrow }) => {
                if (source === nodeGroup || target === nodeGroup) {
                    updateArrowPosition(source, target, arrow);
                }
            });

            startX = e.clientX;
            startY = e.clientY;

            updateJsonSidebar(); // Trigger JSON update on node move
        }
    });

    window.addEventListener("mouseup", () => {
        isDragging = false;
    });

    // Double-click to select as source node
    nodeGroup.addEventListener("dblclick", () => {
        if (selectedSourceNode) {
            selectedSourceNode.querySelector("rect").setAttribute("stroke", "#ffffff");
        }
        selectedSourceNode = nodeGroup;
        rect.setAttribute("stroke", "#00ff00");
    });

    // Single-click to select as target node and draw arrow
    nodeGroup.addEventListener("click", () => {
        if (selectedSourceNode && selectedSourceNode !== nodeGroup) {
            drawArrow(selectedSourceNode, nodeGroup);
            updateJsonSidebar();
            selectedSourceNode.querySelector("rect").setAttribute("stroke", "#ffffff");
            selectedSourceNode = null;
        }
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
            updateJsonSidebar();
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

function updateJsonSidebar(agentName = null) {
    const defaultAgentName = "My Agent";
    const agentDescription = "An example agent using the Ollama LLM definition.";

    const nodes = Array.from(svgCanvas.querySelectorAll("g")).map((nodeGroup) => {
        const aliasInput = nodeGroup.querySelector("foreignObject input");
        const alias = aliasInput ? aliasInput.value : `node${Date.now()}`;

        nodeGroup.dataset.alias = alias;

        const type = nodeGroup.querySelectorAll("text")[1]?.textContent || "unknown.type";
        const params = {};

        nodeGroup.querySelectorAll("foreignObject").forEach((fo) => {
            const inputElement = fo.querySelector("input, select");
            const key = fo.previousSibling?.textContent || "param";
            const value = inputElement?.type === "checkbox" 
                ? inputElement.checked 
                : inputElement?.value || "";

            if (key !== alias && key !== type && key.toLowerCase() !== "alias" && key.toLowerCase() !== "type") {
                params[key] = value;
            }
        });

        return {
            alias,
            type,
            parameters: params
        };
    });

    const edges = connections.map(conn => ({
        from: [conn.source.dataset.alias],
        to: [conn.target.dataset.alias]
    }));

    const agentPayload = {
        name: agentName || document.getElementById("agentNameInput").value || defaultAgentName,
        creator: userId,
        description: agentDescription,
        nodes,
        edges
    };

    const jsonEditor = document.getElementById("jsonEditor");
    if (jsonEditor) {
        jsonEditor.textContent = JSON.stringify(agentPayload, null, 4);
    }
}


// Redraw grid on resize
window.addEventListener('resize', drawSVGGrid);
drawSVGGrid();

window.addEventListener("keydown", (e) => {
    if ((e.key === "Backspace" || e.key === "Delete") && selectedSourceNode) {
        deleteNode(selectedSourceNode);
        updateJsonSidebar();
        selectedSourceNode = null; // Deselect after deletion
    }
});

// Agent Name Input Handler
document.getElementById("agentNameInput").addEventListener("input", (event) => {
    const agentName = event.target.value;
    updateJsonSidebar(agentName);  // Pass the updated name to the function
});