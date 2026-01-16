<template>
  <template v-if="tableMode">
    <td v-for="(p, i) in properties">
      <editor
        v-if="isRequired(p.property) !== false"
        :property="p.property"
        :key="p.property + isRequired(p.property)"
        :schema="p.schema"
        :allow-edit-schema="allowEditSchema"
        :getReference="getReference"
        :initial-value="value[p.propertyName]"
        @update-value="onChange(p.propertyName, $event)"
        :theme="theme"
        @lookup="lookup"
        :allow-lookup="allowLookup"
        :plain-struct="plainStruct"
        :locale="locale"
        :required="isRequired(p.property)"
        :readonly="isReadOnly"
        :disable-collapse="disableCollapse"
        :minItemCountIfNeedFilter="minItemCountIfNeedFilter">
      </editor>
    </td>
    <td v-if="hasDeleteButtonFunction">
      <button type="button" class="w-4 text-sky-500 cursor-pointer ml-1"
                @click="$emit('delete')" title="Delete">
      <XCircleIcon></XCircleIcon>
    </button></td>
  </template>
  <div v-else :class="className">
    <div :class="[theme.title, !titleToShow ? 'border-none' : '']">
      <div class="flex space-x-1 w-full">
        <div class="pl-1" v-if="titleToShow != ''">{{ titleToShow }}</div>
        <button v-if="allowEditSchema && schema.configurable && value !== undefined" @click="configureSchema()"
                title="Configure type" type="button"
                class="w-6 h-6 border border-sky-500 dark:border-sky-800  rounded p-1 button inline-block cursor-pointer hover:bg-sky-300 dark:hover:bg-sky-800">
          <PencilIcon></PencilIcon>
        </button>
      </div>
      <optional :required="required"
                :value="value"
                :isReadOnly="isReadOnly"
                :theme="theme"
                :locale="locale"
                @toggleOptional="toggleOptional()">
      </optional>
      <button v-if="hasDeleteButtonFunction" type="button" class="w-4 text-sky-500 cursor-pointer ml-1"
              @click="$emit('delete')" title="Delete">
        <XCircleIcon></XCircleIcon>
      </button>
    </div>
    <div :class="['flex justify-between', deleteHover ? 'bg-red-100 dark:bg-red-900/50 border-0 rounded' : (hover ? 'bg-indigo-100 dark:bg-indigo-800 border-0 rounded' : '')]">
      <div v-if="value !== undefined"
           :class="[theme.card, noBorder ? '' : 'border border-gray-200 dark:border-gray-700', (!!expression ? 'mx-1 rounded-lg' : 'rounded-lg')]">
        <div class="w-full">
          <nav v-if="tabs.length > 0" class="relative z-0 my-2 px-1 justify-between rounded-lg dark:border-gray-600 flex" aria-label="Tabs">
            <a v-for="(p, i) in tabs"  @click.prevent="currentTab = p" :key="i" href="#" :class="['text-gray-500 dark:text-gray-300 rounded-lg', current == p ? 'bg-gray-100 dark:bg-gray-800' : '', 'relative min-w-0 flex-1 overflow-hidden  py-2 px-2 mx-2 text-sm font-medium text-center focus:z-10 whitespace-nowrap']"
               :aria-current="'page'">
              <span>{{ p }}</span>
            </a>
          </nav>
        </div>
        <div :class="['grid grid-flow-row-dense grid-cols-12 w-full p-1']">
          <div v-if="!!expression" class="text-indigo-500 text-xs pb-1 col-span-12">{{ expression }}</div>
          <p class="text-xs col-span-12 text-center p-2 dark:text-gray-500 flex justify-center"
             v-if="properties.length === 0 && !expression">Object is empty. <span v-if="schema.configurable">You can adapt it with your own properties by clicking</span><PencilIcon class="w-3 mx-2 h-3" v-if="schema.configurable"></PencilIcon>
          </p>
            <div v-for="(p, i) in properties"
                 :class="[(tabs.length > 0 ? ( getMerged(p.schema).tab == current ? 'block' : 'hidden') : 'block'),  'break-inside-avoid-column' + (getMerged(p.schema.align) ? ' text-' + getMerged(p.schema).align : '') + (getMerged(p.schema).colSpan ? ' ' + getMerged(p.schema).colSpan : ' col-span-12')]">
              <editor
                v-if="isRequired(p.property) !== false"
                :property="p.property"
                :key="p.property + isRequired(p.property)"
                :schema="p.schema"
                :allow-edit-schema="allowEditSchema"
                :title="p.schema.title || p.propertyName"
                :getReference="getReference"
                :initial-value="value[p.propertyName]"
                @update-value="onChange(p.propertyName, $event)"
                :theme="theme"
                @lookup="lookup"
                :allow-lookup="allowLookup"
                :plain-struct="plainStruct"
                :locale="locale"
                :required="isRequired(p.property)"
                :readonly="isReadOnly"
                :disable-collapse="disableCollapse"
                :minItemCountIfNeedFilter="minItemCountIfNeedFilter">
              </editor>
          </div>
        </div>
      </div>
      <div v-if="allowLookup && value !== undefined">
        <button type="button" @mouseover="hover = true" @mouseleave="hover = false"
                class="w-4 text-indigo-500 cursor-pointer"
                @click="$emit('lookup', getAllValue(), schema, onChangeExpression)" :title="expression ? 'Edit expression' : 'Apply expression'">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
               stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round"
                  d="M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0021 18V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v12a2.25 2.25 0 002.25 2.25z"/>
          </svg>
        </button>
        <button v-if="expression" type="button" class="w-4 text-red-500 cursor-pointer" @mouseover="deleteHover = true" @mouseleave="deleteHover = false" @click="clearExpression" title="Clear expression">
          <XCircleIcon></XCircleIcon>
        </button>
      </div>
    </div>
    <description :theme="theme" :message="schema.description"></description>
    <description :theme="theme" :message="errorMessage"></description>
  </div>
</template>
<script lang="ts">
import type {PropType} from 'vue'
import * as common from './common'
import Optional from './Optional.vue'
import Description from './Description.vue'
import Editor from './Editor.vue'
import {ChevronDownIcon, ChevronRightIcon, PencilIcon, XCircleIcon} from '@heroicons/vue/24/solid'

//@ts-ignore
const getUniqueValues = (array) => (
  array.reduce((accumulator, currentValue) => (
    accumulator.includes(currentValue) ? accumulator : [...accumulator, currentValue]
  ), [])
)

export default {
  emits: ['update-value', 'delete', 'lookup'],
  components: {
    XCircleIcon, PencilIcon,
    ChevronRightIcon,
    ChevronDownIcon,
    optional: Optional,
    description: Description,
    editor: Editor
  },
  props: {
    schema: {
      type: Object as PropType<common.ObjectSchema>,
      required: true,
    },
    allowLookup: Boolean,
    plainStruct: Boolean,
    noBorder: Boolean,
    initialValue: null,
    title: [String, Number],
    getReference: {
      type: Function as unknown as PropType<(ref: string) => common.Schema | undefined>,
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
    tableMode: Boolean,
    allowEditSchema: Boolean,
    readonly: Boolean,
    required: Boolean,
    hasDeleteButton: {
      type: Boolean,
      required: false,
    },
    disableCollapse: Boolean,
    minItemCountIfNeedFilter: Number
  },
  data: () => {
    return {
      value: {} as { [name: string]: common.ValueType } | undefined,
      isAction: false,
      expression: undefined,
      errorMessage: '' as string | undefined,
      filter: '',
      hover: false,
      deleteHover: false,
      invalidProperties: [] as string[],
      properties: [] as { property: string; schema: common.Schema; propertyName: string }[],
      watchedProperties: [] as string[],
      currentTab: '',
    }
  },
  beforeMount() {
    this.value = this.getValue()
    this.expression = this.extractExpression(this.initialValue)
    this.validate()
    for (const property in this.schema.properties) {
      if (this.schema.properties.hasOwnProperty(property)) {
        const schema = this.schema.properties[property]
        const propertyName = schema.propertyName || property
        this.properties.push({
          propertyName,
          property,
          schema
        })
      }
    }
    this.properties.sort(common.compare)

    for (const property in this.schema.properties) {
      const schema = this.schema.properties[property]
      if (schema && schema.requiredWhen && !this.watchedProperties.includes(schema.requiredWhen[0])) {
        this.watchedProperties.push(schema.requiredWhen[0])
      }
    }
    if (this.value !== undefined) {
      this.emitValue()
    }
  },
  computed: {
    tabs(): { property: string; schema: common.Schema; propertyName: string }[] {

      let tabs = []
      for (const p in this.properties) {
        const prop = this.properties[p]
        //@ts-ignore
        if (this.getMerged(prop.schema).tab) {
          // we have ref, resolve
          tabs.push(this.getMerged(prop.schema).tab)
        }
      }
      return getUniqueValues(tabs)
    },
    current(): string {
      // @ts-ignore
      if (this.currentTab) {
        return this.currentTab
      }
      if (this.tabs.length > 0) {
        // @ts-ignore
        return this.tabs[0]
      }
      return ''
    },
    isReadOnly(): boolean | undefined {
      return this.readonly || this.schema.readonly
    },
    hasDeleteButtonFunction(): boolean {
      return this.hasDeleteButton && !this.isReadOnly
    },
    titleToShow(): string {
      if (this.hasDeleteButton) {
        return common.getTitle(common.findTitle(this.value, this.properties), this.title, this.schema.title)
      }
      return common.getTitle(this.schema.title, this.title)
    },

    className(): string {
      const rowClass = this.errorMessage ? this.theme.errorRow : this.theme.row
      return rowClass
    },
  },
  methods: {
    getMerged(schema) {
      if (!schema) {
        return schema
      }
      if (!schema.$ref) {
        return schema
      }
      return Object.assign({}, this.getReference(schema.$ref) || {}, schema)
    },
    configureSchema() {
      this.schema.configure = true
    },
    lookup(data, schema, cb) {
      this.$emit('lookup', data, schema, cb)
    },
    isRequired(property: string) {
      return common.isRequired(this.schema.required, this.value, this.schema, property)
    },
    toggleOptional() {
      // Don't pass initialValue if it's an expression - we don't want the expression string as fallback value
      const fallbackValue = this.extractExpression(this.initialValue) ? undefined : this.initialValue
      this.value = common.toggleOptional(this.value, this.schema, fallbackValue) as {
        [name: string]: common.ValueType
        [name: string]: common.ValueType
      } | undefined
      this.validate()
      this.emitValue()
    },
    onChangeExpression(expression: string) {
      this.expression = expression
      this.emitValue()
    },
    onChange(property: string, {value, isValid, isAction}: common.ValidityValue<common.ValueType>) {
      this.value![property] = value
      this.isAction = isAction
      for (const p in this.schema.properties) {
        if (this.isRequired(p) === false) {
          this.value![p] = undefined
        }
      }
      for (const p in this.value) {
        if (this.schema.properties && !this.schema.properties.hasOwnProperty(p)) {
          delete this.value![p]
        }
      }

      this.validate()
      if (this.watchedProperties.some(p => p === property)) {
        this.$forceUpdate()
      }
      common.recordInvalidPropertiesOfObject(this.invalidProperties, isValid, property)
      this.emitValue()
    },
    onFilterChange(e: { target: { value: string } }) {
      this.filter = e.target.value
    },
    getAllValue() {
      if (this.value === undefined || this.plainStruct) {
        return this.value
      }
      // New format: expression wrapped in {{expr}}, literals are plain values
      if (this.expression) {
        return `{{${this.expression}}}`
      }
      return this.value || {}
    },
    getValue(): common.ValueType {
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
      this.validate()
      this.emitValue()
    },
    emitValue() {
      this.$emit('update-value', {
        value: this.getAllValue(),
        isValid: this.invalidProperties.length === 0,
        isAction: this.isAction
      })
    },
    validate() {
      this.errorMessage = common.getErrorMessageOfObject(this.value, this.schema, this.locale)
    },
  },
  watch: {},
}
</script>
<style scoped>
.col-span-4 {
  grid-column: span 4 / span 4;
}

.col-span-5 {
  grid-column: span 5 / span 5;
}

.col-span-6 {
  grid-column: span 6 / span 6;
}
</style>
