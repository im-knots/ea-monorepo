document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('contactForm');
  
    form.addEventListener('submit', async (e) => {
      e.preventDefault(); // Prevent the default form submission
  
      const formData = new FormData(form); // Collect form data
      const jsonData = Object.fromEntries(formData.entries()); // Convert to JSON
  
      try {
        const response = await fetch('http://localhost:8082/submit', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(jsonData),
        });
  
        if (response.ok) {
          // Reset the form
          form.reset();
  
          // Trigger the success modal
          const successModal = new bootstrap.Modal(document.getElementById('successModal'));
          successModal.show();
        } else {
          alert('Failed to send message. Please try again later.');
        }
      } catch (error) {
        console.error('Error submitting form:', error);
        alert('An error occurred. Please try again.');
      }
    });
  });
  