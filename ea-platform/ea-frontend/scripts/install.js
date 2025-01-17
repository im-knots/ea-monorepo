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