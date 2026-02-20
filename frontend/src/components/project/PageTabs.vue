<script setup>
import { ref, nextTick } from 'vue'
import { PlusIcon, TrashIcon, XMarkIcon } from '@heroicons/vue/24/outline'

const props = defineProps({
  pages: {
    type: Array,
    default: () => []
  },
  activePage: String,
  editMode: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['selectPage', 'createPage', 'deletePage'])

const showAddDialog = ref(false)
const newPageTitle = ref('')
const showDeleteConfirm = ref(null)
const pageTitleInput = ref(null)

const openAddDialog = () => {
  newPageTitle.value = ''
  showAddDialog.value = true
  nextTick(() => {
    pageTitleInput.value?.focus()
  })
}

const closeAddDialog = () => {
  showAddDialog.value = false
  newPageTitle.value = ''
}

const confirmAdd = () => {
  if (newPageTitle.value.trim()) {
    emit('createPage', newPageTitle.value.trim())
    closeAddDialog()
  }
}

const openDeleteConfirm = (page) => {
  showDeleteConfirm.value = page
}

const closeDeleteConfirm = () => {
  showDeleteConfirm.value = null
}

const confirmDelete = () => {
  if (showDeleteConfirm.value) {
    emit('deletePage', showDeleteConfirm.value.resourceName)
    closeDeleteConfirm()
  }
}
</script>

<template>
  <div class="flex items-center space-x-1 px-4 py-2 border-b border-gray-200 dark:border-gray-700 overflow-x-auto">
    <div
      v-for="page in pages"
      :key="page.resourceName"
      class="flex items-center group"
    >
      <button
        @click="emit('selectPage', page.resourceName)"
        :class="[
          'px-4 py-2 text-sm font-medium rounded-t-lg transition-colors whitespace-nowrap',
          activePage === page.resourceName
            ? 'bg-white dark:bg-gray-900 border border-b-0 border-gray-200 dark:border-gray-700 text-gray-900 dark:text-white'
            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
        ]"
      >
        {{ page.title || page.name }}
      </button>
      <button
        v-if="editMode"
        @click.stop="openDeleteConfirm(page)"
        class="p-1 ml-1 rounded hover:bg-red-100 dark:hover:bg-red-900/50"
        title="Delete page"
      >
        <TrashIcon class="w-4 h-4 text-red-500" />
      </button>
    </div>
    <span v-if="pages.length === 0" class="text-sm text-gray-400 dark:text-gray-500 px-4">
      No pages configured
    </span>
    <!-- Add page button -->
    <button
      v-if="editMode"
      @click="openAddDialog"
      class="flex items-center px-3 py-2 text-sm font-medium text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
      title="Add new page"
    >
      <PlusIcon class="w-4 h-4 mr-1" />
      <span>Add Page</span>
    </button>
  </div>

  <!-- Add Page Dialog -->
  <div v-if="showAddDialog" class="fixed inset-0 z-50 overflow-y-auto">
    <div class="fixed inset-0 bg-black/40 backdrop-blur-md" @click="closeAddDialog"></div>
    <div class="flex min-h-full items-center justify-center p-4">
      <div class="relative bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-sm p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100">
            Create New Page
          </h3>
          <button @click="closeAddDialog" class="text-gray-400 hover:text-gray-500">
            <XMarkIcon class="w-5 h-5" />
          </button>
        </div>
        <input
          ref="pageTitleInput"
          v-model="newPageTitle"
          type="text"
          placeholder="Page title"
          class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-sky-500 focus:border-transparent"
          @keyup.enter="confirmAdd"
        />
        <div class="flex justify-end gap-2 mt-4">
          <button
            @click="closeAddDialog"
            class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            Cancel
          </button>
          <button
            @click="confirmAdd"
            :disabled="!newPageTitle.trim()"
            class="px-4 py-2 text-sm font-medium text-white bg-sky-600 rounded-md hover:bg-sky-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Create
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- Delete Confirm Dialog -->
  <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 overflow-y-auto">
    <div class="fixed inset-0 bg-black/40 backdrop-blur-md" @click="closeDeleteConfirm"></div>
    <div class="flex min-h-full items-center justify-center p-4">
      <div class="relative bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-sm p-6">
        <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
          Delete Page?
        </h3>
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-6">
          Are you sure you want to delete "{{ showDeleteConfirm.title || showDeleteConfirm.name }}"? This action cannot be undone.
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
            Delete
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
