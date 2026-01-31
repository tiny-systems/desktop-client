<template>
  <div :class="className">
    <div :class="theme.title">
      <div class="flex space-x-1 w-full">
        <div class="pl-1" v-if="titleToShow != ''">{{titleToShow}}</div>
        <button v-if="allowEditSchema && schema.configurable && value !== undefined" @click="configureSchema()" title="Configure type" type="button" class="w-6 h-6 border border-sky-500 dark:border-sky-800  rounded p-1 button inline-block cursor-pointer hover:bg-sky-300 dark:hover:bg-sky-800">
          <PencilIcon></PencilIcon>
        </button>
      </div>
      <div :class="theme.buttonGroup" :style="buttonGroupStyle">
        <optional :required="required"
                  :value="value"
                  :isReadOnly="isReadOnly"
                  :theme="theme"
                  :locale="locale"
                  @toggleOptional="toggleOptional()">
        </optional>
      </div>
    </div>
    <div :class="['flex', deleteHover ? 'bg-red-100 dark:bg-red-900/50 border-0 rounded' : (hover ? 'bg-indigo-100 dark:bg-indigo-800 border-0 rounded' : '')]" v-if="value !== undefined">
      <div v-if="canUpload">
        Upload
      </div>
      <textarea v-if="useTextArea"
          :class="[errorMessage ? theme.errorTextarea : theme.textarea, !!expression ? theme.expression : '']"
          @change="onChange($event)" @keyup="onChange($event)" :rows="!!expression ? 3 : 10"
          :disabled="isReadOnly || !!expression" :value="expression || value"
          autocapitalize="off" autocorrect="off" spellcheck="false"></textarea>
      <vue-monaco-editor v-if="useCodeEditor"
        :class="[errorMessage ? theme.errorCodeEditor : theme.codeEditor, !!expression ? theme.expression : '', 'min-h-80']"
        :value="expression || value"
        :language="schema.language"
        :theme="codeEditorTheme"
        @change="onChange({target:{value:$event}})"
        :options="{
          formatOnType: true,
          formatOnPaste: true,
          readonly: isReadOnly || !!expression
        }"
      />
      <input v-if="useInput"
             :class="[errorMessage ? theme.errorInput : theme.input, !!expression ? theme.expression : theme.staticText]"
             :type="getInputType(schema.format || 'text')"
             :name="schema.title || 'name'"
             @change="onChange($event)"
             @keyup="onChange($event)"
             :value="expression || value"
             autocomplete="off"
             autocapitalize="off"
             autocorrect="off"
             spellcheck="false"
             :disabled="isReadOnly || !!expression"/>
      <!-- Standard select for small option lists -->
      <select v-if="useSelectComponent && !useRadioBoxComponent && !useSearchableSelect"
              :class="[errorMessage ? theme.selectError : theme.select, !!expression ? theme.expression : theme.staticText]"
              :value="!!expression ? 'expression' : value"
              :disabled="isReadOnly || !!expression"
              @change="updateSelection($event.target.value)">
        <option v-for="option in options" :key="option.value" :value="option.value">{{option.label}}</option>
        <option v-if="expression" value="expression">{{expression}}</option>
      </select>
      <!-- Searchable select for large option lists (20+ options) -->
      <div v-if="useSelectComponent && !useRadioBoxComponent && useSearchableSelect" ref="searchableSelectRef" class="relative w-full">
        <div class="relative">
          <input
            type="text"
            :class="[errorMessage ? theme.errorInput : theme.input, 'pr-8']"
            :value="searchQuery !== null ? searchQuery : selectedOptionLabel"
            :placeholder="selectedOptionLabel || 'Search...'"
            :disabled="isReadOnly || !!expression"
            @focus="openSearchableDropdown"
            @input="onSearchInput($event)"
            @keydown.escape="closeSearchableDropdown"
            @keydown.enter.prevent="selectFirstFilteredOption"
            @keydown.down.prevent="focusNextOption"
            @keydown.up.prevent="focusPrevOption"
            autocomplete="off"
            autocapitalize="off"
            autocorrect="off"
            spellcheck="false"
          />
          <button
            type="button"
            class="absolute inset-y-0 right-0 flex items-center pr-2"
            @click="toggleSearchableDropdown"
            :disabled="isReadOnly || !!expression"
          >
            <ChevronDownIcon class="h-4 w-4 text-gray-400" />
          </button>
        </div>
        <!-- Teleport dropdown to body to escape overflow clipping -->
        <Teleport to="body">
          <div
            v-if="isSearchableDropdownOpen"
            :style="dropdownStyle"
            class="fixed z-[9999] bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-md shadow-lg max-h-60 overflow-auto"
          >
            <div
              v-for="(option, index) in filteredOptions"
              :key="option.value"
              :class="[
                'px-3 py-2 cursor-pointer text-sm',
                option.value === value ? 'bg-sky-100 dark:bg-sky-900 text-sky-900 dark:text-sky-100' : 'hover:bg-gray-100 dark:hover:bg-gray-800',
                focusedOptionIndex === index ? 'bg-gray-100 dark:bg-gray-800' : ''
              ]"
              @click="selectSearchableOption(option.value)"
              @mouseenter="focusedOptionIndex = index"
            >
              {{ option.label }}
            </div>
            <div v-if="filteredOptions.length === 0" class="px-3 py-2 text-sm text-gray-500">
              No options found
            </div>
          </div>
        </Teleport>
      </div>
      <div v-if="useRadioBoxComponent && !expression" class="w-full">
        <span v-for="option in options" :key="option.value" :class="theme.radiobox">
          <label :class="theme.label">
            <input type="radio"
                   @change="updateSelection(option.value)"
                   :checked="value === option.value"
                   :disabled="isReadOnly" />
            {{option.label}}
          </label>
        </span>
      </div>
      <div v-if="useRadioBoxComponent && !!expression" class="w-full">
        <label :class="theme.expression">
          <input type="radio" :checked="true" name="expression" :readonly="true" />
          {{expression}}
        </label>
      </div>
      <img v-if="willPreviewImage"
           :class="theme.img"
           :style="imagePreviewStyle"
           :src="getImageUrl" />

      <button  v-if="hasDeleteButtonFunction" type="button" class="ml-1 w-4 text-indigo-500 inline-block cursor-pointer ml-1"  @click="$emit('delete')" title="Delete">
        <XCircleIcon></XCircleIcon>
      </button>
      <button v-if="allowLookup && value !== undefined && !isMixedExpression" @mouseover="hover = true" @mouseleave="hover = false" type="button" class="w-4 block text-indigo-500 cursor-pointer mx-1" @click="$emit('lookup', getAllValue(), schema, onChangeExpression)" :title="expression ? 'Edit expression' : 'Apply expression'">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" >
          <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0021 18V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v12a2.25 2.25 0 002.25 2.25z" />
        </svg>
      </button>
      <button v-if="expression" type="button" class="w-4 block text-red-500 cursor-pointer mx-1" @mouseover="deleteHover = true" @mouseleave="deleteHover = false" @click="clearExpression" title="Clear expression">
        <XCircleIcon></XCircleIcon>
      </button>
    </div>
    <description :theme="theme" :message="schema.description" ></description>
    <description :theme="theme" :message="errorMessage" :error="true"></description>
    <!-- Expression chips for mixed expressions -->
    <div v-if="allowLookup && embeddedExpressions.length > 0" class="flex flex-wrap gap-1 mt-1">
      <button
        v-for="(expr, idx) in embeddedExpressions"
        :key="idx"
        @click="editEmbeddedExpression(idx)"
        type="button"
        class="text-xs px-2 py-0.5 bg-indigo-100 dark:bg-indigo-900/50 text-indigo-700 dark:text-indigo-300 rounded hover:bg-indigo-200 dark:hover:bg-indigo-800 font-mono"
        :title="'Edit expression: ' + expr"
      >
        {{ expr.length > 30 ? expr.substring(0, 30) + '...' : expr }}
      </button>
    </div>
  </div>
</template>
<script lang="ts">
import type { PropType } from 'vue'
import { defineAsyncComponent } from 'vue'
import * as common from './common'
import {XCircleIcon, ChevronRightIcon, ChevronDownIcon, PencilIcon} from '@heroicons/vue/24/outline'
import Optional from './Optional.vue'
import Description from './Description.vue'
import { useDark } from "@vueuse/core";

// Lazy load monaco editor to avoid SSR issues
const VueMonacoEditor = defineAsyncComponent(() =>
  import('@guolao/vue-monaco-editor').then(m => m.VueMonacoEditor)
)

export default {
  components: {
    XCircleIcon, ChevronRightIcon, ChevronDownIcon, PencilIcon,
    optional: Optional,
    description: Description,
    'vue-monaco-editor': VueMonacoEditor
  },
  emits: ['delete', 'update-value', 'lookup'],
  props: {
    schema: {
      type: Object as PropType<common.StringSchema>,
      required: true,
    },
    allowLookup: Boolean,
    plainStruct: Boolean,
    initialValue: null,
    allowEditSchema: Boolean,
    title: [String, Number],
    theme: {
      type: Object as PropType<common.Theme>,
      required: true,
    },
    locale: {
      type: Object as PropType<common.Locale>,
      required: true,
    },
    readonly: Boolean,
    required: Boolean,
    hasDeleteButton: {
      type: Boolean,
      required: false
    }
  },
  data: () => {
    return {
      codeEditorTheme: 'vs',
      value: '' as string | undefined,
      expression: undefined,
      portName: null,
      hover: false,
      deleteHover: false,
      errorMessage: '' as string | undefined,
      buttonGroupStyle: common.buttonGroupStyleString,
      imagePreviewStyle: common.imagePreviewStyleString,
      // Searchable select state
      searchQuery: null as string | null,
      isSearchableDropdownOpen: false,
      focusedOptionIndex: -1,
      dropdownPosition: { top: 0, left: 0, width: 0 }
    }
  },
  beforeMount() {
    this.value = this.getValue()
    // Extract expression from {{expr}} pattern
    this.expression = this.extractExpression(this.initialValue)
    this.validate()
    if (this.value !== undefined) {
      this.emitValue()
    }
  },
  mounted() {
    const isDark = useDark({
      onChanged: (dark) => {
        if (dark) {
          this.codeEditorTheme = 'vs-dark'
        } else {
          this.codeEditorTheme = 'vs'
        }
      }
    })
    // Add click outside listener for searchable dropdown
    document.addEventListener('click', this.handleClickOutside)
  },
  beforeUnmount() {
    document.removeEventListener('click', this.handleClickOutside)
  },
  methods: {
    configureSchema() {
      this.schema.configure = true
    },
    getValue():string {
      // For new {{expr}} format, extract actual value (not the expression wrapper)
      if (typeof this.initialValue === 'string') {
        const match = this.initialValue.match(/^\{\{(.+)\}\}$/)
        if (match) {
          // It's an expression - return empty string so field shows as "defined"
          // The actual expression is stored in this.expression
          return ''
        }
        return this.initialValue
      }
      return common.getDefaultValue(this.required, this.schema, this.initialValue) as string
    },
    getInputType(t: string):string {
      if (t === 'date-time') {
        return 'text'
      }
      return t
    },
    onChange(e: { target: { value: string } }) {
      this.value = e.target.value
      this.validate()
      this.emitValue()
    },
    onChangeExpression(dataExpression: string, portName: string) {
      this.expression = dataExpression
      this.portName = portName
      this.validate()
      this.emitValue()
    },
    updateSelection(value: string) {
      this.value = value
      this.validate()
      this.emitValue()
    },
    toggleOptional() {
      // Don't pass initialValue if it's an expression - we don't want the expression string as fallback value
      const fallbackValue = this.extractExpression(this.initialValue) ? undefined : this.initialValue
      this.value = common.toggleOptional(this.value, this.schema, fallbackValue) as string | undefined
      this.validate()
      this.emitValue()
    },
    emitValue() {
      this.$emit('update-value', { value: this.getAllValue(), isValid: !this.errorMessage })
    },
    getAllValue() {
      if(this.value === undefined || this.plainStruct) {
        return this.value
      }
      // New format: expression wrapped in {{expr}}, literals are plain values
      if (this.expression) {
        return `{{${this.expression}}}`
      }
      return this.value || ''
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
      this.validate()
      this.emitValue()
    },
    editEmbeddedExpression(idx: number) {
      const expressions = this.embeddedExpressions
      if (idx < 0 || idx >= expressions.length) return

      const exprToEdit = expressions[idx]
      // Use simple string schema for embedded expressions
      const stringSchema = { type: 'string' }

      // Callback to replace this specific expression in the value
      const onUpdate = (newExpr: string) => {
        if (!newExpr) return
        const currentValue = this.value || ''
        // Find and replace the nth occurrence
        let count = 0
        const newValue = currentValue.replace(/\{\{(.+?)\}\}/g, (match, expr) => {
          if (count === idx) {
            count++
            return `{{${newExpr}}}`
          }
          count++
          return match
        })
        this.value = newValue
        this.validate()
        this.emitValue()
      }

      this.$emit('lookup', `{{${exprToEdit}}}`, stringSchema, onUpdate)
    },
    validate() {
      if (!!this.expression || this.isReadOnly) {
        this.errorMessage = ''
        return;
      }
      this.errorMessage = common.getErrorMessageOfString(this.value, this.schema, this.required,  this.locale)
    },
    // Searchable select methods
    openSearchableDropdown() {
      // Calculate position before opening
      const ref = this.$refs.searchableSelectRef as HTMLElement | undefined
      if (ref) {
        const rect = ref.getBoundingClientRect()
        this.dropdownPosition = {
          top: rect.bottom + window.scrollY,
          left: rect.left + window.scrollX,
          width: rect.width
        }
      }
      this.isSearchableDropdownOpen = true
      this.searchQuery = ''
      this.focusedOptionIndex = -1
    },
    closeSearchableDropdown() {
      this.isSearchableDropdownOpen = false
      this.searchQuery = null
      this.focusedOptionIndex = -1
    },
    toggleSearchableDropdown() {
      if (this.isSearchableDropdownOpen) {
        this.closeSearchableDropdown()
      } else {
        this.openSearchableDropdown()
      }
    },
    onSearchInput(e: { target: { value: string } }) {
      this.searchQuery = e.target.value
      this.focusedOptionIndex = 0
    },
    selectSearchableOption(optionValue: string) {
      this.updateSelection(optionValue)
      this.closeSearchableDropdown()
    },
    selectFirstFilteredOption() {
      if (this.filteredOptions.length > 0) {
        const index = this.focusedOptionIndex >= 0 ? this.focusedOptionIndex : 0
        this.selectSearchableOption(this.filteredOptions[index].value as string)
      }
    },
    focusNextOption() {
      if (this.focusedOptionIndex < this.filteredOptions.length - 1) {
        this.focusedOptionIndex++
      }
    },
    focusPrevOption() {
      if (this.focusedOptionIndex > 0) {
        this.focusedOptionIndex--
      }
    },
    handleClickOutside(e: MouseEvent) {
      if (!this.isSearchableDropdownOpen) return
      const target = e.target as Node
      // Check if click is inside the component
      if (this.$el && this.$el.contains(target)) return
      // Check if click is inside the teleported dropdown (has z-[9999] class)
      const dropdown = document.querySelector('.fixed.z-\\[9999\\]')
      if (dropdown && dropdown.contains(target)) return
      this.closeSearchableDropdown()
    }
  },
  computed: {
    dropdownStyle(): { top: string; left: string; width: string } {
      return {
        top: `${this.dropdownPosition.top}px`,
        left: `${this.dropdownPosition.left}px`,
        width: `${this.dropdownPosition.width}px`
      }
    },
    canPreviewImage(): boolean {
      return common.isImageUrl(this.value + '') || common.isBase64Image(this.value + '')
    },
    canPreview(): boolean | undefined {
      return (!!this.value) && (this.canPreviewImage)
    },
    useTextArea(): boolean | undefined {
      return this.value !== undefined
        && (this.schema.enum === undefined || this.isReadOnly)
        && (this.schema.format === 'textarea')
    },
    useCodeEditor(): boolean | undefined {
      return this.value !== undefined
        && (this.schema.enum === undefined || this.isReadOnly)
        && (this.schema.format === 'code' || this.schema.format === 'json')
    },
    useDatePicker(): boolean | undefined {
      return this.value !== undefined && this.schema.format ==='date-time'
    },
    useInput(): boolean | undefined {
      return this.value !== undefined
        && (this.schema.enum === undefined || this.isReadOnly)
        && !this.useTextArea
        && !this.useCodeEditor
    },
    useSelect(): boolean {
      return this.value !== undefined && this.schema.enum !== undefined && !this.isReadOnly
    },
    useSelectComponent(): boolean {
      return this.useSelect
    },
    useRadioBoxComponent(): boolean {
      return this.useSelect && this.schema.format === 'radiobox'
    },
    useSearchableSelect(): boolean {
      // Always use styled dropdown instead of native select
      return true
    },
    filteredOptions(): { value: string | number; label: string | number }[] {
      if (!this.searchQuery) return this.options
      const query = this.searchQuery.toLowerCase()
      return this.options.filter(opt =>
        String(opt.label).toLowerCase().includes(query) ||
        String(opt.value).toLowerCase().includes(query)
      )
    },
    selectedOptionLabel(): string {
      const selected = this.options.find(opt => opt.value === this.value)
      return selected ? String(selected.label) : ''
    },
    getImageUrl(): string | undefined {
      return this.value
    },
    isReadOnly(): boolean | undefined {
      return this.readonly || this.schema.readonly
    },
    hasDeleteButtonFunction(): boolean {
      return this.hasDeleteButton && !this.isReadOnly
    },
    willPreviewImage(): boolean | '' | undefined {
      return this.value && this.canPreviewImage
    },
    titleToShow(): string {
      return common.getTitle(this.schema.title, this.title)
    },
    options(): {
      value: string | number;
      label: string | number;
    }[] {
      return common.getOptions(this.schema)
    },
    canUpload(): boolean {
      return this.schema.format === 'base64'
    },
    className(): string {
      const rowClass = this.errorMessage ? this.theme.errorRow : this.theme.row
      return this.schema.className ? rowClass + ' ' + this.schema.className : rowClass
    },
    isMixedExpression(): boolean {
      // Mixed expression: value contains {{...}} but isn't a pure expression
      // e.g., "Hello {{$.name}}" or '{"text": "{{$.error}}"}'
      if (this.expression) return false // Pure expression, not mixed
      const val = this.initialValue
      if (typeof val !== 'string') return false
      return val.includes('{{') && val.includes('}}')
    },
    embeddedExpressions(): string[] {
      // Extract all {{...}} expressions from mixed content
      if (!this.isMixedExpression) return []
      const val = this.value || ''
      const matches: string[] = []
      const regex = /\{\{(.+?)\}\}/g
      let match
      while ((match = regex.exec(val)) !== null) {
        matches.push(match[1])
      }
      return matches
    },
  }
}
</script>
