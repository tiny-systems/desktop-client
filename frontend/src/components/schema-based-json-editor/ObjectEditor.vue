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
    <div class="flex justify-between">
      <div v-if="value !== undefined"
           :class="[theme.card, noBorder ? '' : 'border', 'rounded-lg dark:border-gray-800']">
        <div class="w-full">
          <nav v-if="tabs.length > 0" class="relative z-0 my-2 px-1 justify-between rounded-lg dark:border-gray-600 flex" aria-label="Tabs">
            <a v-for="(p, i) in tabs" @click.prevent="currentTab = p" :key="i" href="#" :class="['text-gray-500 dark:text-gray-300 rounded-lg', current == p ? 'bg-gray-100 dark:bg-gray-800' : '', 'relative min-w-0 flex-1 overflow-hidden py-2 px-2 mx-2 text-sm font-medium text-center focus:z-10 whitespace-nowrap']"
               :aria-current="'page'">
              <span>{{ p }}</span>
            </a>
          </nav>
        </div>
        <div :class="['grid grid-flow-row-dense grid-cols-12 w-full']">
          <p class="text-xs col-span-12 text-center p-2 dark:text-gray-500 flex justify-center"
             v-if="properties.length === 0">Object is empty.
          </p>
            <div v-for="(p, i) in properties"
                 :class="[(tabs.length > 0 ? ( getMerged(p.schema).tab == current ? 'block' : 'hidden') : 'block'), 'break-inside-avoid-column' + (getMerged(p.schema.align) ? ' text-' + getMerged(p.schema).align : '') + (getMerged(p.schema).colSpan ? ' ' + getMerged(p.schema).colSpan : ' col-span-12')]">
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
import { XCircleIcon } from '@heroicons/vue/24/solid'

const getUniqueValues = (array: any[]) => (
  array.reduce((accumulator, currentValue) => (
    accumulator.includes(currentValue) ? accumulator : [...accumulator, currentValue]
  ), [])
)

export default {
  emits: ['update-value', 'delete'],
  components: {
    XCircleIcon,
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
      errorMessage: '' as string | undefined,
      filter: '',
      invalidProperties: [] as string[],
      properties: [] as { property: string; schema: common.Schema; propertyName: string }[],
      watchedProperties: [] as string[],
      currentTab: '',
    }
  },
  beforeMount() {
    this.value = this.getValue()
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
    tabs(): string[] {
      let tabs: string[] = []
      for (const p in this.properties) {
        const prop = this.properties[p]
        if (this.getMerged(prop.schema).tab) {
          tabs.push(this.getMerged(prop.schema).tab)
        }
      }
      return getUniqueValues(tabs)
    },
    current(): string {
      if (this.currentTab) {
        return this.currentTab
      }
      if (this.tabs.length > 0) {
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
    getMerged(schema: any) {
      if (!schema) {
        return schema
      }
      if (!schema.$ref) {
        return schema
      }
      return Object.assign({}, this.getReference(schema.$ref) || {}, schema)
    },
    isRequired(property: string) {
      return common.isRequired(this.schema.required, this.value, this.schema, property)
    },
    toggleOptional() {
      this.value = common.toggleOptional(this.value, this.schema, this.initialValue) as {
        [name: string]: common.ValueType
      } | undefined
      this.validate()
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
    getAllValue() {
      if (this.value === undefined || this.plainStruct) {
        return this.value
      }
      return this.value || {}
    },
    getValue(): common.ValueType {
      return common.getDefaultValue(this.required, this.schema, this.initialValue)
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
