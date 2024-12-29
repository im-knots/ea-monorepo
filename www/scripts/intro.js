// Global Variables
let skipRequested = false;
let sceneCompleted = false;
let isHyperspaceActive = false;
let starAnimationFrame;

// DOM Elements
const terminalContainer = document.getElementById("terminal-container");
const typedTextContainer = document.getElementById("typed-text-container");
const starCanvas = document.getElementById("star-canvas");
const ctx = starCanvas.getContext("2d");

// Data for Typing
let typedSoFar = "$ ";
const lines = [
  { text: "./erulabs.ai", delayPerChar: 100 },
  { text: "\n    loading eru........ DONE", delayPerChar: 100 },
  { text: "\n    starting ainulindale................. DONE ", delayPerChar: 100 },
  { text: "\n    launching Ea ................... DONE ", delayPerChar: 100 }
];
let currentLineIndex = 0;
let charIndex = 0;

// Event Listener for Enter Key
document.addEventListener("keydown", (event) => {
  if (event.key === "Enter") {
    skipScene();
  }
});

// Resize Star Canvas
function resizeCanvas() {
  starCanvas.width = window.innerWidth;
  starCanvas.height = window.innerHeight;
}
window.addEventListener("resize", resizeCanvas);
resizeCanvas();

/***************************************************************
 * Scene Management
 ***************************************************************/
function skipScene() {
  if (sceneCompleted) return;

  skipRequested = true;

  if (isHyperspaceActive) {
    stopHyperspace();
    completeSequence();
  } else {
    terminalContainer.style.opacity = 0;
    setTimeout(startHyperspace, 100);
  }
}

function completeSequence() {
  sceneCompleted = true;

  // Redirect to dashboard.html
  window.location.href = "html/dashboard.html";
}

/***************************************************************
 * Terminal Typing
 ***************************************************************/
function createCursor() {
  const c = document.createElement("span");
  c.className = "cursor";
  return c;
}

function typeNextCharacter() {
  if (skipRequested) return;

  const oldCursor = typedTextContainer.querySelector(".cursor");
  if (oldCursor) oldCursor.remove();

  const line = lines[currentLineIndex];
  if (!line) return;

  if (charIndex < line.text.length) {
    typedSoFar += line.text.charAt(charIndex);
    charIndex++;
    typedTextContainer.textContent = typedSoFar;
    typedTextContainer.appendChild(createCursor());
    setTimeout(typeNextCharacter, line.delayPerChar);
  } else {
    typedTextContainer.textContent = typedSoFar;
    currentLineIndex++;
    charIndex = 0;

    if (currentLineIndex < lines.length) {
      setTimeout(typeNextCharacter, 700);
    } else {
      typedTextContainer.appendChild(createCursor());
      setTimeout(() => {
        terminalContainer.style.opacity = 0;
        setTimeout(startHyperspace, 1200);
      }, 1000);
    }
  }
}

setTimeout(typeNextCharacter, 500);

/***************************************************************
 * Starfield (Hyperspace)
 ***************************************************************/
class Star {
  constructor() {
    this.x = starCanvas.width / 2;
    this.y = starCanvas.height / 2;
    this.angle = Math.random() * 2 * Math.PI;
    this.speed = 2 + Math.random() * 5;
    this.size = 2;
  }
  update() {
    this.x += Math.cos(this.angle) * this.speed;
    this.y += Math.sin(this.angle) * this.speed;
    this.speed *= 1.03;
  }
  draw() {
    ctx.fillStyle = "white";
    ctx.fillRect(this.x, this.y, this.size, this.size);
  }
}

let stars = [];

function createStars(count) {
  for (let i = 0; i < count; i++) {
    stars.push(new Star());
  }
}

function animateStars() {
  if (!isHyperspaceActive) return;

  ctx.fillStyle = "rgba(0,0,0,0.3)";
  ctx.fillRect(0, 0, starCanvas.width, starCanvas.height);

  stars.forEach((s, i) => {
    s.update();
    s.draw();
    if (s.x < 0 || s.x > starCanvas.width || s.y < 0 || s.y > starCanvas.height) {
      stars[i] = new Star();
    }
  });

  starAnimationFrame = requestAnimationFrame(animateStars);
}

function startHyperspace() {
  if (sceneCompleted) return;

  starCanvas.style.opacity = 1;
  isHyperspaceActive = true;
  createStars(200);
  animateStars();

  setTimeout(() => {
    if (!sceneCompleted) {
      stopHyperspace();
      completeSequence();
    }
  }, 5000);
}

function stopHyperspace() {
  isHyperspaceActive = false;
  cancelAnimationFrame(starAnimationFrame);
  starCanvas.style.opacity = 0;
}
