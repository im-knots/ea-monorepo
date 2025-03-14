document.addEventListener('DOMContentLoaded', async () => {
    try {
        // Load config.json dynamically
        const configResponse = await fetch('/config/config.json');
        const config = await configResponse.json();
        const apiBaseUrl = config.apiBaseUrl;

        // Fetch header and footer
        fetch('header.html')
            .then(response => response.text())
            .then(data => {
                document.getElementById('site-header').outerHTML = data;
            })
            .catch(error => console.error('Error loading header:', error));

        fetch('footer.html')
            .then(response => response.text())
            .then(data => {
                document.getElementById('site-footer').outerHTML = data;
            })
            .catch(error => console.error('Error loading footer:', error));

        // Contact Form Submission
        const contactForm = document.getElementById('contactForm');
        if (contactForm) {
            contactForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                const formData = new FormData(contactForm);
                const jsonData = Object.fromEntries(formData.entries());

                try {
                    const response = await fetch(`${apiBaseUrl}/submit`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(jsonData),
                    });

                    if (response.ok) {
                        contactForm.reset();
                        const successModal = new bootstrap.Modal(document.getElementById('successModal'));
                        successModal.show();
                        setTimeout(() => {
                            successModal.hide();
                            window.location.hash = '#contact';
                        }, 3000);
                    } else {
                        alert('Failed to send message. Please try again later.');
                    }
                } catch (error) {
                    console.error('Error submitting form:', error);
                    alert('An error occurred. Please try again.');
                }
            });
        }

        // Ensure the modal and its body are properly selected
        const waitlistModal = new bootstrap.Modal(document.getElementById('waitlistModal'));
        const waitlistModalBody = document.querySelector('#waitlistModal .modal-body');

        // Waitlist Form Submission
        const waitlistForm = document.getElementById('waitlistForm');
        if (waitlistForm) {
            waitlistForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                const formData = new FormData(waitlistForm);
                const jsonData = Object.fromEntries(formData.entries());

                try {
                    const response = await fetch(`${apiBaseUrl}/waitlist`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(jsonData),
                    });

                    if (response.ok) {
                        // Update the modal content on success
                        waitlistModalBody.innerHTML = `
                            <div class="text-center">
                                <h5>Thanks for the early interest!</h5>
                                <p>We'll be in touch soon.</p>
                            </div>
                        `;
                        setTimeout(() => {
                            waitlistModal.hide();
                            window.location.hash = '#';
                        }, 3000);
                    } else {
                        alert('Failed to join the waitlist. Please try again later.');
                    }
                } catch (error) {
                    console.error('Error submitting waitlist form:', error);
                    alert('An error occurred. Please try again.');
                }
            });
        }

    } catch (error) {
        console.error('Error loading configuration:', error);
    }
});
