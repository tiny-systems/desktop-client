<script setup>
import { ref, computed, watch } from 'vue'
import { useFlowStore } from '../../stores/flow'
import {
  Combobox,
  ComboboxInput,
  ComboboxOptions,
  ComboboxOption,
  Dialog,
  DialogOverlay,
  TransitionChild,
  TransitionRoot,
} from '@headlessui/vue'
import { MagnifyingGlassIcon } from '@heroicons/vue/24/solid'
import { FolderIcon } from '@heroicons/vue/24/outline'

const props = defineProps({
  modelValue: Boolean,
  position: {
    type: Object,
    default: () => ({ x: 100, y: 100 })
  }
})

const emit = defineEmits(['update:modelValue'])

const flowStore = useFlowStore()
const components = ref([])
const currentComponent = ref(null)
const loading = ref(false)
const error = ref('')
const query = ref('')
const currentTag = ref('')
const searchInputRef = ref(null)

const open = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

// Load components when modal opens
watch(() => props.modelValue, async (isOpen) => {
  if (!isOpen) return

  loading.value = true
  error.value = ''

  try {
    const result = await flowStore.getAvailableComponents()
    components.value = result || []
  } catch (e) {
    error.value = e.message || 'Failed to load components'
    components.value = []
  } finally {
    loading.value = false
  }
}, { immediate: true })

// Filter components by search query and tag
const filteredComponents = computed(() => {
  return components.value.filter((component) => {
    const name = component.name || ''
    const description = component.description || ''
    const info = component.info || ''
    const tags = component.tags || []
    const searchLower = query.value.toLowerCase()

    const matchesQuery = query.value === '' ||
      name.toLowerCase().includes(searchLower) ||
      description.toLowerCase().includes(searchLower) ||
      info.toLowerCase().includes(searchLower) ||
      tags.some(tag => tag.toLowerCase().includes(searchLower))

    const matchesTag = currentTag.value === '' || tags.includes(currentTag.value)

    return matchesQuery && matchesTag
  })
})

// Get unique tags from filtered components
const filteredTags = computed(() => {
  const allTags = components.value
    .map(component => component.tags || [])
    .flat()
  return [...new Set(allTags)].sort()
})

const selectTag = (tag) => {
  currentTag.value = currentTag.value === tag ? '' : tag
}

// Create a new node instance
const createInstance = async (component) => {
  if (!component) return

  loading.value = true
  error.value = ''

  try {
    await flowStore.addNode(
      component.name,
      component.description,
      component.module,
      component.version,
      props.position.x,
      props.position.y
    )
    open.value = false
    query.value = ''
    currentTag.value = ''
  } catch (e) {
    error.value = e.message || 'Failed to create node'
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  open.value = false
  query.value = ''
  currentTag.value = ''
  currentComponent.value = null
}
</script>

<template>
  <TransitionRoot :show="open" as="template" @after-leave="query = ''">
    <Dialog as="div" class="fixed inset-0 z-50 overflow-y-auto p-4 sm:p-6 md:p-20" @close="open = false" :initial-focus="searchInputRef">
      <!-- Backdrop overlay -->
      <TransitionChild
        as="template"
        enter="ease-out duration-300"
        enter-from="opacity-0"
        enter-to="opacity-100"
        leave="ease-in duration-200"
        leave-from="opacity-100"
        leave-to="opacity-0"
      >
        <DialogOverlay class="fixed inset-0 bg-gray-500/25 dark:bg-black/75 backdrop-blur-sm transition-opacity" />
      </TransitionChild>

      <!-- Dialog content -->
      <TransitionChild
        v-if="open"
        as="template"
        enter="ease-out duration-300"
        enter-from="opacity-0 scale-95"
        enter-to="opacity-100 scale-100"
        leave="ease-in duration-200"
        leave-from="opacity-100 scale-100"
        leave-to="opacity-0 scale-95"
      >
        <Combobox
          as="div"
          class="relative z-10 mx-auto max-w-4xl transform overflow-hidden rounded-xl bg-white dark:bg-gray-900 shadow-2xl ring-1 ring-black/5 dark:ring-white/10 transition-all"
          @update:model-value="createInstance"
        >
          <!-- Search input -->
          <div class="relative">
            <MagnifyingGlassIcon class="pointer-events-none absolute top-3.5 left-4 h-6 w-6 text-gray-400" />
            <ComboboxInput
              ref="searchInputRef"
              class="h-12 w-full border-0 bg-transparent pl-11 pr-4 text-gray-800 dark:text-gray-200 placeholder-gray-400 focus:ring-0 sm:text-sm focus:outline-none"
              placeholder="Search components..."
              @change="query = $event.target.value"
            />
          </div>

          <!-- Error message -->
          <div v-if="error" class="px-4 py-2 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-sm">
            {{ error }}
          </div>

          <!-- Tags filter -->
          <div v-if="filteredTags.length > 0" class="p-2 lg:px-8 space-x-1 border-b border-gray-100 dark:border-gray-800">
            <button
              v-for="tag in filteredTags"
              :key="tag"
              type="button"
              @click="selectTag(tag)"
              :class="[
                'text-xs inline-block py-1 px-2 rounded uppercase transition-colors',
                currentTag === tag
                  ? 'bg-sky-500 text-white ring-2 ring-sky-400'
                  : 'bg-sky-100 dark:bg-sky-900 text-sky-700 dark:text-sky-300 hover:bg-sky-200 dark:hover:bg-sky-800'
              ]"
            >
              {{ tag }}
            </button>
          </div>

          <!-- Components list with preview -->
          <div class="grid md:grid-cols-12" @mouseleave="currentComponent = null">
            <ComboboxOptions
              v-if="filteredComponents.length > 0"
              static
              :class="[
                'max-h-96 scroll-py-2 divide-y divide-gray-100 dark:divide-gray-800 overflow-y-auto',
                currentComponent ? 'md:col-span-8' : 'col-span-12'
              ]"
            >
              <li class="p-2">
                <h2 class="sr-only">Components</h2>
                <ul class="text-sm text-gray-700 dark:text-gray-300">
                  <ComboboxOption
                    v-for="component in filteredComponents"
                    :key="`${component.module}-${component.name}`"
                    :value="component"
                    as="template"
                    v-slot="{ active }"
                  >
                    <li
                      :class="[
                        'flex cursor-pointer select-none items-center rounded-md px-3 py-2',
                        active ? 'bg-sky-600 text-white' : ''
                      ]"
                      @mouseover="currentComponent = component"
                    >
                      <span class="lg:ml-3 flex-none">
                        {{ component.description }}
                      </span>
                      <span
                        :class="[
                          'ml-3 flex-auto text-left text-xs truncate',
                          active ? 'text-sky-100' : 'text-gray-400'
                        ]"
                      >
                        {{ component.info }}
                        <span
                          v-for="tag in (component.tags || [])"
                          :key="tag"
                          class="text-xs font-thin inline-block py-0.5 px-1 ml-1 uppercase rounded bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-300"
                        >
                          {{ tag }}
                        </span>
                      </span>
                    </li>
                  </ComboboxOption>
                </ul>
              </li>
            </ComboboxOptions>

            <!-- Component preview panel -->
            <div
              v-if="currentComponent"
              class="col-span-4 p-4 text-sm dark:text-gray-300 text-center border-l border-gray-100 dark:border-gray-800"
            >
              <p class="font-bold text-gray-900 dark:text-white">{{ currentComponent.name }}</p>
              <p class="mt-1 text-gray-600 dark:text-gray-400">{{ currentComponent.info }}</p>
              <p class="mt-3">
                <span class="inline-block bg-green-100 dark:bg-green-900/50 rounded-md py-1 px-2 text-xs text-green-700 dark:text-green-300">
                  {{ currentComponent.module }}
                  <span v-if="currentComponent.version" class="opacity-75">({{ currentComponent.version }})</span>
                </span>
              </p>
            </div>
          </div>

          <!-- Empty state -->
          <div v-if="!loading && filteredComponents.length === 0" class="py-8 px-6 text-center">
            <FolderIcon class="mx-auto h-8 w-8 text-gray-400" />
            <p v-if="query !== ''" class="mt-3 text-sm text-gray-600 dark:text-gray-400">
              No components found matching "{{ query }}"
            </p>
            <p v-else class="mt-3 text-sm text-gray-600 dark:text-gray-400">
              No components available.
            </p>
            <p class="mt-2 text-xs text-gray-500 dark:text-gray-500">
              Make sure modules are installed in the cluster.
            </p>
          </div>

          <!-- Loading overlay -->
          <div v-if="loading" class="absolute inset-0 bg-white/50 dark:bg-gray-900/50 flex items-center justify-center">
            <div class="flex items-center gap-2 text-gray-600 dark:text-gray-400">
              <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-sky-500"></div>
              <span class="text-sm">Loading...</span>
            </div>
          </div>
        </Combobox>
      </TransitionChild>
    </Dialog>
  </TransitionRoot>
</template>
