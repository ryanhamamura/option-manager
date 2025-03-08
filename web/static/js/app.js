document.addEventListener('DOMContentLoaded', function() {
    // Initialize all event listeners
    initializeTradeButtons();
    
    // Set up auto-refresh for dashboard
    setupAutoRefresh();
});

/**
 * Initialize event listeners for trade action buttons
 */
function initializeTradeButtons() {
    // Place trade buttons
    document.querySelectorAll('.place-trade-btn').forEach(btn => {
        btn.addEventListener('click', function(e) {
            const tradeId = this.getAttribute('data-id');
            const tradeText = this.getAttribute('data-text');
            placeTrade(tradeId, tradeText);
        });
    });

    // Execute trade buttons
    document.querySelectorAll('.execute-trade-btn').forEach(btn => {
        btn.addEventListener('click', function(e) {
            const tradeId = this.getAttribute('data-id');
            updateTradeStatus(tradeId, 'EXECUTED');
        });
    });

    // Delete trade buttons
    document.querySelectorAll('.delete-trade-btn').forEach(btn => {
        btn.addEventListener('click', function(e) {
            const tradeId = this.getAttribute('data-id');
            deleteTrade(tradeId);
        });
    });
}

/**
 * Place a trade with the broker
 * @param {string} tradeId - The ID of the trade to place
 * @param {string} tradeText - The formatted text to send to the broker
 */
async function placeTrade(tradeId, tradeText) {
    try {
        // Copy trade text to clipboard for the user to paste into broker
        await navigator.clipboard.writeText(tradeText);
        
        // Show message that text was copied
        showToast('Trade details copied to clipboard!');
        
        // Call the API to mark the trade as placed
        const response = await fetch(`/api/trades/${tradeId}/place`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error('Failed to place trade');
        }
        
        // Reload the trade list
        setTimeout(() => {
            refreshTradesList();
        }, 1000);
        
    } catch (error) {
        console.error('Error placing trade:', error);
        showToast('Error placing trade', true);
    }
}

/**
 * Update the status of a trade
 * @param {string} tradeId - The ID of the trade to update
 * @param {string} status - The new status (PENDING, PLACED, EXECUTED, FAILED)
 */
async function updateTradeStatus(tradeId, status) {
    try {
        const response = await fetch(`/api/trades/${tradeId}/status`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ status })
        });
        
        if (!response.ok) {
            throw new Error('Failed to update trade status');
        }
        
        showToast(`Trade marked as ${status.toLowerCase()}`);
        
        // Reload the trade list
        setTimeout(() => {
            refreshTradesList();
        }, 1000);
        
    } catch (error) {
        console.error('Error updating trade status:', error);
        showToast('Error updating trade status', true);
    }
}

/**
 * Delete a trade
 * @param {string} tradeId - The ID of the trade to delete
 */
async function deleteTrade(tradeId) {
    if (!confirm('Are you sure you want to delete this trade?')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/trades/${tradeId}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            throw new Error('Failed to delete trade');
        }
        
        showToast('Trade deleted successfully');
        
        // Reload the trade list
        setTimeout(() => {
            refreshTradesList();
        }, 1000);
        
    } catch (error) {
        console.error('Error deleting trade:', error);
        showToast('Error deleting trade', true);
    }
}

/**
 * Refresh the trades list
 */
async function refreshTradesList() {
    // Get current URL to preserve status filter
    const url = new URL(window.location.href);
    // Reload the page
    window.location.href = url.href;
}

/**
 * Set up auto-refresh for the dashboard
 */
function setupAutoRefresh() {
    // Auto-refresh every 30 seconds
    setInterval(() => {
        refreshTradesList();
    }, 30000);
}

/**
 * Show a toast notification
 * @param {string} message - The message to display
 * @param {boolean} isError - Whether this is an error message
 */
function showToast(message, isError = false) {
    const toast = document.getElementById('toast');
    const toastMessage = document.getElementById('toast-message');
    
    // Set message and style
    toastMessage.textContent = message;
    
    if (isError) {
        toast.classList.remove('bg-green-500');
        toast.classList.add('bg-red-500');
    } else {
        toast.classList.remove('bg-red-500');
        toast.classList.add('bg-green-500');
    }
    
    // Show the toast
    toast.classList.remove('translate-y-20');
    
    // Hide after 3 seconds
    setTimeout(() => {
        toast.classList.add('translate-y-20');
    }, 3000);
}
