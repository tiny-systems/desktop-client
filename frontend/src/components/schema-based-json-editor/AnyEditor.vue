<template>
  <div :class="className">
    <div :class="theme.title">
      <div class="flex space-x-1 w-full">
        <div class="pl-1" v-if="titleToShow != ''">{{titleToShow}}</div>
        <div v-if="(allowEditSchema && schema.configurable) || schema.type">
          <div class="flex space-x-1 w-full relative">
            <button v-if="allowEditSchema && schema.configurable && !schema.configure && value !== undefined && !isReadOnly" @click="configureSchema()" title="Configure type" type="button" class="w-6 h-6 border border-sky-500 dark:border-sky-800  rounded p-1 button inline-block cursor-pointer hover:bg-sky-300 dark:hover:bg-sky-800">
              <PencilIcon></PencilIcon>
            </button>
            <button @click="applySchema" type="button" class="w-6 h-6 border border-sky-500 dark:border-sky-800  rounded p-1 button inline-block cursor-pointer hover:bg-sky-300 dark:hover:bg-sky-800" title="Apply schema changes" v-if="schema.configure && schema.configurable && allowEditSchema && dirty">
              <ArrowDownOnSquareIcon></ArrowDownOnSquareIcon>
            </button>
            <Popover>
              <PopoverButton as="button" type="button" class="w-6 h-6 border border-sky-500 dark:border-sky-800  rounded p-1 button inline-block cursor-pointer hover:bg-sky-300 dark:hover:bg-sky-800" title="Show schema" v-if="schema.configure && allowEditSchema">
                <CodeBracketIcon></CodeBracketIcon>
              </PopoverButton>
              <transition
                enter-active-class="transition duration-200 ease-out"
                enter-from-class="translate-y-1 opacity-0"
                enter-to-class="translate-y-0 opacity-100"
                leave-active-class="transition duration-150 ease-in"
                leave-from-class="translate-y-0 opacity-100"
                leave-to-class="translate-y-1 opacity-0">
                <PopoverPanel as="div" class="flex flex-col overflow-hidden top-6 h-60 p-3 min-w-fit absolute z-5s0 w-64 text-xs font-light text-gray-500 bg-white rounded-lg border border-gray-200 shadow-sm opacity-0 transition-opacity duration-300 dark:text-gray-300 dark:border-gray-600 dark:bg-gray-800">
                  <textarea class="text-xs w-full h-full rounded-sm border border-sky-500 focus:ring-0 resize-none dark:bg-black">{{schemaEdit}}</textarea>
                  <button type="button" @click="copyContent(JSON.stringify(schemaEdit))" class="w-full mt-2 h-6 border border-sky-500 dark:border-sky-800 rounded-sm p-1 button inline-block cursor-pointer hover:bg-sky-300 dark:hover:bg-sky-800">Copy JSON schema</button>
                </PopoverPanel>
              </transition>
            </Popover>
            <Popover>
              <PopoverButton as="button" type="button" class="w-6 h-6 border border-sky-500 dark:border-sky-800  rounded p-1 button inline-block cursor-pointer hover:bg-sky-300 dark:hover:bg-sky-800" title="Import schema" v-if="schema.configure && allowEditSchema">
                <DocumentArrowUpIcon></DocumentArrowUpIcon>
              </PopoverButton>
              <transition
                enter-active-class="transition duration-200 ease-out"
                enter-from-class="translate-y-1 opacity-0"
                enter-to-class="translate-y-0 opacity-100"
                leave-active-class="transition duration-150 ease-in"
                leave-from-class="translate-y-0 opacity-100"
                leave-to-class="translate-y-1 opacity-0">
                <PopoverPanel as="div" class="flex flex-col overflow-hidden top-6 h-60 p-3 min-w-fit absolute z-5s0 w-64 text-xs font-light text-gray-500 bg-white rounded-lg border border-gray-200 shadow-sm opacity-0 transition-opacity duration-300 dark:text-gray-300 dark:border-gray-600 dark:bg-gray-800">
                  <textarea v-model="localChangesStr" class="text-xs w-full h-full rounded-sm border border-sky-500 focus:ring-0 resize-none dark:bg-black"></textarea>
                  <button type="button" @click="importLocalChanges" class="w-full mt-2 h-6 border border-sky-500 dark:border-sky-800 rounded-sm p-1 button inline-block cursor-pointer hover:bg-sky-300 dark:hover:bg-sky-800">Import JSON schema</button>
                </PopoverPanel>
              </transition>
            </Popover>
            <button @click="discardChanges" type="button" class="w-6 h-6 border border-sky-500 dark:border-sky-800  rounded p-1 button inline-block cursor-pointer hover:bg-sky-300 dark:hover:bg-sky-800" title="Discard schema changes" v-if="schema.configure">
              <XMarkIcon></XMarkIcon>
            </button>
          </div>
        </div>
      </div>
      <div :class="theme.buttonGroup" :style="buttonGroupStyle">
        <optional :required="required" :value="value" :isReadOnly="isReadOnly"
                  :theme="theme" :locale="locale" @toggleOptional="toggleOptional()">
        </optional>
      </div>
      <div :class="theme.buttonGroup">
        <button v-if="hasDeleteButtonFunction" type="button" class="w-4 text-sky-500 cursor-pointer ml-1" @click="$emit('delete')" title="Delete">
          <XCircleIcon></XCircleIcon>
        </button>
      </div>
    </div>
    <div v-if="allowLookup && (value !== undefined) && !isReadOnly" :class="['w-full text-left flex justify-between', deleteHover ? 'bg-red-100 dark:bg-red-900/50 border-0 rounded' : (hover ? 'bg-indigo-100 dark:bg-indigo-800 border-0 rounded' : '')]">
      <label class="text-indigo-500 text-xs pb-1 pl-1">{{expression}}</label>
      <div class="flex">
        <button @mouseover="hover = true" @mouseleave="hover = false" type="button" class="w-4 inline-block text-indigo-500 cursor-pointer" :title="expression ? 'Edit expression' : 'Apply expression'" @click="$emit('lookup', getAllValue(), schema, onChangeExpression)">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" >
            <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0021 18V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v12a2.25 2.25 0 002.25 2.25z" />
          </svg>
        </button>
        <button v-if="expression" type="button" class="w-4 inline-block text-red-500 cursor-pointer" @mouseover="deleteHover = true" @mouseleave="deleteHover = false" title="Clear expression" @click="clearExpression">
          <XCircleIcon></XCircleIcon>
        </button>
      </div>
    </div>
    <div v-if="allowEditSchema && schema.configure" class="border-2 rounded p-1 border-sky-500 dark:border-sky-800">
      <span class="text-xs dark:text-gray-500">Create your own schema.</span>
        <json-schema-editor class="mt-1" :value="schemaEdit" :root="true"></json-schema-editor>
        <p class="text-xs text-sky-400 pt-5" v-if="dirty">Don't forget to apply your recent changes of schema.</p>
    </div>
    <description v-if="(allowEditSchema && schema.configurable) || schema.type" :theme="theme" :message="schema.description"></description>
    <code v-if="!schema.configure && !!expression" class="text-xs p-1 bg-indigo-50 dark:bg-indigo-900/50 text-indigo-600 dark:text-indigo-300 w-full block font-mono">{{"{{" + expression + "}}"}}</code>
    <vue-json-pretty v-else-if="!schema.configure && isObject" :data="value" :deep="2" :show-length="true" class="text-xs p-1 bg-blue-50 dark:bg-gray-800 w-full block overflow-auto max-h-96" />
    <code v-else-if="!schema.configure" class="text-xs p-1 bg-blue-50 dark:bg-blue-500 w-full block">{{value}}</code>
  </div>
</template>
<script lang="ts">
import {ref} from 'vue'
import type {PropType} from 'vue'
import * as common from './common'
import Optional from './Optional.vue'
import Description from './Description.vue'
import JsonSchemaEditor from './JsonSchemaEditor.vue'
import VueJsonPretty from 'vue-json-pretty'
import 'vue-json-pretty/lib/styles.css'
import {
  XCircleIcon,
  XMarkIcon,
  ArrowDownOnSquareIcon,
  EyeIcon,
  CodeBracketIcon,
  DocumentArrowUpIcon,
  ChevronRightIcon,
  ChevronDownIcon
} from '@heroicons/vue/24/outline'
import { PencilIcon } from '@heroicons/vue/24/solid'
import {
  Popover,
  PopoverButton,
  PopoverPanel
} from '@headlessui/vue'

export default {
  emits: ['update-value', 'delete', 'lookup'],
  props: {
    schema: {
      required: true,
    },
    property: String,
    initialValue: null,
    title: [String, Number],
    theme: {
      type: Object as PropType<common.Theme>,
      required: true,
    },
    locale: {
      type: Object as PropType<common.Locale>,
      required: true,
    },
    allowLookup: Boolean,
    plainStruct: Boolean,
    readonly: Boolean,
    required: Boolean,
    allowEditSchema: Boolean,
    hasDeleteButton: {
      type: Boolean,
      required: false,
    }
  },
  data: () => {
    return {
      localChanges: null,
      dirty: false,
      expression: undefined,
      hover: false,
      deleteHover: false,
      importChanges: null,
      value: undefined as common.ValueType | undefined,
      buttonGroupStyle: common.buttonGroupStyleString,
    }
  },
  computed: {
    className(): string {
      const rowClass = this.theme.row
      return this.schema.className ? rowClass + ' ' + this.schema.className : rowClass
    },
    titleToShow(): string {
      return common.getTitle(this.schema.title, this.title)
    },
    isReadOnly(): boolean | undefined {
      return this.readonly || this.schema.readonly
    },
    hasDeleteButtonFunction(): boolean {
      return this.hasDeleteButton && !this.isReadOnly
    },
    isObject(): boolean {
      return this.value !== null && this.value !== undefined && typeof this.value === 'object'
    },
    localChangesStr: {
      get() {
        return JSON.stringify(this.localChanges)
      },
      set(v) {
        this.importChanges = JSON.parse(v)
      }
    },
    schemaEdit: {
      get() {
        if (this.localChanges) {
          return this.localChanges
        }
        const s = {}
        //@ts-ignore
        s[this.property] = Object.assign({}, this.schema);
        //@ts-ignore
        this.localChanges = s
        return this.localChanges
      },
      set(v) {
        this.localChanges = v
      },
    },
  },
  methods: {
    importLocalChanges() {
      this.localChanges = this.importChanges
    },
    discardChanges() {
      //@ts-ignore
      this.localchanges = null
      //@ts-ignore
      Object.assign(this.schema,{configure: false})
      //@ts-ignore
      this.dirty = false
    },
    configureSchema() {
      //@ts-ignore
      this.schema.configure = true
      //@ts-ignore
      this.dirty = false
    },
    async copyContent (text: string) {
      try {
        await navigator.clipboard.writeText(text);
        console.log('content copied to clipboard');
        /* Resolved - text copied to clipboard successfully */
      } catch (err) {
        console.error('failed to copy: ', err);
      }
    },
    applySchema() {
      //@ts-ignore
      const val = Object.values(this.localChanges)[0]
      //@todo replace with emit
      //@ts-ignore
      const typeBefore = this.schema ? this.schema.type : ''
      //@ts-ignore
      Object.assign(this.schema, val, {configure: false})

      //@ts-ignore
      if (typeBefore != this.schema.type) {
        //reset value only if type changed
        this.$emit('update-value', {value: undefined, isValid: true})
      }
      //@ts-ignore
      if (this.schema && this.schema.type === '') {
        // empty schema means any
        //@ts-ignore
        delete this.schema['type']
      }
      //@ts-ignore
      this.dirty = false
    },
    //@ts-ignore
    getAllValue() {
      if (this.value === undefined) {
        return this.value
      }
      // New format: expression wrapped in {{expr}}, literals are plain values
      // Always return expression wrapper when there's an expression (needed for lookup)
      //@ts-ignore
      if (this.expression) {
        //@ts-ignore
        return `{{${this.expression}}}`
      }
      if (this.plainStruct) {
        return this.value
      }
      //@ts-ignore
      return this.value
    },
    emitValue() {
      //@ts-ignore
      this.$emit('update-value', { value: this.getAllValue(), isValid: true })
    },
    onChangeExpression(dataExpression: string) {
      //@ts-ignore
      this.expression = dataExpression
      this.emitValue()
    },
    getValue(): any {
      // For new {{expr}} format, extract actual value (not the expression wrapper)
      if (typeof this.initialValue === 'string') {
        const match = this.initialValue.match(/^\{\{(.+)\}\}$/)
        if (match) {
          // It's an expression - return empty object so field shows as "defined"
          // The actual expression is stored in this.expression
          return {}
        }
      }
      return common.getDefaultValue(this.required, this.schema, this.initialValue)
    },
    extractExpression(val: any): string | undefined {
      if (typeof val === 'string') {
        // New format: {{expression}}
        const match = val.match(/^\{\{(.+)\}\}$/)
        if (match) {
          return match[1]
        }
      }
      return undefined
    },
    clearExpression() {
      this.expression = undefined
      this.emitValue()
    }
  },
  beforeMount() {
    //@ts-ignore
    this.expression = this.extractExpression(this.initialValue)
    //@ts-ignore
    this.value = this.getValue()
  },
  watch: {
    schemaEdit: {
      //@ts-ignore
      handler(newValue, oldValue) {
        //@ts-ignore
        this.dirty = true
      },
      deep: true
    }
  },
  components: {
    ChevronDownIcon, ChevronRightIcon,
    XCircleIcon, XMarkIcon, ArrowDownOnSquareIcon, EyeIcon, CodeBracketIcon, PencilIcon, DocumentArrowUpIcon,
    Popover, PopoverPanel, PopoverButton,
    Optional,
    Description,
    JsonSchemaEditor,
    VueJsonPretty,
  }
}
</script>
