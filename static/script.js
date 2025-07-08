// Helper to format seconds into hh:mm:ss
function formatDuration(seconds) {
    const h = Math.floor(seconds / 3600).toString().padStart(2, '0');
    const m = Math.floor((seconds % 3600) / 60).toString().padStart(2, '0');
    const s = (seconds % 60).toString().padStart(2, '0');
    return `${h}:${m}:${s}`;
}

let activeDurationInterval;
let currentPage = 1;
const pageSize = 10;
let currentStatus = 'all';
let currentSortBy = 'start_time';
let currentOrderBy = 'desc';

document.addEventListener("DOMContentLoaded", () => {
    const filterStatusEl = document.getElementById("filterStatus");
    const sortByEl = document.getElementById("sortBy");
    const orderByEl = document.getElementById("orderBy");

    function handleControlChange() {
        currentPage = 1;
        currentStatus = filterStatusEl.value;
        currentSortBy = sortByEl.value;
        currentOrderBy = orderByEl.value;
        fetchTasks();
    }

    filterStatusEl.addEventListener("change", handleControlChange);
    sortByEl.addEventListener("change", handleControlChange);
    orderByEl.addEventListener("change", handleControlChange);

    // Form handler
    const createForm = document.getElementById('createForm');
    createForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const titleInput = document.getElementById('titleInput');
        const detailInput = document.getElementById('detailInput');
        const title = titleInput.value.trim();
        const detail = detailInput.value.trim();
        if (title) {
            await createTask(title, detail);
            titleInput.value = '';
            detailInput.value = '';
        }
    });

    // Initial fetch
    fetchTasks();
    getRunningTask();
});

// Fetch and render tasks
async function fetchTasks() {
    try {
        const params = new URLSearchParams({
            page: currentPage,
            limit: pageSize,
            status: currentStatus,
            sortBy: currentSortBy,
            orderBy: currentOrderBy,
        });

        const res = await fetch(`/tasks?${params.toString()}`);
        const responseData = await res.json();

        const tasks = responseData.data || [];
        const pagination = responseData.pagination;

        const listDiv = document.getElementById('taskList');
        listDiv.innerHTML = '';

        tasks.forEach(task => {
            const card = document.createElement('div');
            card.className = `p-4 rounded-lg shadow bg-white border border-gray-200 flex flex-col space-y-2`;

            let status = 'Stopped';
            let statusColor = 'bg-gray-100 text-gray-800';

            if (!task.end_time) {
                if (task.is_paused) {
                    status = 'Paused';
                    statusColor = 'bg-yellow-100 text-yellow-900';
                } else {
                    status = 'Running';
                    statusColor = 'bg-green-100 text-green-900';
                }
            }
            
            card.innerHTML = `
                <div class="flex justify-between items-center">
                    <span class="font-semibold text-lg">${task.title}</span>
                    <span class="text-xs px-2 py-1 rounded ${statusColor}">${status}</span>
                </div>
                ${task.detail ? `<div class="text-sm text-gray-700 mt-1">${task.detail}</div>` : ''}
                <div class="text-sm text-gray-600">
                    Started: ${new Date(task.start_time).toLocaleString()}<br>
                    ${task.end_time ? `Ended: ${new Date(task.end_time).toLocaleString()}<br>` : ''}
                    Duration: ${task.is_paused == 1 ? formatDuration(task.paused_duration) : formatDuration(task.duration)}
                </div>
                <div class="mt-2 flex flex-wrap gap-2">
                    ${!task.end_time ? `
                        ${task.is_paused
                        ? `<button class="px-3 py-1 bg-green-500 text-white rounded hover:bg-green-600" onclick="resumeTask(${task.id})">Resume</button>`
                        : `<button class="px-3 py-1 bg-yellow-500 text-white rounded hover:bg-yellow-600" onclick="pauseTask(${task.id})">Pause</button>`
                    }
                        <button class="px-3 py-1 bg-red-500 text-white rounded hover:bg-red-600" onclick="stopTask(${task.id})">Stop</button>
                    ` : ''}
                    ${task.is_paused || task.end_time ? `<button class="px-3 py-1 bg-gray-400 text-white rounded hover:bg-gray-500" onclick="deleteTask(${task.id})">Delete</button>` : ``}
                </div>
            `;
            listDiv.appendChild(card);
        });

        renderPaginationControls(pagination);
    } catch (err) {
        console.error('Failed to fetch tasks:', err);
    }
}

async function getRunningTask() {
    const res = await fetch(`/tasks/running`);
    const task = await res.json();

    let activeFound = task.id > 0 ? true : false;

    const activeDiv = document.getElementById('activeTask');
    activeDiv.innerHTML = '';

    if (!activeFound) {
        activeDiv.innerHTML = '<p class="text-gray-500">No active task.</p>';
    } else {
        const card = document.createElement('div');
        card.className = `flex flex-col space-y-2`;
        card.innerHTML = `
            <div class="flex justify-between items-center">
                <span class="font-semibold text-lg text-blue-800">${task.title}</span>
                <span class="text-xs px-2 py-1 rounded bg-blue-100 text-blue-800">Running</span>
            </div>
            <div class="text-sm text-gray-600">
                Started: ${new Date(task.start_time).toLocaleString()}<br>
                ${task.end_time ? `Ended: ${new Date(task.end_time).toLocaleString()}<br>` : ''}
                Duration: <span class="duration-live" data-last-resume-time="${task.last_resume_time}" data-previous-tracked-duration="${task.duration}">${formatDuration(task.duration)}</span>
            </div>
            <div class="mt-2 flex flex-wrap gap-2">
                ${!task.end_time ? `
                    ${task.is_paused
                    ? `<button class="px-3 py-1 bg-green-500 text-white rounded hover:bg-green-600" onclick="resumeTask(${task.id})">Resume</button>`
                    : `<button class="px-3 py-1 bg-yellow-500 text-white rounded hover:bg-yellow-600" onclick="pauseTask(${task.id})">Pause</button>`
                }
                    <button class="px-3 py-1 bg-red-500 text-white rounded hover:bg-red-600" onclick="stopTask(${task.id})">Stop</button>
                ` : ''}
            </div>
        `;
        activeDiv.appendChild(card);

        clearInterval(activeDurationInterval);
        const activeCard = document.querySelector('#activeTask .duration-live');
        if (activeCard) {
            const lastResumeTime = new Date(activeCard.dataset.lastResumeTime);
            const previousTrackedDuration = parseInt(activeCard.dataset.previousTrackedDuration, 10);
            
            activeDurationInterval = setInterval(() => {
                const currentRunDuration = Math.floor((Date.now() - lastResumeTime.getTime()) / 1000);
                const totalDuration = previousTrackedDuration + currentRunDuration;
                activeCard.textContent = formatDuration(totalDuration);
            }, 1000);
        }
    }
}

// CRUD handlers
async function createTask(title, detail) {
    try {
        const res = await fetch('/tasks', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title, detail })
        });
        const data = await res.json();
        if (data.error) alert(data.error);
        await fetchTasks();
        await getRunningTask();
    } catch (err) {
        console.error('Error creating task:', err);
    }
}

async function pauseTask(id) {
    await fetch(`/tasks/${id}/pause`, { method: 'POST' });
    await fetchTasks();
    await getRunningTask();
}

async function resumeTask(id) {
    const res = await fetch(`/tasks/${id}/resume`, { method: 'POST' });
    const data = await res.json();
    if (data.error) alert(data.error);
    await fetchTasks();
    await getRunningTask();
}

async function stopTask(id) {
    await fetch(`/tasks/${id}/stop`, { method: 'POST' });
    await fetchTasks();
    await getRunningTask();
}

async function deleteTask(id) {
    if (confirm('Are you sure you want to delete this task?')) {
        await fetch(`/tasks/${id}`, { method: 'DELETE' });
        await fetchTasks();
    }
}

function renderPaginationControls(pagination) {
    const controlDiv = document.getElementById("paginationControls");
    if (!controlDiv) return;

    controlDiv.innerHTML = "";

    if (!pagination || pagination.totalPages < 1) {
        return
    }

    const btnBaseClasses = "px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 hover:bg-gray-50";
    const btnDisabledClasses = "opacity-50 cursor-not-allowed";

    const prevButton = document.createElement("button");
    prevButton.innerText = "Previous";
    prevButton.className = `${btnBaseClasses} rounded-l-md`;
    if (!pagination.hasPrev) {
        prevButton.disabled = true;
        prevButton.className += ` ${btnDisabledClasses}`;
    }
    prevButton.onclick = () => {
        currentPage = currentPage - 1;
        fetchTasks();
    };

    const nextButton = document.createElement("button");
    nextButton.innerText = "Next";
    nextButton.className = `${btnBaseClasses} rounded-r-md`;
    if (!pagination.hasNext) {
        nextButton.disabled = true;
        nextButton.className = ` ${btnDisabledClasses}`;
    }
    nextButton.onclick = () => {
        currentPage = currentPage + 1;
        fetchTasks();
    };

    const pageInfo = document.createElement("span");
    pageInfo.className = "px-4 py-2 text-sm text-gray-700";
    pageInfo.innerText = `Page ${pagination.currentPage} of ${pagination.totalPages}`;

    controlDiv.append(prevButton, pageInfo, nextButton);
}

