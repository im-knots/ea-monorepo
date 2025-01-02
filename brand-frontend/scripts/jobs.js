document.addEventListener('DOMContentLoaded', () => {
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

    fetch('apply.html')
    .then(response => response.text())
    .then(data => {
        document.body.insertAdjacentHTML('beforeend', data); // Add modal to the end of the body
    })
    .catch(error => console.error('Error loading apply modal:', error));

    const applyForm = document.getElementById('applyForm');
    const applyModal = new bootstrap.Modal(document.getElementById('applyModal'));

    if (applyForm) {
        applyForm.addEventListener('submit', async (e) => {
            e.preventDefault(); // Prevent default form submission

            const formData = new FormData(applyForm); // Collect form data

            try {
                const response = await fetch('http://localhost:8082/apply', {
                    method: 'POST',
                    body: formData,
                });

                if (response.ok) {
                    // Change modal content to success message
                    waitlistModalBody.innerHTML = `
                        <div class="text-center">
                            <h5>Thanks for applying!</h5>
                            <p>We'll be in touch soon.</p>
                        </div>
                    `;
                    setTimeout(() => {
                        waitlistModal.hide(); // Close the modal after showing the message
                        window.location.hash = '#'; // Redirect to the top of the page (or another section)
                    }, 3000); // Adjust the delay (3 seconds here) as needed
                } else {
                    alert('Failed to submit the application. Please try again.');
                }
            } catch (error) {
                console.error('Error submitting application:', error);
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
                    setTimeout(() => {
                        waitlistModal.hide(); 
                        window.location.hash = '#contact'; 
                    }, 3000); 
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