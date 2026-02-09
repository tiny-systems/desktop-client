<template>
  <div class="json-schema-editor">
    <div class="flex items-center gap-2 mt-2">
      <!-- Expand button or spacer for alignment -->
      <div class="w-5 flex-shrink-0 flex items-center justify-center">
        <button v-if="isObject" type="button" @click="hidden = !hidden" class="p-0.5">
          <ChevronRightIcon v-if="hidden" class="h-4 w-4 text-blue-500 cursor-pointer"/>
          <ChevronDownIcon v-else class="h-4 w-4 text-blue-500 cursor-pointer"/>
        </button>
      </div>
      <!-- Field name -->
      <div class="w-32 flex-shrink-0">
        <input type="text" :disabled="disabled || root" :value="pickKey" @blur="onInputName"
               class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2 dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500 disabled:text-slate-500 disabled:shadow-none"/>
      </div>
      <!-- Required checkbox -->
      <div class="w-20 flex-shrink-0">
        <label v-if="root" class="dark:text-gray-500 whitespace-nowrap flex items-center">
          <input type="checkbox" :disabled="!isObject && !isArray || !parent" @change="onRootCheck" title="Required"
                 class="w-4 h-4 text-blue-600 bg-gray-100 rounded border-gray-300 focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-800 dark:border-gray-600"/>
          <span class="text-xs pl-1">Required</span>
        </label>
        <label v-else class="dark:text-gray-500 whitespace-nowrap flex items-center">
          <input type="checkbox" :disabled="isItem || !parent" :checked="checked" @change="onCheck" title="Required"
                 class="w-4 h-4 text-blue-600 bg-gray-100 rounded border-gray-300 focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-800 dark:border-gray-600"/>
          <span class="text-xs pl-1">Required</span>
        </label>
      </div>
      <!-- Type dropdown -->
      <div class="w-28 flex-shrink-0">
        <select v-model="pickValue.type" @change="onChangeType"
                class="schema-select bg-gray-50 border border-gray-300 text-gray-900 text-xs rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2 dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500 disabled:text-slate-500 disabled:shadow-none">
          <option v-for="t in filteredTypeNames" :key="t" :value="t">{{ t }}</option>
        </select>
      </div>
      <!-- Title input -->
      <div class="flex-1 min-w-0">
        <input v-model="pickValue.title" :placeholder="local['title']"
               class="bg-gray-50 border border-gray-300 text-gray-900 text-xs rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2 dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500 disabled:text-slate-500 disabled:shadow-none"/>
      </div>
      <!-- Action buttons -->
      <div class="flex items-center gap-1 flex-shrink-0">
        <a v-if="isObject" href="#" @click.prevent="addChild" :title="local['add_child_node']">
          <PlusIcon class="w-4 h-4 text-blue-500 cursor-pointer"/>
        </a>
        <a v-if="pickValue.type" href="#" @click.prevent="onSetting" :title="local['adv_setting']">
          <CogIcon class="w-4 h-4 text-blue-500 cursor-pointer"/>
        </a>
        <a v-if="!root && !isItem" href="#" @click.prevent="removeNode" :title="local['remove_node']">
          <XMarkIcon class="w-4 h-4 text-blue-500 cursor-pointer"/>
        </a>
      </div>
    </div>
    <template v-if="!hidden && pickValue.properties && !isArray">
      <JsonSchemaEditor v-for="(item, key, index) in pickValue.properties" :value="{[key]:item}" :parent="pickValue"
                        :key="index" :deep="deep+1" :root="false" :lang="lang"/>
    </template>
    <template v-if="isArray && pickValue.items">
      <JsonSchemaEditor :value="{items:pickValue.items}" :deep="deep+1" disabled isItem :root="false"
                        :lang="lang"/>
    </template>

    <!-- Advanced Settings Modal -->
    <div v-if="modalVisible" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="fixed inset-0 bg-black/40 backdrop-blur-md" @click="modalVisible = false"></div>
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="relative bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-lg p-4">
          <form @submit.prevent="handleOk">
            <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Advanced Settings</h3>
            <div class="grid gap-x-2 gap-y-2 md:grid-cols-2">
              <div v-for="(item, key) in advancedValue" :key="key">
                <label class="dark:text-gray-500 text-sm" :for="key">{{ local[key] || key }}</label>
                <input type="number" v-model="advancedValue[key]" :id="key"
                       class="bg-gray-50 border border-gray-300 text-gray-900 text-xs rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2 dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500 disabled:shadow-none"
                       v-if="advancedAttr[key].type === 'integer' || advancedAttr[key].type === 'number'" :placeholder="key"/>
                <span v-else-if="advancedAttr[key].type === 'boolean'" style="display:inline-block;width:100%">
                  <input :id="key" type="checkbox" v-model="advancedValue[key]" class="w-4 h-4 my-3 text-blue-600 bg-gray-100 rounded border-gray-300 focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-800 dark:border-gray-600"/>
                </span>
                <textarea @blur="changeEnumValue" :value="enumText" :id="key" :rows="2" v-else-if="key === 'enum'" :placeholder="local['enum_msg']" class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2 dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500 disabled:shadow-none"></textarea>
                <select :id="key" v-else-if="advancedAttr[key].type === 'array'" v-model="advancedValue[key]" style="width:100%" :title="local[key]" class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2 dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500">
                  <option>{{ local['none'] }}</option>
                  <option :key="t" v-for="t in advancedAttr[key].enums">{{ t }}</option>
                </select>
                <input :id="key" type="text" v-model="advancedValue[key]" v-else style="width:100%" :placeholder="key" class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2 dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500 disabled:shadow-none"/>
              </div>
            </div>
            <div class="pt-4 text-right space-x-2">
              <button type="button" @click="modalVisible = false" class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700">Close</button>
              <button type="submit" class="px-4 py-2 text-sm font-medium text-white bg-sky-600 rounded-md hover:bg-sky-700">Save</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import { isNull } from './util'
import { TYPE, TYPE_NAME } from './type/type'
import { CheckIcon, ChevronDownIcon, ChevronRightIcon, CogIcon, PlusIcon, XMarkIcon } from '@heroicons/vue/24/outline'
import LocalProvider from './local-provider'

export default {
  name: 'JsonSchemaEditor',
  components: {
    ChevronRightIcon,
    ChevronDownIcon, CogIcon, PlusIcon, XMarkIcon, CheckIcon
  },
  props: {
    value: {
      type: Object,
      required: true
    },
    disabled: {
      type: Boolean,
      default: false
    },
    isItem: {
      type: Boolean,
      default: false
    },
    deep: {
      type: Number,
      default: 0
    },
    root: {
      type: Boolean,
      default: true
    },
    parent: {
      type: Object,
      default: null
    },
    lang: {
      type: String,
      default: 'en_US'
    }
  },
  computed: {
    pickValue() {
      return Object.values(this.value)[0]
    },
    pickKey() {
      return Object.keys(this.value)[0]
    },
    isObject() {
      return this.pickValue.type === 'object'
    },
    isArray() {
      return this.pickValue.type === 'array'
    },
    filteredTypeNames() {
      // Filter out empty string type for cleaner dropdown
      return TYPE_NAME.filter(t => t !== '')
    },
    checked() {
      return this.parent && this.parent.required && this.parent.required.indexOf(this.pickKey) >= 0
    },
    advanced() {
      return TYPE[this.pickValue.type]
    },
    advancedAttr() {
      return TYPE[this.pickValue.type].attr
    },
    ownProps() {
      return ['type', 'title', 'properties', 'items', 'required', ...Object.keys(this.advancedAttr)]
    },
    advancedNotEmptyValue() {
      const jsonNode = Object.assign({}, this.advancedValue)
      for (let key in jsonNode) {
        isNull(jsonNode[key]) && delete jsonNode[key]
        if (key === 'enum' && jsonNode[key].length === 0) {
          delete jsonNode[key]
        }
      }
      return jsonNode
    },
    completeNodeValue() {
      const t = {}
      const basicValue = { ...this.pickValue }
      this._pickDiffKey().forEach(key => delete basicValue[key])
      return Object.assign({}, basicValue, t, this.advancedNotEmptyValue)
    },
    enumText() {
      const t = this.advancedValue['enum']
      if (!t) return ''
      if (!t.length) return ''
      return t.join('\n')
    }
  },
  data() {
    return {
      TYPE_NAME,
      hidden: false,
      countAdd: 1,
      modalVisible: false,
      advancedValue: {},
      addProp: {},
      local: LocalProvider(this.lang)
    }
  },
  methods: {
    onInputName(e) {
      const oldKey = this.pickKey
      const newKey = e.target.value
      if (oldKey === newKey) return

      const nodeValue = this.parent.properties[oldKey]

      delete this.parent.properties[oldKey]
      this.parent.properties[newKey] = nodeValue

      const requireds = this.parent.required || []
      const oldIndex = requireds.indexOf(oldKey)
      if (requireds.length > 0 && oldIndex > -1) {
        requireds.splice(oldIndex, 1)
        requireds.push(newKey)
        this.parent['required'] = [...new Set(requireds)]
      }
    },
    onChangeType() {
      delete this.pickValue['properties']
      delete this.pickValue['items']
      delete this.pickValue['required']
      delete this.pickValue['format']
      delete this.pickValue['enum']

      if (this.isArray) {
        this.pickValue['items'] = { type: 'string' }
      }
    },
    onCheck(e) {
      this._checked(e.target.checked, this.parent)
    },
    onRootCheck(e) {
      this._deepCheck(e.target.checked, this.pickValue)
    },
    changeEnumValue(e) {
      const pickType = this.pickValue.type
      const value = e.target.value
      let arr = value.split('\n')
      if (arr.length === 0 || (arr.length === 1 && arr[0] === '')) {
        this.advancedValue.enum = null
      } else {
        if (pickType === 'string') {
          this.advancedValue.enum = arr.map(item => item)
        } else {
          this.advancedValue.enum = arr.map(item => +item)
        }
      }
    },
    _deepCheck(checked, node) {
      if (node.type === 'object' && node.properties) {
        checked ? node['required'] = Object.keys(node.properties) : (delete node['required'])
        Object.keys(node.properties).forEach(key => this._deepCheck(checked, node.properties[key]))
      } else if (node.type === 'array' && node.items.type === 'object') {
        checked ? node.items['required'] = Object.keys(node.items.properties) : (delete node.items['required'])
        Object.keys(node.items.properties).forEach(key => this._deepCheck(checked, node.items.properties[key]))
      }
    },
    _checked(checked, parent) {
      let required = parent.required
      if (checked) {
        required || (this.parent['required'] = [])

        required = this.parent.required
        required.indexOf(this.pickKey) === -1 && required.push(this.pickKey)
      } else {
        const pos = required.indexOf(this.pickKey)
        pos >= 0 && required.splice(pos, 1)
      }
      required.length === 0 && (delete parent['required'])
    },
    addChild() {
      let loop = true
      let name = ''
      const node = this.pickValue
      node.properties || (node['properties'] = {})
      const props = node.properties

      while (loop) {
        name = this._joinName()
        if (props[name] === undefined) {
          loop = false
        }
      }
      const type = 'string'
      this.hidden = false
      props[name] = { type: type }
    },
    removeNode() {
      const { properties, required } = this.parent
      delete properties[this.pickKey]
      if (required) {
        const pos = required.indexOf(this.pickKey)
        pos >= 0 && required.splice(pos, 1)
        required.length === 0 && (delete this.parent['required'])
      }
    },
    _joinName() {
      return `field_${this.deep}_${this.countAdd++}`
    },
    onSetting() {
      this.modalVisible = true
      this.advancedValue = { ...this.advanced.value }
      for (const k in this.advancedValue) {
        if (this.pickValue[k]) {
          this.advancedValue[k] = this.pickValue[k]
        }
      }
    },

    handleOk() {
      this.modalVisible = false
      for (const key in this.advancedValue) {
        if (isNull(this.advancedValue[key]) || (this.advancedValue[key] && this.advancedValue[key].length === 0)) {
          delete this.pickValue[key]
        } else {
          this.pickValue[key] = this.advancedValue[key]
        }
      }
      const diffKey = this._pickDiffKey()
      diffKey.forEach(key => delete this.pickValue[key])
    },
    _pickDiffKey() {
      const keys = Object.keys(this.pickValue)
      return keys.filter(item => this.ownProps.indexOf(item) === -1)
    }
  }
}
</script>

<style>
/* Custom styling for select dropdowns to fix WebKit appearance */
.json-schema-editor .schema-select {
  -webkit-appearance: none;
  -moz-appearance: none;
  appearance: none;
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
  background-position: right 0.5rem center;
  background-repeat: no-repeat;
  background-size: 1em 1em;
  padding-right: 2rem;
}

/* Dark mode arrow color */
.dark .json-schema-editor .schema-select {
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%239ca3af' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
}

/* Fix child indentation */
.json-schema-editor .json-schema-editor {
  margin-left: 1.5rem;
}
</style>
