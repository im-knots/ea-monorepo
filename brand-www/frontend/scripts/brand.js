document.addEventListener('DOMContentLoaded', () => {
    // Contact form submission
    const contactForm = document.getElementById('contactForm');
    if (contactForm) {
        contactForm.addEventListener('submit', async (e) => {
            e.preventDefault(); // Prevent the default form submission

            const formData = new FormData(contactForm); // Collect form data
            const jsonData = Object.fromEntries(formData.entries()); // Convert to JSON

            try {
                const response = await fetch('http://localhost:8082/submit', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(jsonData),
                });

                if (response.ok) {
                    contactForm.reset();
                    // Show success modal for contact form
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
    }

    // Subscription form submission
    const subscribeForm = document.getElementById('subscribeForm');
    if (subscribeForm) {
        subscribeForm.addEventListener('submit', async (e) => {
            e.preventDefault(); // Prevent the default form submission

            const email = subscribeForm.querySelector('input[name="email"]').value;

            if (!email) {
                alert('Please enter a valid email address.');
                return;
            }

            try {
                const response = await fetch('http://localhost:8082/subscribe', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email }),
                });

                if (response.ok) {
                    subscribeForm.reset();
                    // Show success modal for subscription
                    const subscriptionSuccessModal = new bootstrap.Modal(document.getElementById('subscriptionSuccessModal'));
                    subscriptionSuccessModal.show();
                } else {
                    alert('Failed to subscribe. Please try again later.');
                }
            } catch (error) {
                console.error('Error subscribing:', error);
                alert('An error occurred. Please try again.');
            }
        });
    }
});
