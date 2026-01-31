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
        <optional :required="required" :value="value" :isReadOnly="isReadOnly"
                  :theme="theme" :locale="locale" @toggleOptional="toggleOptional()">
        </optional>
      </div>
    </div>
    <description :theme="theme" :message="schema.description"></description>
    <div v-if="schema.enum">
      <span v-for="option in options" :key="option.value" :class="theme.checkbox">
      <label :class="theme.label">
        <input type="checkbox"
               :class="theme.checkboxInput"
               @change="onChangeCheckbox(option.value)"
               :checked="isChecked(option.value)"
               :disabled="isReadOnly" />
        {{option.label}}
      </label>
    </span>
    </div>
    <div v-else :class="[theme.card, deleteHover ? 'bg-red-100 dark:bg-red-900/50 border-0 rounded' : (hover ? 'bg-indigo-100 dark:bg-indigo-800 border-0 rounded' : '')]">
      <div :class="['flex justify-between']">
        <table v-if="schema.tableMode" class="w-full table-auto">
          <thead v-if="hasType(realItemsSchema.type, 'object')">
          <tr>
            <th v-for="(prop, idx) in realItemsSchema.properties" class="ucfirst font-semibold text-xs dark:text-gray-500">{{ getTitle(prop.title, idx)}}</th>
          </tr>
          </thead>
          <tr v-if="!!expression">
            <td :colspan="realItemsSchema.properties.length" v-if="!!expression" class="text-indigo-500 text-xs pb-1 pl-1">{{expression}}</td>
          </tr>
          <tr v-for="item in filteredValues" :key="(1 + item.i) * renderSwitch" :data-index="item.i">
            <editor :schema="schema.items"
                    :title="item.i"
                    :getReference="getReference"
                    :initial-value="value[item.i]"
                    @update-value="onChange(item.i, $event)"
                    :theme="theme"
                    :locale="locale"
                    :allow-lookup="allowLookup"
                    :table-mode="true"
                    :plain-struct="plainStruct"
                    :allow-edit-schema="allowEditSchema"
                    :required="true"
                    :readonly="isReadOnly"
                    @delete="onDeleteFunction(item.i)"
                    @lookup="lookup"
                    :has-delete-button="true"
                    :disable-collapse="disableCollapse"
                    :minItemCountIfNeedFilter="minItemCountIfNeedFilter">
            </editor>
          </tr>
        </table>
        <div v-else :class="['relative w-full gap-2']">
          <div v-if="!!expression" class="text-indigo-500 text-xs pb-1 pl-1">{{expression}}</div>
          <div v-for="item in filteredValues" :key="(1 + item.i) * renderSwitch" :data-index="item.i" :class="theme.rowContainer" class="break-inside-avoid-column border-b last:border-b-0 border-gray-100 dark:border-gray-700" >
            <editor :schema="schema.items"
                    :title="item.i"
                    :getReference="getReference"
                    :initial-value="value[item.i]"
                    @update-value="onChange(item.i, $event)"
                    :theme="theme"
                    :locale="locale"
                    :allow-lookup="allowLookup"
                    :plain-struct="plainStruct"
                    :allow-edit-schema="allowEditSchema"
                    :required="true"
                    :readonly="isReadOnly"
                    @delete="onDeleteFunction(item.i)"
                    @lookup="lookup"
                    :has-delete-button="true"
                    :disable-collapse="disableCollapse"
                    :minItemCountIfNeedFilter="minItemCountIfNeedFilter">
            </editor>
          </div>
        </div>

        <div class="whitespace-nowrap flex">
          <button type="button" @mouseover="hover = true" @mouseleave="hover = false"  class="w-4 text-indigo-500 cursor-pointer" v-if="allowLookup && value !== undefined" @click="$emit('lookup', getAllValue(), schema, changeExpression)" :title="expression ? 'Edit expression' : 'Apply expression'">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0021 18V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v12a2.25 2.25 0 002.25 2.25z" />
            </svg>
          </button>
          <button v-if="expression" type="button" class="w-4 text-red-500 cursor-pointer" @mouseover="deleteHover = true" @mouseleave="deleteHover = false" @click="clearExpression" title="Clear expression">
            <XCircleIcon></XCircleIcon>
          </button>
          <button v-if="hasAddButton && !expression" @click="addItem()" title="Add" type="button" class="w-4 text-sky-500 cursor-pointer ml-1">
            <DocumentPlusIcon></DocumentPlusIcon>
          </button>
          <button v-if="hasDeleteButtonFunction" type="button" class="w-4 text-sky-500 cursor-pointer ml-1" @click="$emit('delete')" title="Delete">
            <XCircleIcon></XCircleIcon>
          </button>
        </div>
      </div>
    </div>
    <description :theme="theme" :message="errorMessage" :error="true"></description>
  </div>

</template>
<script lang="ts">
import {isProxy, toRaw} from 'vue'
import type {PropType} from 'vue'
import * as common from './common'

import Optional from './Optional.vue'
import Description  from './Description.vue'
import Editor from './Editor.vue'
import { XCircleIcon, ChevronRightIcon,ChevronDownIcon, DocumentPlusIcon, PencilIcon} from '@heroicons/vue/24/outline'
import type {Schema} from "./common";

export default {
  emits: ['update-value', 'delete', 'lookup'],
  props: {
    schema: {
      type: Object as PropType<common.ArraySchema>,
      required: true,
    },
    allowLookup: Boolean,
    plainStruct: Boolean,
    initialValue: null,
    title: [String, Number],
    getReference: {
      type: Function as unknown as PropType<(name: string) => common.Schema | undefined>,
      required: true,
    },
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
    allowEditSchema: Boolean,
    hasDeleteButton: {
      type: Boolean,
      required: false,
    },
    disableCollapse: Boolean,
    minItemCountIfNeedFilter: Number
  },
  data: () => {
    return {
      renderSwitch: 1,
      hover: false,
      deleteHover: false,
      expression: undefined,
      value: [] as common.ValueType[] | undefined,
      isAction: false,
      errorMessage: '',
      buttonGroupStyle: common.buttonGroupStyleString,
      filter: '',
      invalidIndexes: [] as number[]
    }
  },
  computed: {
    filteredValues(): common.ValueType[] {
      return this.getValue.map((p, i) => ({ p, i }))
        .filter(({ p, i }) => common.filterArray(p, i, this.schema.items, this.filter))
    },
    getValue(): common.ValueType[] {
      if (this.value !== undefined) {
        return this.value
      }
      return []
    },
    realItemsSchema():Schema {
      return this.getRealSchemaRecursive(this.schema.items)
    },
    isReadOnly(): boolean | undefined {
      return this.readonly || this.schema.readonly
    },
    hasDeleteButtonFunction(): boolean {
      return this.hasDeleteButton && !this.isReadOnly
    },
    hasAddButton(): boolean {
      return !this.isReadOnly && this.value !== undefined && !this.schema.enum
    },
    titleToShow(): string {
      return common.getTitle(this.schema.title, this.title)
    },
    className(): string {
      const rowClass = this.errorMessage ? this.theme.errorRow : this.theme.row
      return this.schema.className ? rowClass + ' ' + this.schema.className : rowClass
    },
    options(): { value: string | number, label: string | number }[] {
      return common.getOptions(this.schema)
    },
  },
  beforeMount() {
    this.value = this.getInitialValue()

    if (!Array.isArray(this.value) && this.value) {
      this.value = []
    }
    this.expression = this.extractExpression(this.initialValue)
    this.validate()
  },
  beforeUnmount() {
  },
  mounted() {
  },
  methods: {
    configureSchema() {
      this.schema.configure = true
    },
    //@ts-ignore
    lookup(data, schema, cb) {
      this.$emit('lookup', data, schema, cb)
    },
    toggleOptional() {
      // Don't pass initialValue if it's an expression - we don't want the expression string as fallback value
      const fallbackValue = this.extractExpression(this.initialValue) ? undefined : this.initialValue
      this.value = common.toggleOptional(this.value, this.schema, fallbackValue) as common.ValueType[] | undefined
      this.validate()
      this.emitValue()
    },
    addItem() {
      this.value!.push(common.getDefaultValue(true, this.schema.items, undefined)!)
      this.validate()
      this.emitValue()
    },
    onDeleteFunction(i: number) {
      this.value!.splice(i, 1)
      this.invalidIndexes = []
      this.errorMessage = ''
      this.validate()
      this.renderSwitch = -this.renderSwitch
      this.emitValue()
    },
    changeExpression(expression: string) {
      this.expression = expression
      this.emitValue()
    },
    onChange(i: number, { value, isValid, isAction }: common.ValidityValue<common.ValueType>) {
      this.value![i] = value
      this.validate()
      common.recordInvalidIndexesOfArray(this.invalidIndexes, isValid, i)
      this.emitValue()
    },
    onFilterChange(e: { target: { value: string } }) {
      this.filter = e.target.value
    },
    isChecked(value: any) {
      return this.value && this.value.indexOf(value) !== -1
    },
    onChangeCheckbox(value: any) {
      if (this.value) {
        const index = this.value.indexOf(value)
        if (index !== -1) {
          this.value.splice(index, 1)
        } else {
          this.value.push(value)
        }
        this.validate()
        this.emitValue()
      }
    },
    validate() {
      if (!!this.expression) {
        this.errorMessage = ''
      }
      this.errorMessage = common.getErrorMessageOfArray(this.value, this.schema, this.locale)
    },
    emitValue() {
      this.$emit('update-value', { value: this.getAllValue(), isValid: !this.errorMessage && this.invalidIndexes.length === 0, isAction: this.isAction  })
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
      return this.value || []
    },
    getInitialValue(): common.ValueType[] {
      // For new {{expr}} format, extract actual value (not the expression wrapper)
      if (typeof this.initialValue === 'string') {
        const match = this.initialValue.match(/^\{\{(.+)\}\}$/)
        if (match) {
          // It's an expression - return empty array so field shows as "defined"
          // The actual expression is stored in this.expression
          return []
        }
      }
      return common.getDefaultValue(this.required, this.schema, this.initialValue) as common.ValueType[]
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
    getTitle(title1, title2) {
      return common.getTitle(title1, title2)
    },
    getRealSchemaRecursive(s: Schema): Schema {
      if ( s && s.$ref) {
        // ref overrides defs in some props
        const reference = this.getReference(s.$ref)
        if (reference) {
          if (s.title !== undefined) {
            reference.title = s.title
          }
          if (s.description !== undefined) {
            reference.description = s.description
          }
          if (s.readonly) {
            reference.readonly = s.readonly
          }
          if (s.configurable) {
            reference.configurable = s.configurable
          }
          return this.getRealSchemaRecursive(reference)
        }
      }
      return s
    },
    //@ts-ignore
    hasType(types, check) {

      let val = types

      if (isProxy(val)) {
        val = toRaw(val)
      }
      if (Array.isArray(val)) {
        return val.includes(check)
      }
      return val == check
    }
  },
  components: {
    XCircleIcon, ChevronRightIcon, ChevronDownIcon, DocumentPlusIcon, PencilIcon,
    optional: Optional,
    description: Description,
    editor: Editor
  }
}
</script>
