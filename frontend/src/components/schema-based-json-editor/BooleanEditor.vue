<template>
  <div :class="className">
    <div :class="theme.title">
      <div class="flex space-x-1 w-full">
        <div class="pl-1" v-if="titleToShow != '' && schema.format != 'button'">{{titleToShow}}</div>
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
        <button v-if="hasDeleteButtonFunction" type="button" class="w-4 text-sky-500 cursor-pointer ml-1" @click="$emit('delete')" title="Delete">
          <XCircleIcon></XCircleIcon>
        </button>
      </div>
    </div>
    <div v-if="value !== undefined" :class="['flex justify-between mx-1', deleteHover ? 'bg-red-100 dark:bg-red-900/50 border-0 rounded' : (hover ? 'bg-indigo-100 dark:bg-indigo-800 border-0 rounded' : '')]">
      <div class="w-full">
        <!-- add expression support here -->
        <label :class="theme.label" v-if="schema.format === 'checkbox'">
          <input type="checkbox"
                 :class="[theme.checkboxInput, !!expression ? theme.expression : theme.staticText]"
                 @change="onChange()"
                 :checked="!!expression ? true : value"
                 :disabled="isReadOnly || !!expression" />
          {{expression || locale.info.true}}
        </label>
        <select v-else-if="schema.format === 'select'" :class="[theme.select, !!expression ? theme.expression : theme.staticText]"
                :value="!!expression ? null :value"
                :disabled="isReadOnly || !!expression"
                @change="onChange()">
          <option :value="true" v-if="!expression">{{locale.info.true}}</option>
          <option :value="false" v-if="!expression">{{locale.info.false}}</option>
          <option :value="null"  v-if="!!expression">{{expression}}</option>
        </select>
        <template v-else-if="schema.format === 'button'">
          <button type="button" :class="[theme.button]" :value="value.toString()" @click="emitAction">{{titleToShow}}</button>
        </template>
        <template v-else>
          <span :class="theme.radiobox" v-if="!!expression">
            <label :class="theme.expression">
              <input type="radio" :checked="true" :readonly="true" />
              {{expression}}
            </label>
          </span>
          <span :class="theme.radiobox" v-if="!expression">
            <label :class="theme.radioboxLabel">
              <input type="radio"
                     @change="onChange()"
                     :checked="value"
                     :value="true"
                     :class="theme.radioboxInput"
                     :disabled="isReadOnly" />
              {{locale.info.true}}
            </label>
          </span>
          <span :class="theme.radiobox" v-if="!expression">
            <label :class="theme.radioboxLabel">
              <input type="radio"
                     @change="onChange()"
                     :checked="!value"
                     :value="false"
                     :disabled="isReadOnly" />
                {{locale.info.false}}
            </label>
          </span>
        </template>
      </div>
      <button type="button" class="w-4 text-indigo-500 cursor-pointer" @mouseover="hover = true" @mouseleave="hover = false" v-if="allowLookup && value !== undefined && !isReadOnly" @click="$emit('lookup', getAllValue(), schema, onChangeExpression)" :title="expression ? 'Edit expression' : 'Apply expression'">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" >
          <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0021 18V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v12a2.25 2.25 0 002.25 2.25z" />
        </svg>
      </button>
      <button v-if="expression && !isReadOnly" type="button" class="w-4 text-red-500 cursor-pointer" @mouseover="deleteHover = true" @mouseleave="deleteHover = false" @click="clearExpression" title="Clear expression">
        <XCircleIcon></XCircleIcon>
      </button>
    </div>
    <description :theme="theme" :message="schema.description"></description>
  </div>

</template>
<script lang="ts">
import type { PropType } from 'vue'
import Optional from './Optional.vue'
import Description  from './Description.vue'
import * as common from './common'
import { XCircleIcon, PencilIcon } from '@heroicons/vue/24/solid'

export default {
  emits: ['update-value', 'delete', 'lookup'],
  components: {
    XCircleIcon, PencilIcon,
    optional: Optional,
    description: Description
  },
  props: {
    schema: {
      type: Object as PropType<common.BooleanSchema>,
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
      required: false,
    },
  },
  data: () => {
    return {
      expression: undefined,
      hover: false as boolean,
      deleteHover: false as boolean,
      value: false as boolean | undefined,
      buttonGroupStyle: common.buttonGroupStyleString,
    }
  },
  beforeMount() {
    //@ts-ignore
    this.value = this.getValue() ?? false
    //@ts-ignore
    this.expression = this.extractExpression(this.initialValue)
    //@ts-ignore
    this.emitValue()
  },
  computed: {
    isReadOnly(): boolean | undefined {
      //@ts-ignore
      return this.readonly || this.schema.readonly
    },
    hasDeleteButtonFunction(): boolean {
      //@ts-ignore
      return this.hasDeleteButton && !this.isReadOnly
    },
    titleToShow(): string {
      return common.getTitle(this.schema.title, this.title)
    },
    className(): string {
      const rowClass = this.theme.row
      return this.schema.className ? rowClass + ' ' + this.schema.className : rowClass
    },
  },
  methods: {
    onChange() {
      //@ts-ignore
      this.value = !this.value
      this.emitValue()
    },
    configureSchema() {
      //@ts-ignore
      this.schema.configure = true
    },
    onChangeExpression(expression: string) {
      //@ts-ignore
      this.expression = expression
      this.emitValue()
    },
    emitValue() {
      //@ts-ignore
      this.$emit('update-value', { value: this.getAllValue(), isValid: true })
    },
    emitAction() {
      //@ts-ignore
      this.value = true
      //@ts-ignore
      this.$emit('update-value', { value: this.getAllValue(), isValid: true, isAction: true })
    },
    //@ts-ignore
    getAllValue() {
      //@ts-ignore
      if(this.value === undefined) {
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
      return !!this.value
    },
    getValue(): boolean {
      // For new {{expr}} format, extract actual value (not the expression wrapper)
      if (typeof this.initialValue === 'string') {
        const match = this.initialValue.match(/^\{\{(.+)\}\}$/)
        if (match) {
          // It's an expression - return false so field shows as "defined"
          // The actual expression is stored in this.expression
          return false
        }
      }
      return common.getDefaultValue(this.required, this.schema, this.initialValue) as boolean
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
  }
}
</script>
