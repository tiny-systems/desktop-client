<script setup>
import { ref, markRaw, computed } from 'vue'
import JSONEditor from '../schema-based-json-editor/JSONEditor.vue'
import { PencilSquareIcon, TrashIcon, DocumentDuplicateIcon } from '@heroicons/vue/24/outline'

const props = defineProps({
  widget: {
    type: Object,
    required: true
  },
  readonly: {
    type: Boolean,
    default: false
  },
  editMode: {
    type: Boolean,
    default: false
  },
  pages: {
    type: Array,
    default: () => []
  }
})

// Local state for page selection dropdown
const showPagesDropdown = ref(false)

// Get the effective schema - use customized schema if available, otherwise defaultSchema
// Deep clone and mark as raw to prevent Vue reactivity from wrapping it
const schemaSnapshot = computed(() => {
  // Use schema (customized) if available and not empty, otherwise use defaultSchema
  const schema = props.widget?.schema && Object.keys(props.widget.schema).length > 0
    ? props.widget.schema
    : props.widget?.defaultSchema
  return schema
    ? markRaw(JSON.parse(JSON.stringify(schema)))
    : null
})

// Key for JSONEditor - changes when data changes to force re-render
// Uses _updateTime timestamp set by handleNodeUpdate
const editorKey = computed(() => {
  return props.widget?.id + '-' + (props.widget?._updateTime || 0)
})

const emit = defineEmits(['action', 'update', 'edit-schema', 'reset-schema', 'update-title', 'update-pages'])

// Computed: which pages is this widget currently on
const widgetPages = computed(() => props.widget.pages || [])

// Check if widget is on a specific page
const isOnPage = (pageName) => {
  return widgetPages.value.includes(pageName)
}

// Toggle page assignment
const togglePage = (pageName) => {
  const currentPages = [...widgetPages.value]
  const idx = currentPages.indexOf(pageName)
  if (idx >= 0) {
    currentPages.splice(idx, 1)
  } else {
    currentPages.push(pageName)
  }
  emit('update-pages', { widget: props.widget, pages: currentPages })
}

const handleUpdateValue = (event) => {
  if (event.isAction) {
    // This is an action button press - send signal
    emit('action', {
      nodeName: props.widget.nodeName,
      port: props.widget.port,
      data: event.value
    })
  } else {
    emit('update', {
      nodeName: props.widget.nodeName,
      port: props.widget.port,
      data: event.value
    })
  }
}

const editWidgetSchema = () => {
  emit('edit-schema', props.widget)
}

const resetWidgetSchema = () => {
  emit('reset-schema', props.widget)
}

const handleTitleChange = (event) => {
  emit('update-title', { widget: props.widget, title: event.target.value })
}
</script>

<template>
  <div class="widget-container h-full flex flex-col bg-white dark:bg-gray-900 overflow-hidden group">
    <div class="widget-header flex items-center justify-between">
      <template v-if="editMode">
        <input
          type="text"
          :value="widget.title || widget.nodeName"
          @change="handleTitleChange"
          @keyup.enter="handleTitleChange($event); $event.target.blur()"
          class="text-xs font-medium text-gray-700 dark:text-gray-300 bg-transparent border border-gray-300 dark:border-gray-600 rounded px-1 py-0.5 w-full"
        />
        <div class="flex items-center ml-1 space-x-0.5 relative">
          <!-- Pages dropdown -->
          <div class="relative">
            <button
              @click="showPagesDropdown = !showPagesDropdown"
              class="p-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 opacity-0 group-hover:opacity-100 transition-opacity"
              title="Assign to pages"
            >
              <DocumentDuplicateIcon class="w-4 h-4 text-gray-500 dark:text-gray-400" />
            </button>
            <!-- Backdrop to close dropdown when clicking outside -->
            <div
              v-if="showPagesDropdown"
              class="fixed inset-0 z-40"
              @click="showPagesDropdown = false"
            ></div>
            <div
              v-if="showPagesDropdown"
              class="absolute right-0 top-6 z-50 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-md shadow-lg py-1 min-w-32"
            >
              <div
                v-for="page in pages"
                :key="page.resourceName"
                @click="togglePage(page.resourceName)"
                class="px-3 py-1 text-xs cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center justify-between"
              >
                <span class="text-gray-700 dark:text-gray-300">{{ page.title || page.name }}</span>
                <span v-if="isOnPage(page.resourceName)" class="text-sky-500 ml-2">✓</span>
              </div>
              <div v-if="pages.length === 0" class="px-3 py-1 text-xs text-gray-400">No pages</div>
            </div>
          </div>
          <button
            @click="editWidgetSchema"
            class="p-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 opacity-0 group-hover:opacity-100 transition-opacity"
            title="Edit widget schema"
          >
            <PencilSquareIcon class="w-4 h-4 text-gray-500 dark:text-gray-400" />
          </button>
          <button
            v-if="widget.schema && Object.keys(widget.schema).length > 0"
            @click="resetWidgetSchema"
            class="p-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 opacity-0 group-hover:opacity-100 transition-opacity"
            title="Reset widget schema"
          >
            <TrashIcon class="w-4 h-4 text-gray-500 dark:text-gray-400" />
          </button>
        </div>
      </template>
      <h3 v-else class="text-sm font-medium text-gray-700 dark:text-gray-300 truncate w-full text-center">
        {{ widget.title || widget.nodeName }}
      </h3>
    </div>
    <div class="widget-content flex-1 overflow-hidden">
      <JSONEditor
        v-if="schemaSnapshot && Object.keys(schemaSnapshot).length > 0"
        :key="editorKey"
        :schema="schemaSnapshot"
        :initial-value="widget.data"
        :readonly="readonly"
        theme="small"
        no-border
        disable-collapse
        @update-value="handleUpdateValue"
      />
      <div v-else class="flex items-center justify-center h-full text-gray-400 dark:text-gray-500 text-sm">
        No schema available
      </div>
    </div>
  </div>
</template>

<style scoped>
.widget-container {
  min-height: 60px;
}
</style>
