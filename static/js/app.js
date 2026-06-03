document.addEventListener('DOMContentLoaded', () => {
    // DOM Elements
    const bookingModal = document.getElementById('bookingModal');
    const reviewModal = document.getElementById('reviewModal');
    const openBookingBtns = document.querySelectorAll('.open-booking');
    const openReviewBtn = document.getElementById('openReviewBtn');
    const closeBtns = document.querySelectorAll('.close-btn');
    const bookingForm = document.getElementById('bookingForm');
    const reviewForm = document.getElementById('reviewForm');
    const reviewsGrid = document.getElementById('reviewsGrid');
    const toast = document.getElementById('toast');

    // Populate service options dynamically
    const serviceSelect = document.getElementById('bookingService');
    if (serviceSelect) {
        fetch('/api/services')
            .then(res => res.json())
            .then(services => {
                serviceSelect.innerHTML = '<option value="" disabled selected>Select a Service</option>';
                services.forEach(service => {
                    const opt = document.createElement('option');
                    opt.value = service.name;
                    opt.textContent = `${service.name} (${service.cost})`;
                    serviceSelect.appendChild(opt);
                });
            })
            .catch(err => console.error('Error fetching services:', err));
    }

    // Load approved reviews
    loadReviews();

    function loadReviews() {
        if (!reviewsGrid) return;
        
        fetch('/api/reviews')
            .then(res => res.json())
            .then(reviews => {
                if (reviews.length === 0) {
                    reviewsGrid.innerHTML = `
                        <div class="review-card" style="grid-column: 1/-1; text-align: center; border-left: none;">
                            <p class="review-text">"No reviews approved yet. Be the first to leave a review!"</p>
                        </div>
                    `;
                    return;
                }

                reviewsGrid.innerHTML = '';
                reviews.forEach(review => {
                    const card = document.createElement('div');
                    card.className = 'review-card';
                    
                    const stars = '★'.repeat(review.rating) + '☆'.repeat(5 - review.rating);
                    
                    card.innerHTML = `
                        <div class="stars">${stars}</div>
                        <div class="review-text">"${escapeHTML(review.text)}"</div>
                        <div class="review-author">— ${escapeHTML(review.name)}</div>
                    `;
                    reviewsGrid.appendChild(card);
                });
            })
            .catch(err => {
                console.error('Error loading reviews:', err);
                reviewsGrid.innerHTML = '<p class="review-text" style="color: #e74c3c;">Failed to load reviews.</p>';
            });
    }

    // Modal Control Functions
    function openModal(modal) {
        modal.classList.add('active');
        document.body.style.overflow = 'hidden';
    }

    function closeModal(modal) {
        modal.classList.remove('active');
        document.body.style.overflow = 'auto';
    }

    // Event Listeners for Modals
    openBookingBtns.forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.preventDefault();
            openModal(bookingModal);
        });
    });

    if (openReviewBtn) {
        openReviewBtn.addEventListener('click', (e) => {
            e.preventDefault();
            openModal(reviewModal);
        });
    }

    closeBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            closeModal(bookingModal);
            closeModal(reviewModal);
        });
    });

    // Close modals on overlay click
    window.addEventListener('click', (e) => {
        if (e.target === bookingModal) closeModal(bookingModal);
        if (e.target === reviewModal) closeModal(reviewModal);
    });

    // Booking Form Submit
    if (bookingForm) {
        bookingForm.addEventListener('submit', (e) => {
            e.preventDefault();

            const submitBtn = bookingForm.querySelector('button[type="submit"]');
            submitBtn.disabled = true;
            submitBtn.textContent = 'Booking...';

            const payload = {
                name: document.getElementById('bookingName').value.trim(),
                email: document.getElementById('bookingEmail').value.trim(),
                phone: document.getElementById('bookingPhone').value.trim(),
                service: document.getElementById('bookingService').value,
                date: document.getElementById('bookingDate').value,
                time: document.getElementById('bookingTime').value,
                notes: document.getElementById('bookingNotes').value.trim()
            };

            fetch('/api/bookings', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            })
            .then(async res => {
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Booking failed');
                return data;
            })
            .then(data => {
                showToast('Appointment booked successfully! Check email/WhatsApp.', 'success');
                bookingForm.reset();
                closeModal(bookingModal);
            })
            .catch(err => {
                showToast(err.message, 'error');
            })
            .finally(() => {
                submitBtn.disabled = false;
                submitBtn.textContent = 'Book Appointment';
            });
        });
    }

    // Review Form Submit
    if (reviewForm) {
        reviewForm.addEventListener('submit', (e) => {
            e.preventDefault();

            const submitBtn = reviewForm.querySelector('button[type="submit"]');
            submitBtn.disabled = true;
            submitBtn.textContent = 'Submitting...';

            // Get selected star rating
            const ratingInput = reviewForm.querySelector('input[name="rating"]:checked');
            if (!ratingInput) {
                showToast('Please select a rating star!', 'error');
                submitBtn.disabled = false;
                submitBtn.textContent = 'Submit Review';
                return;
            }

            const payload = {
                name: document.getElementById('reviewName').value.trim(),
                rating: parseInt(ratingInput.value),
                text: document.getElementById('reviewText').value.trim()
            };

            fetch('/api/reviews', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            })
            .then(async res => {
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Submission failed');
                return data;
            })
            .then(data => {
                showToast('Thank you! Review submitted. Pending admin approval.', 'success');
                reviewForm.reset();
                closeModal(reviewModal);
            })
            .catch(err => {
                showToast(err.message, 'error');
            })
            .finally(() => {
                submitBtn.disabled = false;
                submitBtn.textContent = 'Submit Review';
            });
        });
    }

    // Toast Notification Helper
    function showToast(message, type = 'success') {
        if (!toast) return;
        toast.textContent = message;
        toast.className = 'toast show';
        
        if (type === 'success') {
            toast.classList.add('toast-success');
        } else {
            toast.classList.add('toast-error');
        }

        setTimeout(() => {
            toast.classList.remove('show');
        }, 4000);
    }

    // Helper to escape HTML characters
    function escapeHTML(str) {
        return str
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    }
});
