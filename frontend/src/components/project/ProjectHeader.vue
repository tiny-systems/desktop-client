<script setup>
import { ref } from 'vue'
import { ArrowLeftIcon, EllipsisVerticalIcon, TrashIcon } from '@heroicons/vue/24/outline'

const props = defineProps({
  title: String,
  clusterName: String,
  projectName: String,
})

const emit = defineEmits(['close', 'delete-project'])

const showMoreMenu = ref(false)
const showDeleteConfirm = ref(false)

const openDeleteConfirm = () => {
  showMoreMenu.value = false
  showDeleteConfirm.value = true
}

const closeDeleteConfirm = () => {
  showDeleteConfirm.value = false
}

const confirmDelete = () => {
  emit('delete-project')
  showDeleteConfirm.value = false
}
</script>

<template>
  <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center space-x-4">
      <button @click="emit('close')" class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full" title="Back to projects">
        <ArrowLeftIcon class="w-5 h-5 text-gray-500 dark:text-gray-400" />
      </button>
      <h1 class="text-xl font-semibold text-gray-900 dark:text-white">{{ title }}</h1>
      <span class="text-sm text-gray-500 dark:text-gray-400" v-if="clusterName">
        Context: {{ clusterName }}
      </span>
    </div>

    <!-- More menu -->
    <div class="relative">
      <button
        @click="showMoreMenu = !showMoreMenu"
        class="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full"
        title="More options"
      >
        <EllipsisVerticalIcon class="w-5 h-5 text-gray-500 dark:text-gray-400" />
      </button>

      <!-- Backdrop to close menu -->
      <div
        v-if="showMoreMenu"
        class="fixed inset-0 z-40"
        @click="showMoreMenu = false"
      ></div>

      <!-- Dropdown menu -->
      <div
        v-if="showMoreMenu"
        class="absolute right-0 top-10 z-50 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-md shadow-lg py-1 min-w-40"
      >
        <button
          @click="openDeleteConfirm"
          class="w-full px-4 py-2 text-left text-sm text-red-600 dark:text-red-400 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center space-x-2"
        >
          <TrashIcon class="w-4 h-4" />
          <span>Delete Project</span>
        </button>
      </div>
    </div>
  </div>

  <!-- Delete Confirmation Dialog -->
  <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 overflow-y-auto" @keydown.enter="confirmDelete" @keydown.escape="closeDeleteConfirm">
    <div class="fixed inset-0 bg-black/40 backdrop-blur-md" @click="closeDeleteConfirm"></div>
    <div class="flex min-h-full items-center justify-center p-4">
      <div class="relative bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-md p-6">
        <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
          Delete Project?
        </h3>
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-6">
          Are you sure you want to delete "<strong class="text-gray-700 dark:text-gray-300">{{ title }}</strong>"?
          This will permanently delete all flows, nodes, and resources associated with this project. This action cannot be undone.
        </p>
        <div class="flex justify-end gap-2">
          <button
            @click="closeDeleteConfirm"
            class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            Cancel
          </button>
          <button
            @click="confirmDelete"
            class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700"
          >
            Delete Project
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
