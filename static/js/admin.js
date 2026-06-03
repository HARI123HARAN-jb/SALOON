document.addEventListener('DOMContentLoaded', () => {
    // DOM Sections
    const authSection = document.getElementById('authSection');
    const dashboardSection = document.getElementById('dashboardSection');
    
    // Login Form
    const loginForm = document.getElementById('loginForm');
    const usernameInput = document.getElementById('adminUsername');
    const passwordInput = document.getElementById('adminPassword');
    
    // Tabs & Filters
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');
    const bookingFilter = document.getElementById('bookingFilter');
    
    // Logged In Admin Details
    const logoutBtn = document.getElementById('logoutBtn');
    
    // Summary Cards
    const totalBookingsCard = document.getElementById('totalBookings');
    const pendingBookingsCard = document.getElementById('pendingBookings');
    const activeReviewsCard = document.getElementById('activeReviews');
    
    // Data Tables
    const bookingsTableBody = document.getElementById('bookingsTableBody');
    const reviewsTableBody = document.getElementById('reviewsTableBody');
    const toast = document.getElementById('toast');

    // Authentication Guard
    checkAuth();

    function checkAuth() {
        fetch('/api/admin/check')
            .then(res => res.json())
            .then(data => {
                if (data.authenticated) {
                    showDashboard();
                } else {
                    showLogin();
                }
            })
            .catch(() => showLogin());
    }

    function showLogin() {
        authSection.style.display = 'block';
        dashboardSection.style.display = 'none';
        document.body.className = 'admin-body';
    }

    function showDashboard() {
        authSection.style.display = 'none';
        dashboardSection.style.display = 'block';
        document.body.className = 'admin-body';
        loadDashboardData();
    }

    // Login Submission
    if (loginForm) {
        loginForm.addEventListener('submit', (e) => {
            e.preventDefault();
            
            const submitBtn = loginForm.querySelector('button[type="submit"]');
            submitBtn.disabled = true;
            submitBtn.textContent = 'Logging in...';

            const payload = {
                username: usernameInput.value.trim(),
                password: passwordInput.value
            };

            fetch('/api/admin/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            })
            .then(async res => {
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Login failed');
                return data;
            })
            .then(() => {
                showToast('Login successful', 'success');
                loginForm.reset();
                showDashboard();
            })
            .catch(err => {
                showToast(err.message, 'error');
            })
            .finally(() => {
                submitBtn.disabled = false;
                submitBtn.textContent = 'Login';
            });
        });
    }

    // Logout
    if (logoutBtn) {
        logoutBtn.addEventListener('click', (e) => {
            e.preventDefault();
            fetch('/api/admin/logout', { method: 'POST' })
                .then(() => {
                    showToast('Logged out successfully', 'success');
                    showLogin();
                })
                .catch(() => showLogin());
        });
    }

    // Tab Switching
    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const targetTab = btn.getAttribute('data-tab');
            
            tabBtns.forEach(b => b.classList.remove('active'));
            tabContents.forEach(c => c.classList.remove('active'));
            
            btn.classList.add('active');
            document.getElementById(targetTab).classList.add('active');
        });
    });

    // Booking Status Filter
    if (bookingFilter) {
        bookingFilter.addEventListener('change', () => {
            loadBookings(bookingFilter.value);
        });
    }

    // Dashboard Data Aggregator
    function loadDashboardData() {
        loadBookings(bookingFilter ? bookingFilter.value : '');
        loadReviews();
    }

    // Fetch and Render Bookings
    function loadBookings(status = '') {
        if (!bookingsTableBody) return;

        let url = '/api/admin/bookings';
        if (status) url += `?status=${status}`;

        fetch(url)
            .then(res => res.json())
            .then(bookings => {
                // Update stats cards
                updateBookingStats(bookings);

                if (bookings.length === 0) {
                    bookingsTableBody.innerHTML = '<tr><td colspan="7" style="text-align:center;">No bookings found.</td></tr>';
                    return;
                }

                bookingsTableBody.innerHTML = '';
                bookings.forEach(booking => {
                    const tr = document.createElement('tr');
                    const formattedDate = new Date(booking.created_at).toLocaleDateString();
                    
                    let actionHtml = '';
                    if (booking.status === 'pending') {
                        actionHtml = `
                            <div class="action-group">
                                <button class="action-btn action-btn-confirm" onclick="updateBookingStatus('${booking.id}', 'confirmed')">Confirm</button>
                                <button class="action-btn action-btn-cancel" onclick="updateBookingStatus('${booking.id}', 'cancelled')">Cancel</button>
                            </div>
                        `;
                    } else if (booking.status === 'confirmed') {
                        actionHtml = `
                            <div class="action-group">
                                <button class="action-btn action-btn-complete" onclick="updateBookingStatus('${booking.id}', 'completed')">Complete</button>
                                <button class="action-btn action-btn-cancel" onclick="updateBookingStatus('${booking.id}', 'cancelled')">Cancel</button>
                            </div>
                        `;
                    } else {
                        actionHtml = '<span style="color: #666;">No Actions</span>';
                    }

                    tr.innerHTML = `
                        <td><strong>${escapeHTML(booking.name)}</strong><br><small style="color: #888;">${escapeHTML(booking.phone)}</small></td>
                        <td>${escapeHTML(booking.service)}</td>
                        <td>${escapeHTML(booking.date)}<br><small style="color: #d4af37;">${escapeHTML(booking.time)}</small></td>
                        <td><small style="color: #888;">${escapeHTML(booking.notes || '—')}</small></td>
                        <td>${formattedDate}</td>
                        <td><span class="badge badge-${booking.status}">${booking.status}</span></td>
                        <td>${actionHtml}</td>
                    `;
                    bookingsTableBody.appendChild(tr);
                });
            })
            .catch(err => {
                console.error(err);
                showToast('Failed to load bookings', 'error');
            });
    }

    // Fetch and Render Reviews
    function loadReviews() {
        if (!reviewsTableBody) return;

        fetch('/api/admin/reviews')
            .then(res => res.json())
            .then(reviews => {
                // Update stats card for approved reviews
                const approvedCount = reviews.filter(r => r.status === 'approved').length;
                if (activeReviewsCard) activeReviewsCard.textContent = approvedCount;

                if (reviews.length === 0) {
                    reviewsTableBody.innerHTML = '<tr><td colspan="5" style="text-align:center;">No reviews found.</td></tr>';
                    return;
                }

                reviewsTableBody.innerHTML = '';
                reviews.forEach(review => {
                    const tr = document.createElement('tr');
                    const stars = '★'.repeat(review.rating) + '☆'.repeat(5 - review.rating);
                    
                    let actionHtml = '';
                    if (review.status === 'pending') {
                        actionHtml = `
                            <div class="action-group">
                                <button class="action-btn action-btn-complete" onclick="updateReviewStatus('${review.id}', 'approved')">Approve</button>
                                <button class="action-btn action-btn-cancel" onclick="deleteReview('${review.id}')">Delete</button>
                            </div>
                        `;
                    } else {
                        actionHtml = `
                            <div class="action-group">
                                <button class="action-btn action-btn-cancel" onclick="deleteReview('${review.id}')">Delete</button>
                            </div>
                        `;
                    }

                    tr.innerHTML = `
                        <td><strong>${escapeHTML(review.name)}</strong></td>
                        <td style="color: #d4af37; font-size: 16px;">${stars}</td>
                        <td style="font-style: italic;">"${escapeHTML(review.text)}"</td>
                        <td><span class="badge ${review.status === 'approved' ? 'badge-completed' : 'badge-pending'}">${review.status}</span></td>
                        <td>${actionHtml}</td>
                    `;
                    reviewsTableBody.appendChild(tr);
                });
            })
            .catch(err => {
                console.error(err);
                showToast('Failed to load reviews', 'error');
            });
    }

    // Helper: Update booking dashboard totals
    function updateBookingStats(bookings) {
        if (totalBookingsCard) totalBookingsCard.textContent = bookings.length;
        
        const pendingCount = bookings.filter(b => b.status === 'pending').length;
        if (pendingBookingsCard) pendingBookingsCard.textContent = pendingCount;
    }

    // Expose action trigger functions globally so onclick="..." in dynamically built rows works.
    window.updateBookingStatus = function(id, status) {
        fetch(`/api/admin/bookings/${id}/status`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status: status })
        })
        .then(async res => {
            const data = await res.json();
            if (!res.ok) throw new Error(data.error || 'Failed to update booking');
            return data;
        })
        .then(() => {
            showToast(`Booking marked as ${status}`, 'success');
            loadDashboardData();
        })
        .catch(err => showToast(err.message, 'error'));
    };

    window.updateReviewStatus = function(id, status) {
        fetch(`/api/admin/reviews/${id}/approve`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status: status })
        })
        .then(async res => {
            const data = await res.json();
            if (!res.ok) throw new Error(data.error || 'Failed to update review');
            return data;
        })
        .then(() => {
            showToast(`Review status updated to ${status}`, 'success');
            loadDashboardData();
        })
        .catch(err => showToast(err.message, 'error'));
    };

    window.deleteReview = function(id) {
        if (!confirm('Are you sure you want to permanently delete this review?')) return;

        fetch(`/api/admin/reviews/${id}`, {
            method: 'DELETE'
        })
        .then(async res => {
            const data = await res.json();
            if (!res.ok) throw new Error(data.error || 'Failed to delete review');
            return data;
        })
        .then(() => {
            showToast('Review deleted permanently', 'success');
            loadDashboardData();
        })
        .catch(err => showToast(err.message, 'error'));
    };

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
