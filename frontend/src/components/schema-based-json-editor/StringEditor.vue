<template>
  <div :class="className">
    <div :class="theme.title">
      <div class="flex space-x-1 w-full">
        <div class="pl-1" v-if="titleToShow != ''">{{titleToShow}}</div>
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
    <div class="flex" v-if="value !== undefined">
      <textarea v-if="useTextArea"
          :class="[errorMessage ? theme.errorTextarea : theme.textarea]"
          @change="onChange($event)" @keyup="onChange($event)" rows="10"
          :disabled="isReadOnly" :value="value"></textarea>
      <vue-monaco-editor v-if="useCodeEditor"
        :class="[errorMessage ? theme.errorCodeEditor : theme.codeEditor, 'min-h-80']"
        :value="value"
        :language="schema.language"
        :theme="codeEditorTheme"
        @change="onChange({target:{value:$event}})"
        :options="{
          formatOnType: true,
          formatOnPaste: true,
          readonly: isReadOnly
        }"
      />
      <input v-if="useInput"
             :class="[errorMessage ? theme.errorInput : theme.input]"
             :type="getInputType(schema.format || 'text')"
             :name="schema.title || 'name'"
             @change="onChange($event)"
             @keyup="onChange($event)"
             :value="value"
             autocomplete="off"
             :disabled="isReadOnly"/>
      <!-- Standard select for small option lists -->
      <select v-if="useSelectComponent && !useRadioBoxComponent && !useSearchableSelect"
              :class="[errorMessage ? theme.selectError : theme.select]"
              :value="value"
              :disabled="isReadOnly"
              @change="updateSelection($event.target.value)">
        <option v-for="option in options" :key="option.value" :value="option.value">{{option.label}}</option>
      </select>
      <!-- Searchable select for large option lists (20+ options) -->
      <div v-if="useSelectComponent && !useRadioBoxComponent && useSearchableSelect" ref="searchableSelectRef" class="relative w-full">
        <div class="relative">
          <input
            type="text"
            :class="[errorMessage ? theme.errorInput : theme.input, 'pr-8']"
            :value="searchQuery !== null ? searchQuery : selectedOptionLabel"
            :placeholder="selectedOptionLabel || 'Search...'"
            :disabled="isReadOnly"
            @focus="openSearchableDropdown"
            @input="onSearchInput($event)"
            @keydown.escape="closeSearchableDropdown"
            @keydown.enter.prevent="selectFirstFilteredOption"
            @keydown.down.prevent="focusNextOption"
            @keydown.up.prevent="focusPrevOption"
          />
          <button
            type="button"
            class="absolute inset-y-0 right-0 flex items-center pr-2"
            @click="toggleSearchableDropdown"
            :disabled="isReadOnly"
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
      <div v-if="useRadioBoxComponent" class="w-full">
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
      <img v-if="willPreviewImage"
           :class="theme.img"
           :style="imagePreviewStyle"
           :src="getImageUrl" />

      <button v-if="hasDeleteButtonFunction" type="button" class="ml-1 w-4 text-sky-500 inline-block cursor-pointer ml-1" @click="$emit('delete')" title="Delete">
        <XCircleIcon></XCircleIcon>
      </button>
    </div>
    <description :theme="theme" :message="schema.description" ></description>
    <description :theme="theme" :message="errorMessage" :error="true"></description>
  </div>
</template>
<script lang="ts">
import type { PropType } from 'vue'
import { defineAsyncComponent } from 'vue'
import * as common from './common'
import { XCircleIcon, ChevronDownIcon } from '@heroicons/vue/24/outline'
import Optional from './Optional.vue'
import Description from './Description.vue'

// Lazy load monaco editor
const VueMonacoEditor = defineAsyncComponent(() =>
  import('@guolao/vue-monaco-editor').then(m => m.VueMonacoEditor)
)

export default {
  components: {
    XCircleIcon, ChevronDownIcon,
    optional: Optional,
    description: Description,
    'vue-monaco-editor': VueMonacoEditor
  },
  emits: ['delete', 'update-value'],
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
    this.validate()
    if (this.value !== undefined) {
      this.emitValue()
    }
  },
  mounted() {
    // Check for dark mode preference
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      this.codeEditorTheme = 'vs-dark'
    }
    // Add click outside listener for searchable dropdown
    document.addEventListener('click', this.handleClickOutside)
  },
  beforeUnmount() {
    document.removeEventListener('click', this.handleClickOutside)
  },
  methods: {
    getValue():string {
      if (typeof this.initialValue === 'string') {
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
    updateSelection(value: string) {
      this.value = value
      this.validate()
      this.emitValue()
    },
    toggleOptional() {
      this.value = common.toggleOptional(this.value, this.schema, this.initialValue) as string | undefined
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
      return this.value || ''
    },
    validate() {
      if (this.isReadOnly) {
        this.errorMessage = ''
        return;
      }
      this.errorMessage = common.getErrorMessageOfString(this.value, this.schema, this.required, this.locale)
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
      return this.options.length > 20
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
    className(): string {
      const rowClass = this.errorMessage ? this.theme.errorRow : this.theme.row
      return this.schema.className ? rowClass + ' ' + this.schema.className : rowClass
    },
  }
}
</script>
