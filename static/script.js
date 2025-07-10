document.addEventListener('alpine:init', () => {
  Alpine.store('confirmModal', {
    show: false,
    idToDelete: null
  });
});

function taskManager() {
  return {
    tasks: [],
    activeTask: null,
    newTitle: '',
    newDetail: '',
    pagination: {},
    filters: {
      status: 'all',
      sortBy: 'start_time',
      orderBy: 'desc',
    },
    currentPage: 1,
    pageSize: 10,
    liveDuration: '00:00:00',
    activeDurationInterval: null,
    activeTab: 'active',
    editingTask: null,
    confirmModal: {
      show: false,
      idToDelete: null,
    },

    init() {
      this.fetchTasks();
      this.getRunningTask();
      this.$watch('filters', () => {
        this.currentPage = 1;
        this.fetchTasks();
      }, { deep: true });
    },

    formatDuration(seconds) {
      const h = Math.floor(seconds / 3600).toString().padStart(2, '0');
      const m = Math.floor((seconds % 3600) / 60).toString().padStart(2, '0');
      const s = (seconds % 60).toString().padStart(2, '0');
      return `${h}:${m}:${s}`;
    },

    async fetchTasks() {
      try {
        const params = new URLSearchParams({
          page: this.currentPage,
          limit: this.pageSize,
          status: this.filters.status,
          sortBy: this.filters.sortBy,
          orderBy: this.filters.orderBy,
        });
        const res = await fetch(`/tasks?${params.toString()}`);
        const data = await res.json();
        this.tasks = data.data || [];
        this.pagination = data.pagination;
      } catch (err) {
        console.error('Failed to fetch tasks:', err);
      }
    },

    async getRunningTask() {
      try {
        const res = await fetch(`/tasks/running`);
        const task = await res.json();
        if (task && task.id > 0) {
          this.activeTask = task;
          this.startLiveDuration();
        } else {
          this.activeTask = null;
          this.stopLiveDuration();
        }
      } catch (err) {
        console.error('Failed to get running task:', err);
        this.activeTask = null;
      }
    },

    startLiveDuration() {
        if (this.activeDurationInterval) clearInterval(this.activeDurationInterval);
        if (!this.activeTask) return;

        const lastResumeTime = new Date(this.activeTask.last_resume_time);
        const previousTrackedDuration = parseInt(this.activeTask.duration, 10);

        this.activeDurationInterval = setInterval(() => {
            const currentRunDuration = Math.floor((Date.now() - lastResumeTime.getTime()) / 1000);
            const totalDuration = previousTrackedDuration + currentRunDuration;
            this.liveDuration = this.formatDuration(totalDuration);
        }, 1000);
    },

    stopLiveDuration() {
        if (this.activeDurationInterval) {
            clearInterval(this.activeDurationInterval);
            this.activeDurationInterval = null;
        }
    },

    async createTask() {
      if (!this.newTitle.trim()) return;
      try {
        const res = await fetch('/tasks', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ title: this.newTitle, detail: this.newDetail })
        });
        const data = await res.json();
        if (data.error) this.showToast(data.error, "error");
        this.newTitle = '';
        this.newDetail = '';
        this.fetchTasks();
        this.getRunningTask();
      } catch (err) {
        console.error('Error creating task:', err);
      }
    },

    async pauseTask(id) {
      await fetch(`/tasks/${id}/pause`, { method: 'POST' });
      this.fetchTasks();
      this.getRunningTask();
    },

    async resumeTask(id) {
      const res = await fetch(`/tasks/${id}/resume`, { method: 'POST' });
      const data = await res.json();
      if (data.error) this.showToast(data.error), "error";
      this.fetchTasks();
      this.getRunningTask();
    },

    async stopTask(id) {
      await fetch(`/tasks/${id}/stop`, { method: 'POST' });
      this.fetchTasks();
      this.getRunningTask();
    },

    async deleteTask(id) {
      try {
        const res = await fetch(`/tasks/${id}`, { method: 'DELETE' });
        if (!res.ok) {
          const data = await res.json();
          throw new Error(data.error || 'Failed to delete task');
        }
        this.fetchTasks();
        this.showToast('Task deleted successfully');
      } catch (error) {
        this.showToast(error.message, 'error');
      }
    },

    changePage(page) {
        if (page > 0 && page <= this.pagination.totalPages) {
            this.currentPage = page;
            this.fetchTasks();
        }
    },

    getTaskStatus(task) {
        if (!task.end_time) {
            return task.is_paused ? 'Paused' : 'Running';
        }
        return 'Stopped';
    },

    getTaskStatusColor(task) {
        if (!task.end_time) {
            return task.is_paused ? 'bg-yellow-100 text-yellow-900' : 'bg-green-100 text-green-900';
        }
        return 'bg-gray-100 text-gray-800';
    },

    showDeleteConfirm(id) {
      this.confirmModal.idToDelete = id;
      this.confirmModal.show = true;
    },

    showEditModal(task) {
      this.editingTask = { ...task };
    },

    cancelEdit() {
      this.editingTask = null;
    },

    showToast(message, type = 'success') {
      this.$dispatch('notify', { message, type });
    },

    showConfirm(message, callback) {
      this.$refs.confirmModal.message = message;
      this.$refs.confirmModal.callback = callback;
      this.$refs.confirmModal.show = true;
    },

    async updateTask() {
      if (!this.editingTask) return;

      try {
        const res = await fetch(`/tasks/${this.editingTask.id}`,
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            title: this.editingTask.title,
            detail: this.editingTask.detail,
          })
        });
        if (res.ok) {
          const shouldRefreshActive = !this.editingTask.end_time && !this.editingTask.is_paused;
          this.fetchTasks();
          if (shouldRefreshActive) {
            this.getRunningTask();
          }
          this.cancelEdit();
        } else {
          const data = await res.json();
          this.showToast(data.error || 'Failed to update task', "error");
        }
      } catch (err) {
        console.error('Error updating task:', err);
      }
    }
  }
}
