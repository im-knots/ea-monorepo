// Sidebar Toggle
const toggleSidebar = document.getElementById('toggle-sidebar');
const sidebar = document.getElementById('sidebar');

toggleSidebar.addEventListener('click', () => {
  sidebar.classList.toggle('collapsed');
});