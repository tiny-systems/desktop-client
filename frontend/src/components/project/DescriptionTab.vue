<script setup>
import { ref, computed, onMounted } from 'vue'
import { PencilIcon, EyeIcon } from '@heroicons/vue/24/outline'
import { marked } from 'marked'

const GoApp = window.go.main.App

const props = defineProps({
  ctx: String,
  ns: String,
  projectName: String,
})

const description = ref('')
const editing = ref(false)
const saving = ref(false)
const loaded = ref(false)

const renderedHtml = computed(() => {
  if (!description.value) return '<p class="text-gray-400 dark:text-gray-500 italic">No description yet. Click the edit button to add one.</p>'
  return marked(description.value)
})

const loadDescription = async () => {
  if (!GoApp) return
  try {
    const details = await GoApp.GetProjectDetails(props.ctx, props.ns, props.projectName)
    description.value = details.description || ''
    loaded.value = true
  } catch (err) {
    console.error('Failed to load description:', err)
    loaded.value = true
  }
}

const saveDescription = async () => {
  if (!GoApp || saving.value) return
  saving.value = true
  try {
    await GoApp.SaveProjectDescription(props.ctx, props.ns, props.projectName, description.value)
  } catch (err) {
    console.error('Failed to save description:', err)
  } finally {
    saving.value = false
  }
}

const toggleEdit = async () => {
  if (editing.value) {
    // Switching from edit to preview — save
    await saveDescription()
  }
  editing.value = !editing.value
}

onMounted(loadDescription)
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- Toolbar -->
    <div class="flex items-center justify-end px-4 py-2 border-b border-gray-200 dark:border-gray-700">
      <button
        @click="toggleEdit"
        class="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg text-sm transition-colors text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700"
      >
        <PencilIcon v-if="!editing" class="w-4 h-4" />
        <EyeIcon v-else class="w-4 h-4" />
        <span>{{ editing ? 'Preview' : 'Edit' }}</span>
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto">
      <!-- Edit mode -->
      <textarea
        v-if="editing"
        v-model="description"
        @blur="saveDescription"
        placeholder="Write a project description using Markdown..."
        class="w-full h-full p-4 bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 font-mono text-sm resize-none focus:outline-none"
      />
      <!-- Preview mode -->
      <div
        v-else
        class="p-4 prose prose-sm dark:prose-invert max-w-none"
        v-html="renderedHtml"
      />
    </div>
  </div>
</template>
