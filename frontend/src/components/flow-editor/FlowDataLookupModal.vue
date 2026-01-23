<script setup>
import { ref, computed, watch } from 'vue'
import { XMarkIcon } from '@heroicons/vue/24/outline'
import VueJsonPretty from 'vue-json-pretty'
import 'vue-json-pretty/lib/styles.css'
import { RunExpression } from '../../../wailsjs/go/main/App'
import { useDark } from '@vueuse/core'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  sourceData: {
    type: Object,
    default: () => ({})
  },
  sourceSchema: {
    type: Object,
    default: () => ({})
  },
  targetSchema: {
    type: Object,
    default: () => ({})
  },
  fullSchema: {
    type: Object,
    default: () => ({})
  },
  initialExpression: {
    type: String,
    default: ''
  },
  portName: {
    type: String,
    default: ''
  },
  fieldTitle: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['close', 'apply'])
const isDark = useDark()

// Expression input
const dataExpressionResult = ref('')
const expressionValidationResult = ref(null)
const expressResultValid = ref(false)
const expressResultValidError = ref(null)
const testing = ref(false)

// Initialize expression when modal opens
watch(() => props.show, (newVal) => {
  if (newVal) {
    let expression = null
    if (typeof props.initialExpression === 'string') {
      const match = props.initialExpression.match(/^\{\{(.+)\}\}$/)
      if (match) {
        expression = match[1]
      } else if (props.initialExpression && !props.initialExpression.startsWith('{{')) {
        expression = props.initialExpression
      }
    }
    dataExpressionResult.value = expression || ''
    expressionValidationResult.value = null
    expressResultValid.value = false
    expressResultValidError.value = null
  }
})

// Clean schema - matches platform's cleanUpSchema function
const cleanUpSchema = (obj) => {
  if (!obj || typeof obj !== 'object') return obj

  const copy = JSON.parse(JSON.stringify(obj))

  const cleanRecursive = (o) => {
    for (let prop in o) {
      if (prop === 'configure' || prop === 'configurable' || prop === 'default' ||
          prop === 'title' || prop === 'requiredWhen' || prop === 'optionalWhen' ||
          prop === 'expression' || prop === 'path' || prop === 'propertyOrder' ||
          prop === 'description' || prop === 'enumTitles' || prop === 'colSpan' ||
          prop === 'align' || prop === 'style') {
        delete o[prop]
      }
      if (prop === 'type' && o[prop] === '') {
        delete o[prop]
      }
      if (prop === 'type' && o[prop] === 'object') {
        o['additionalProperties'] = true
      }
      if (typeof o[prop] === 'object') {
        cleanRecursive(o[prop])
      }
    }
  }

  cleanRecursive(copy)
  return copy
}

const cleanedTargetSchema = computed(() => {
  let copy = JSON.parse(JSON.stringify(props.targetSchema || {}))

  if (copy.items && copy.items['$ref']) {
    const ref = copy.items['$ref'].substring('#/$defs/'.length)
    if (copy['$defs'] === undefined) {
      copy['$defs'] = {}
    }
    if (props.fullSchema?.['$defs']?.[ref]) {
      copy['$defs'][ref] = props.fullSchema['$defs'][ref]
    }
  }

  return cleanUpSchema(copy)
})

const sourceDataString = computed(() => {
  try {
    return JSON.stringify(props.sourceData)
  } catch {
    return '{}'
  }
})

const targetSchemaString = computed(() => {
  try {
    return JSON.stringify(cleanedTargetSchema.value)
  } catch {
    return '{}'
  }
})

watch(dataExpressionResult, () => {
  expressResultValid.value = false
  expressResultValidError.value = null
})

const addExpression = async (force = false) => {
  if (expressResultValid.value || force) {
    emit('apply', dataExpressionResult.value, props.portName)
    emit('close')
    return
  }

  if (!dataExpressionResult.value) {
    emit('close')
    return
  }

  testing.value = true
  try {
    const response = await RunExpression(
      dataExpressionResult.value,
      sourceDataString.value,
      targetSchemaString.value
    )

    if (response.result) {
      try {
        expressionValidationResult.value = JSON.parse(response.result)
      } catch {
        expressionValidationResult.value = response.result
      }
    }

    expressResultValid.value = response.validSchema
    expressResultValidError.value = response.validationError || null
  } catch (err) {
    expressResultValid.value = false
    expressResultValidError.value = String(err)
  } finally {
    testing.value = false
  }
}

const close = () => {
  emit('close')
}

const buttonText = computed(() => {
  if (testing.value) return 'Testing...'
  if (expressResultValid.value) return 'Inject Expression'
  if (dataExpressionResult.value) return 'Test'
  return 'Close'
})
</script>

<template>
  <Teleport to="body">
    <div
      v-if="show"
      class="fixed inset-0 z-50 overflow-y-auto p-4 sm:p-6 md:p-20"
    >
      <!-- Backdrop -->
      <div
        class="fixed inset-0 bg-gray-500/25 dark:bg-black/75 backdrop-blur-sm"
        @click="close"
      />

      <!-- Modal -->
      <div class="relative rounded-lg bg-white dark:bg-black dark:border dark:border-gray-800 shadow-xl p-1 max-w-6xl mx-auto dark:text-gray-300">
        <!-- Close button -->
        <div class="absolute top-0 right-0 pt-2 pr-2">
          <button
            @click="close"
            type="button"
            class="rounded-md text-gray-300 hover:text-gray-500 dark:bg-gray-600 dark:text-gray-800 focus:outline-none dark:hover:text-gray-400"
          >
            <XMarkIcon class="h-6 w-6" />
          </button>
        </div>

        <form @submit.prevent="addExpression(false)">
          <div class="sm:flex sm:items-start">
            <div class="mt-3 text-center sm:mt-0 sm:text-left w-full">
              <!-- Title -->
              <h3 class="font-medium leading-6 text-gray-900 dark:text-gray-300">
                {{ fieldTitle || 'Context' }}
              </h3>

              <!-- 3 columns -->
              <div class="py-2 flex w-full">
                <!-- Source data -->
                <div class="w-2/6 pr-2">
                  <div class="json-panel pt-2">
                    <div class="text-xs font-medium text-center">Source data</div>
                    <VueJsonPretty
                      v-if="sourceData"
                      :height="300"
                      :highlight-selected-node="true"
                      :theme="isDark ? 'dark' : 'light'"
                      v-model:selectedValue="dataExpressionResult"
                      :node-selectable="() => true"
                      :data="sourceData"
                      root-path="$"
                      selectable-type="single"
                    />
                  </div>
                </div>

                <!-- Schema -->
                <div class="w-2/6">
                  <div class="text-xs font-medium text-center">Required JSON schema of the result</div>
                  <div class="json-panel">
                    <VueJsonPretty
                      :height="300"
                      :theme="isDark ? 'dark' : 'light'"
                      :data="cleanedTargetSchema"
                    />
                  </div>
                </div>

                <!-- Result -->
                <div class="w-2/6">
                  <div class="text-xs font-medium text-center">The result of the expression</div>
                  <div class="json-panel text-xs">
                    <VueJsonPretty
                      v-if="typeof expressionValidationResult === 'object' && expressionValidationResult !== null"
                      :height="300"
                      :theme="isDark ? 'dark' : 'light'"
                      :data="expressionValidationResult"
                    />
                    <div v-else class="p-2">{{ expressionValidationResult ?? 'null' }}</div>
                  </div>
                </div>
              </div>

              <!-- Status -->
              <div
                v-if="expressResultValidError"
                class="bg-red-100 text-xs rounded-md py-2 px-3 mb-1 text-red-700 dark:bg-red-900 dark:text-gray-300"
              >
                {{ expressResultValidError }}
              </div>
              <div
                v-else-if="expressResultValid"
                class="bg-green-100 text-xs rounded-md py-2 px-3 mb-1 text-green-700 dark:bg-green-900 dark:text-gray-300"
              >
                The result of the expression returns data which is valid to the JSON schema
              </div>

              <!-- Expression + buttons -->
              <div class="grid grid-cols-12">
                <div class="col-span-9">
                  <textarea
                    v-model="dataExpressionResult"
                    class="placeholder-gray-400 py-3 border shadow-sm text-xs appearance-none rounded-md w-full py-2 px-2 text-gray-700 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
                    placeholder="Expression"
                    rows="1"
                  />
                </div>
                <div class="flex col-span-3 space-x-2 pl-1 justify-items-stretch">
                  <button
                    type="submit"
                    :disabled="testing"
                    class="px-4 py-3 border w-full border-transparent text-xs font-medium rounded-md text-sky-700 bg-sky-100 hover:bg-sky-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-sky-500 dark:bg-gray-900 dark:hover:bg-gray-800 dark:text-sky-500 disabled:opacity-50"
                  >
                    {{ buttonText }}
                  </button>
                  <button
                    v-if="expressResultValidError"
                    type="button"
                    @click="addExpression(true)"
                    class="px-4 py-3 border w-full border-transparent text-xs font-medium rounded-md text-sky-700 bg-sky-100 hover:bg-sky-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-sky-500 dark:bg-gray-900 dark:hover:bg-gray-800 dark:text-sky-500"
                  >
                    Inject anyway
                  </button>
                </div>
              </div>
            </div>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.json-panel {
  height: 300px;
  overflow: scroll;
}

:deep(.vjs-tree) {
  font-size: 10px !important;
  line-height: 1.4 !important;
}

:deep(.vjs-tree-node:hover) {
  background-color: rgba(59, 130, 246, 0.15) !important;
}

:deep(.vjs-tree-node.is-highlight) {
  background-color: rgba(59, 130, 246, 0.25) !important;
}

.dark :deep(.vjs-tree-node:hover) {
  background-color: rgba(59, 130, 246, 0.25) !important;
}

.dark :deep(.vjs-tree-node.is-highlight) {
  background-color: rgba(59, 130, 246, 0.35) !important;
}
</style>
