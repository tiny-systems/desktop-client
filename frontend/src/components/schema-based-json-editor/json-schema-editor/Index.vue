<template>
  <div>
    <div class="grid gap-x-1 gap-y-1 grid-cols-12 mt-1 items-center">
      <div class="col-span-3 flex items-center">
        <button v-if="pickValue.type==='object'" type="button" @click="hidden = !hidden" class="flex-shrink-0">
          <chevron-right-icon v-if="hidden" class="h-3 w-3 text-sky-500 inline-block cursor-pointer"/>
          <chevron-down-icon v-else class="h-3 w-3 text-sky-500 inline-block cursor-pointer"/>
        </button>
        <input type="text" :disabled="disabled || root" :value="pickKey" @blur="onInputName"
               class="schema-input bg-transparent border border-gray-600 text-gray-300 text-xs rounded focus:ring-sky-500 focus:border-sky-500 block w-full px-2 py-1.5 dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-500 dark:text-gray-300 disabled:text-gray-500 disabled:bg-gray-800/50"
               autocomplete="off" autocapitalize="off" autocorrect="off" spellcheck="false"/>
      </div>
      <div v-if="root" class="col-span-2 flex items-center justify-center">
        <label class="text-gray-400 whitespace-nowrap flex items-center cursor-pointer">
          <input type="checkbox" :disabled="!isObject && !isArray || !parent" @change="onRootCheck" title="Required"
                 class="w-3.5 h-3.5 text-sky-600 bg-gray-700 rounded border-gray-600 focus:ring-sky-500 focus:ring-offset-gray-800 focus:ring-1"/>
          <span class="text-xs pl-1">Required</span>
        </label>
      </div>
      <div v-else class="col-span-2 flex items-center justify-center">
        <label class="text-gray-400 whitespace-nowrap flex items-center cursor-pointer">
          <input type="checkbox" :disabled="isItem || !parent" :checked="checked" @change="onCheck" title="Required"
                 class="w-3.5 h-3.5 text-sky-600 bg-gray-700 rounded border-gray-600 focus:ring-sky-500 focus:ring-offset-gray-800 focus:ring-1"/>
          <span class="text-xs pl-1">Required</span>
        </label>
      </div>
      <div class="col-span-3">
        <select v-model="pickValue.type" @change="onChangeType"
                class="type-select bg-gray-800 border border-gray-600 text-gray-300 text-xs rounded focus:ring-sky-500 focus:border-sky-500 block w-full px-2 py-1.5">
          <option v-for="t in TYPE_NAME" :key="t" :value="t">
            {{ t || '(any)' }}
          </option>
        </select>
      </div>
      <div class="col-span-3">
        <input v-model="pickValue.title" :placeholder="local['title']"
               class="schema-input bg-transparent border border-gray-600 text-gray-300 text-xs rounded focus:ring-sky-500 focus:border-sky-500 block w-full px-2 py-1.5 dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-500 dark:text-gray-300 disabled:text-gray-500 disabled:bg-gray-800/50"
               autocomplete="off" autocapitalize="off" autocorrect="off" spellcheck="false"/>
      </div>
      <div class="col-span-1 flex items-center justify-end space-x-1">
        <button v-if="isObject" type="button" @click="addChild" :title="local['add_child_node']" class="p-0.5 hover:bg-gray-700 rounded">
          <plus-icon class="w-4 h-4 text-sky-500 cursor-pointer"/>
        </button>
        <button v-if="pickValue.type" type="button" @click="onSetting" :title="local['adv_setting']" class="p-0.5 hover:bg-gray-700 rounded">
          <cog-icon class="w-4 h-4 text-sky-500 cursor-pointer"/>
        </button>
        <button v-if="!root && !isItem" type="button" @click="removeNode" :title="local['remove_node']" class="p-0.5 hover:bg-gray-700 rounded">
          <XMarkIcon class="w-4 h-4 text-red-400 cursor-pointer"/>
        </button>
      </div>
    </div>
    <template v-if="!hidden&&pickValue.properties && !isArray">
      <json-schema-editor-item v-for="(item,key,index) in pickValue.properties" :value="{[key]:item}" :parent="pickValue"
                          :key="index" :deep="deep+1" :root="false" class="ml-8" :lang="lang"/>
    </template>
    <template v-if="isArray && pickValue.items">
      <json-schema-editor-item :value="{items:pickValue.items}" :deep="deep+1" disabled isItem :root="false" class="ml-5"
                          :lang="lang"/>
    </template>

    <TransitionRoot as="template" :show="modalVisible">
      <Dialog as="div" class="relative z-50" :open="modalVisible" @close="modalVisible = false">
        <div class="fixed inset-0 bg-black/60 backdrop-blur-sm" />
        <div class="fixed inset-0 overflow-y-auto">
          <div class="flex min-h-full items-center justify-center p-4">
            <DialogPanel class="relative bg-gray-900 border border-gray-700 rounded-lg text-xs text-gray-300 p-4 shadow-xl w-full max-w-md">
              <h3 class="text-sm font-medium text-gray-200 mb-3">Advanced Settings</h3>
              <form @submit.prevent="handleOk">
                <div class="grid gap-3 md:grid-cols-2">
                  <div v-for="(item, key) in advancedValue" :key="key">
                    <label class="block text-gray-400 text-xs mb-1" :for="key">{{ local[key] || key }}</label>
                    <input type="number" v-model="advancedValue[key]" :id="key"
                           class="modal-input bg-gray-800 border border-gray-600 text-gray-300 text-xs rounded focus:ring-sky-500 focus:border-sky-500 block w-full px-2 py-1.5"
                           v-if="advancedAttr[key].type === 'integer' || advancedAttr[key].type === 'number'" :placeholder="key"/>
                    <div v-else-if="advancedAttr[key].type === 'boolean'" class="flex items-center h-8">
                      <input :id="key" type="checkbox" v-model="advancedValue[key]" class="w-4 h-4 text-sky-600 bg-gray-700 rounded border-gray-600 focus:ring-sky-500 focus:ring-offset-gray-800"/>
                    </div>
                    <textarea @blur="changeEnumValue" :value="enumText" :id="key" :rows="2" v-else-if="key === 'enum'" :placeholder="local['enum_msg']" class="modal-input bg-gray-800 border border-gray-600 text-gray-300 text-xs rounded focus:ring-sky-500 focus:border-sky-500 block w-full px-2 py-1.5 resize-none"></textarea>
                    <select :id="key" v-else-if="advancedAttr[key].type === 'array'" v-model="advancedValue[key]" :title="local[key]" class="modal-select bg-gray-800 border border-gray-600 text-gray-300 text-xs rounded focus:ring-sky-500 focus:border-sky-500 block w-full px-2 py-1.5">
                      <option value="">{{ local['none'] }}</option>
                      <option v-for="t in advancedAttr[key].enums" :key="t" :value="t">{{ t }}</option>
                    </select>
                    <input :id="key" type="text" v-model="advancedValue[key]" v-else :placeholder="key" class="modal-input bg-gray-800 border border-gray-600 text-gray-300 text-xs rounded focus:ring-sky-500 focus:border-sky-500 block w-full px-2 py-1.5"/>
                  </div>
                </div>
                <div class="pt-4 flex justify-end space-x-2">
                  <button type="button" @click="modalVisible = false" class="px-3 py-1.5 text-xs rounded border border-gray-600 text-gray-400 hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-sky-500">Cancel</button>
                  <button type="submit" class="px-3 py-1.5 text-xs rounded bg-sky-600 text-white hover:bg-sky-700 focus:outline-none focus:ring-2 focus:ring-sky-500">Save</button>
                </div>
              </form>
            </DialogPanel>
          </div>
        </div>
      </Dialog>
    </TransitionRoot>
  </div>
</template>
<script>
import {isNull} from './util'
import {TYPE, TYPE_NAME} from './type/type'
import {CheckIcon, ChevronDownIcon, ChevronRightIcon, CogIcon, PlusIcon, XMarkIcon} from '@heroicons/vue/24/outline';
import {Dialog, DialogOverlay, DialogPanel, DialogTitle, TransitionChild, TransitionRoot} from '@headlessui/vue'
import LocalProvider from './local-provider'

export default {
  name: 'JsonSchemaEditorItem',
  components: {
    ChevronRightIcon,
    ChevronDownIcon, CogIcon, PlusIcon, XMarkIcon, CheckIcon,
    Dialog, DialogOverlay, TransitionChild, TransitionRoot, DialogPanel, DialogTitle,
    // Register self for recursive rendering
    JsonSchemaEditorItem: () => import('./Index.vue')
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
    lang: { // i18n language
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
    checked() {
      return this.parent && this.parent.required && this.parent.required.indexOf(this.pickKey) >= 0
    },
    advanced() {
      return TYPE[this.pickValue.type]
    },
    advancedAttr() {
      return TYPE[this.pickValue.type].attr
    },
    advancedNotEmptyValue() {
      const jsonNode = Object.assign({}, this.advancedValue);
      for (let key in jsonNode) {
        isNull(jsonNode[key]) && delete jsonNode[key]
        if (key === 'enum' && jsonNode[key].length === 0) {
          delete jsonNode[key]
        }
      }
      return jsonNode
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
      // eslint-disable-next-line vue/no-mutating-props
      this.parent.properties[newKey] = nodeValue

      const requireds = this.parent.required || []
      const oldIndex = requireds.indexOf(oldKey)
      if (requireds.length > 0 && oldIndex > -1) {
        requireds.splice(oldIndex, 1)
        requireds.push(newKey)
        // eslint-disable-next-line vue/no-mutating-props
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
        this.pickValue['items'] = {type: 'string'}
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
          this.advancedValue.enum = arr.map(item => item);
        } else {
          this.advancedValue.enum = arr.map(item => +item);
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
        // eslint-disable-next-line vue/no-mutating-props
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
      node.properties || (node['properties'] = {}) // this.$set(node,'properties',{})
      const props = node.properties

      while (loop) {
        name = this._joinName()
        if (props[name] === undefined) {
          loop = false
        }
      }
      const type = 'string'
      this.hidden = false
      props[name] = {type: type} //this.$set(props,name,{type: type})
    },
    removeNode() {
      const {properties, required} = this.parent
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
      this.advancedValue = {...this.advanced.value}
      for (const k in this.advancedValue) {
        if (this.pickValue[k]) {
          this.advancedValue[k] = this.pickValue[k]
        }
      }
    },

    handleOk() {
      this.modalVisible = false
      for (const key in this.advancedValue) {
        if (isNull(this.advancedValue[key]) || (this.advancedValue[key] && this.advancedValue[key].length ===0)) {
          delete this.pickValue[key]
        } else {
          this.pickValue[key] = this.advancedValue[key]
        }
      }
      // Note: We intentionally do NOT delete keys that aren't in ownProps.
      // This preserves custom schema properties like 'example', 'default', '$ref', 'configurable', etc.
      // that may have been set by the backend or other schema editors.
    }
  }
}
</script>

<style scoped>
/* Fix native select dropdown styling in dark mode */
.type-select,
.modal-select {
  appearance: none;
  -webkit-appearance: none;
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
  background-position: right 0.5rem center;
  background-repeat: no-repeat;
  background-size: 1.5em 1.5em;
  padding-right: 2rem;
  color-scheme: dark;
}

.type-select option,
.modal-select option {
  background-color: #1f2937;
  color: #d1d5db;
  padding: 0.5rem;
}

/* Schema input consistent styling */
.schema-input,
.modal-input {
  background-color: #1f2937;
}

.schema-input:disabled {
  background-color: rgba(31, 41, 55, 0.5);
  cursor: not-allowed;
}

/* Ensure inputs don't have double borders */
.schema-input:focus,
.modal-input:focus,
.type-select:focus,
.modal-select:focus {
  outline: none;
}
</style>
