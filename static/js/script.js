// --- Configuration ---
const API_BASE_URL = '/api/view/entries';

// --- Reusable Fetch and Render Function ---
async function fetchAndRenderEntries(days) {
    const url = days ? `${API_BASE_URL}?days=${days}` : `${API_BASE_URL}?all=true`;
    const tableBody = document.getElementById('table-body');
    tableBody.innerHTML = '<tr><td colspan="4">Loading...</td></tr>';

    try {
        const response = await fetch(url);

        if (response.status === 204) {
            tableBody.innerHTML = '<tr><td colspan="4">No entries found for this range.</td></tr>';
            return;
        }

        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const data = await response.json();
        renderTable(data, tableBody);

    } catch (error) {
        console.error("Fetch error:", error);
        tableBody.innerHTML = `<tr><td colspan="4" style="color:red;">Error loading data: ${error.message}</td></tr>`;
    }
}


// --- Rendering Logic ---
function renderTable(data, tableBody) {
    tableBody.innerHTML = ''; // Clear existing rows

    if (data.length === 0) {
        tableBody.innerHTML = '<tr><td colspan="4">No entries found.</td></tr>';
        return;
    }

    data.forEach(item => {
        const row = tableBody.insertRow();
        const finishedClass = item.isFinished ? 'finished-true' : 'finished-false';
        row.classList.add(finishedClass);

        row.insertCell().textContent = item.title;
        row.insertCell().textContent = item.date;
        row.insertCell().textContent = item.isFinished ? 'Finished' : 'In Progress';

        // Actions cell with dynamic IDs for Edit/Delete
        const actionsCell = row.insertCell();
        actionsCell.className = 'actions';
        actionsCell.innerHTML = `
            <a href="/ui/edit/${item.id}">Edit</a>
            <a href="/ui/delete/${item.id}">Delete</a>
        `;
    });
}


// --- Event Handlers (Example: If you have navigation buttons/links) ---

// Assuming you have links/buttons that trigger the change:
// <a href="#" onclick="loadView(7)">Last Week</a>
// <a href="#" onclick="loadView(30)">Last Month</a>

window.loadView = function(days) {
    fetchAndRenderEntries(days);
}
