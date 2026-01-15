<template>
  <div :class="className">
    <div :class="theme.title">
      <div class="flex space-x-1 w-full">
        <div class="pl-1" v-if="titleToShow != ''">{{titleToShow}}</div>
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
    <div v-else :class="theme.card">
      <div class="flex justify-between">
        <div v-if="schema.tableMode" class="w-full">
          <table class="w-full table-auto">
            <thead v-if="hasType(realItemsSchema.type, 'object')">
            <tr>
              <th v-for="(prop, idx) in realItemsSchema.properties" class="ucfirst font-semibold text-xs dark:text-gray-500">{{ getTitle(prop.title, idx)}}</th>
            </tr>
            </thead>
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
                      :has-delete-button="true"
                      :disable-collapse="disableCollapse"
                      :minItemCountIfNeedFilter="minItemCountIfNeedFilter">
              </editor>
            </tr>
          </table>
        </div>
        <div v-else class="relative w-full gap-2">
          <div v-for="item in filteredValues" :key="(1 + item.i) * renderSwitch" :data-index="item.i" :class="theme.rowContainer" class="break-inside-avoid-column border-b last:border-b-0 dark:border-gray-800" >
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
                    :has-delete-button="true"
                    :disable-collapse="disableCollapse"
                    :minItemCountIfNeedFilter="minItemCountIfNeedFilter">
            </editor>
          </div>
        </div>

        <div class="whitespace-nowrap flex">
          <button v-if="hasAddButton" @click="addItem()" title="Add" type="button" class="w-4 text-sky-500 cursor-pointer ml-1">
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
import Description from './Description.vue'
import Editor from './Editor.vue'
import { XCircleIcon, DocumentPlusIcon } from '@heroicons/vue/24/outline'
import type {Schema} from "./common";

export default {
  emits: ['update-value', 'delete'],
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
      value: [] as common.ValueType[] | undefined,
      isAction: false,
      errorMessage: '',
      buttonGroupStyle: common.buttonGroupStyleString,
      filter: '',
      invalidIndexes: [] as number[]
    }
  },
  computed: {
    filteredValues(): {p: common.ValueType, i: number}[] {
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
    this.validate()
  },
  methods: {
    toggleOptional() {
      this.value = common.toggleOptional(this.value, this.schema, this.initialValue) as common.ValueType[] | undefined
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
    onChange(i: number, { value, isValid, isAction }: common.ValidityValue<common.ValueType>) {
      this.value![i] = value
      this.validate()
      common.recordInvalidIndexesOfArray(this.invalidIndexes, isValid, i)
      this.emitValue()
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
      this.errorMessage = common.getErrorMessageOfArray(this.value, this.schema, this.locale)
    },
    emitValue() {
      this.$emit('update-value', { value: this.getAllValue(), isValid: !this.errorMessage && this.invalidIndexes.length === 0, isAction: this.isAction })
    },
    getAllValue() {
      if(this.value === undefined || this.plainStruct) {
        return this.value
      }
      return this.value || []
    },
    getInitialValue(): common.ValueType[] {
      return common.getDefaultValue(this.required, this.schema, this.initialValue) as common.ValueType[]
    },
    getTitle(title1: any, title2: any) {
      return common.getTitle(title1, title2)
    },
    getRealSchemaRecursive(s: Schema): Schema {
      if ( s && s.$ref) {
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
    hasType(types: any, check: string) {
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
    XCircleIcon, DocumentPlusIcon,
    optional: Optional,
    description: Description,
    editor: Editor
  }
}
</script>
