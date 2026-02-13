<template>
  <div :class="className">
    <div :class="theme.title">
      <div class="flex space-x-1 w-full">
        <div class="pl-1" v-if="titleToShow != ''">{{titleToShow}}</div>
        <button v-if="allowEditSchema && schema.configurable && value !== undefined" @click="configureSchema()" title="Configure type" type="button" class="w-6 h-6 border border-sky-500 dark:border-sky-800  rounded p-1 button inline-block cursor-pointer hover:bg-sky-300 dark:hover:bg-sky-800">
          <PencilSquareIcon></PencilSquareIcon>
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
      <input v-if="useInput"
       :class="[errorMessage ? theme.errorInput : theme.input, !!expression ? theme.expression : theme.staticText]"
       :type="expression ? 'text' : 'number'"
       @change="onChange($event)"
       @keyup="onChange($event)"
       :value="expression || value"
       :disabled="isReadOnly || !!expression"
       :step="step"
      />
      <select v-if="useSelectComponent"
              :class="[theme.select, !!expression ? theme.expression : theme.staticText]"
              :value="!!expression ? 'expression' : value"
              :disabled="isReadOnly || !!expression"
              @change="updateSelection(+$event.target.value)">
        <option v-if="!!expression" value="expression">{{expression}}</option>
        <option v-for="option in options" :key="option.value" :value="option.value">{{option.label}}</option>
      </select>
      <div v-if="useRadioBoxComponent && !!expression" class="w-full">
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
      <button type="button" @mouseover="hover = true" @mouseleave="hover = false" class="mx-1 w-4 text-indigo-500 block cursor-pointer" v-if="allowLookup && value !== undefined" @click="$emit('lookup', getAllValue(), schema, onChangeExpression)" :title="expression ? 'Edit expression' : 'Apply expression'">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0021 18V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v12a2.25 2.25 0 002.25 2.25z" />
        </svg>
      </button>
      <button v-if="expression" type="button" class="mx-1 w-4 text-red-500 block cursor-pointer" @mouseover="deleteHover = true" @mouseleave="deleteHover = false" @click="clearExpression" title="Clear expression">
        <XCircleIcon></XCircleIcon>
      </button>
      <button  v-if="hasDeleteButtonFunction" type="button" class="ml-1 w-4 text-indigo-500 inline-block cursor-pointer ml-1"  @click="$emit('delete')" title="Delete">
        <XCircleIcon></XCircleIcon>
      </button>
    </div>
    <description :theme="theme" :message="schema.description"></description>
    <description :theme="theme" :message="errorMessage" :error="true"></description>
  </div>

</template>
<script lang="ts">
import type { PropType } from 'vue'
import * as common from './common'
import Optional from './Optional.vue'
import Description from './Description.vue'
import {PencilSquareIcon, XCircleIcon} from '@heroicons/vue/24/outline'

export default {
  emits: ['update-value', 'delete', 'lookup'],
  components: {
    XCircleIcon, PencilSquareIcon,
    optional: Optional,
    description: Description
  },
  props: {
    schema: {
      type: Object as PropType<common.NumberSchema>,
      required: true,
    },
    allowEditSchema: Boolean,
    allowLookup: Boolean,
    plainStruct: Boolean,
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
    readonly: Boolean,
    required: Boolean,
    hasDeleteButton: {
      type: Boolean,
      required: false
    }
  },
  data: () => {
    return {
      hover: false,
      deleteHover: false,
      value: 0 as number | undefined,
      expression: undefined,
      errorMessage: '' as string | undefined,
      buttonGroupStyle: common.buttonGroupStyleString,
    }
  },
  beforeMount() {
    //@ts-ignore
    this.value = this.getValue()
    //@ts-ignore
    this.expression = this.extractExpression(this.initialValue)
    //@ts-ignore
    this.validate()
    //@ts-ignore
    if (this.value !== undefined) {
      //@ts-ignore
      this.emitValue()
    }
  },
  computed: {
    useInput(): boolean | undefined {
      //@ts-ignore
      return this.value !== undefined && (this.schema.enum === undefined || this.isReadOnly)
    },
    useSelect(): boolean {
      //@ts-ignore
      return this.value !== undefined && (this.schema.enum !== undefined && !this.isReadOnly)
    },
    useSelectComponent(): boolean {
      //@ts-ignore
      return this.useSelect && (this.schema.format === 'select' || this.schema.format === undefined)
    },
    useRadioBoxComponent(): boolean {
      //@ts-ignore
      return this.useSelect && this.schema.format === 'radiobox'
    },
    isReadOnly(): boolean | undefined {
      //@ts-ignore
      return this.readonly || this.schema.readonly
    },
    hasDeleteButtonFunction(): boolean {
      //@ts-ignore
      return this.hasDeleteButton && !this.isReadOnly
    },
    titleToShow(): string {
      //@ts-ignore
      return common.getTitle(this.schema.title, this.title)
    },
    options(): {
      value: string | number;
      label: string | number;
    }[] {
      //@ts-ignore
      return common.getOptions(this.schema)
    },
    className(): string {
      const rowClass = this.errorMessage ? this.theme.errorRow : this.theme.row
      return this.schema.className ? rowClass + ' ' + this.schema.className : rowClass
    },
    step(): number | "any" | undefined {
      return common.getNumberStep(this.schema)
    },
  },
  methods: {
    configureSchema() {
      this.schema.configure = true
    },
    onChange(e: { target: { value: string } }) {
      this.value = this.schema.type === 'integer' ? common.toInteger(e.target.value) : common.toNumber(e.target.value)
      this.validate()
      this.emitValue()
    },
    onChangeExpression(expression: string) {
      this.expression = expression
      this.validate()
      this.emitValue()
    },
    updateSelection(value: number) {
      this.value = value
      this.validate()
      this.emitValue()
    },
    validate() {
      if (!!this.expression) {
        this.errorMessage = ''
      }
      this.errorMessage = common.getErrorMessageOfNumber(this.value, this.schema, this.locale)
    },
    emitValue() {
      this.$emit('update-value', { value: this.getAllValue(), isValid: !this.errorMessage })
    },
    getAllValue() {
      if(this.value === undefined) {
        return this.value
      }
      // New format: expression wrapped in {{expr}}, literals are plain values
      // Always return expression wrapper when there's an expression (needed for lookup)
      if (this.expression) {
        return `{{${this.expression}}}`
      }
      if (this.plainStruct) {
        return this.value
      }
      return this.value ?? 0
    },
    getValue(): number {
      // For new {{expr}} format, extract actual value (not the expression wrapper)
      if (typeof this.initialValue === 'string') {
        const match = this.initialValue.match(/^\{\{(.+)\}\}$/)
        if (match) {
          // It's an expression - return 0 so field shows as "defined"
          // The actual expression is stored in this.expression
          return 0
        }
      }
      return common.getDefaultValue(this.required, this.schema, this.initialValue) as number
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
    }
  }
}
</script>
