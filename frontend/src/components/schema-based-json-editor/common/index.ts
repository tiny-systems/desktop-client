// @ts-ignore
import { isObject, isInteger, toNumber, toInteger } from 'lodash'
export const buttonGroupStyleString = "margin-left: 10px";
export const imagePreviewStyleString = "";
export { toNumber, toInteger }
export type Schema = ObjectSchema | ArraySchema | NumberSchema | StringSchema | BooleanSchema | NullSchema | AnySchema
export type ValueType = { [name: string]: any } | any[] | number | boolean | string | null | undefined

type EqualCondition = [string, '===' | 'equal', any]
type InCondition = [string, 'in', any]
type IsUndefinedCondition = [string, 'isUndefined']

export const minItemCountIfNeedFilter = 6

export function isBase64Image(value?: string) {
  if (!value) {
    return false
  }
  return value.indexOf(`data:image/`) === 0
    && value.indexOf(`;base64,`) !== -1
}

export function toggleOptional(value: ValueType | undefined, schema: Schema, initialValue: any) {
  if (value === undefined) {
    return getDefaultValue(true, schema, initialValue)
  } else {
    return undefined
  }
}

export interface CommonSchema {
  $schema?: string;
  $id?: string,
  $ref?: string;
  $defs?: {
    [name: string]: Schema;
  };
  title?: string;
  description?: string;
  default?: ValueType;
  readonly?: boolean;
  propertyOrder?: number | string;
  requiredWhen?: EqualCondition | InCondition | IsUndefinedCondition;
  optionalWhen?: EqualCondition | InCondition | IsUndefinedCondition;
  propertyName?: string;
  oneOf?: [CommonSchema]
  configurable?: boolean
  configure?: boolean
  tab?: string
  className?: string
  colSpan?: string
  align?: string
}

export interface ValidityValue<T> {
  value: T;
  isValid: boolean;
  isAction?: boolean;
}

/**
 * @public
 */
export type AnySchema = CommonSchema & {
  type: undefined;
}

/**
 * @public
 */
export type ObjectSchema = CommonSchema & {
  type: 'object';
  properties: { [name: string]: Schema };
  required?: string[];
  maxProperties?: number;
  minProperties?: number;
  additionalProperties?: boolean;
}

/**
 * @public
 */
export type ArraySchema = CommonSchema & {
  type: 'array';
  items: Schema;
  minItems?: number;
  uniqueItems?: boolean;
  enum?: ValueType[];
  enumTitles?: string[];
  format?: 'select';
  tableMode?: boolean;
}

/**
 * @public
 */
export type NumberSchema = CommonSchema & {
  type: 'number' | 'integer';
  minimum?: number;
  exclusiveMinimum?: number;
  maximum?: number;
  exclusiveMaximum?: number;
  enum?: number[];
  multipleOf?: number;
  enumTitles?: string[];
  format?: 'select' | 'radiobox';
  step?: number | 'any';
}

/**
 * @public
 */
export type StringSchema = CommonSchema & {
  type: 'string';
  language?: string;
  format?: 'textarea' | 'color' | 'date' | 'datetime' | 'date-time' | 'datetime-local' | 'time' | 'month' | 'email' | 'uri' | 'url' | 'week' | 'hostname' | 'ipv4' | 'ipv6' | 'code' | 'base64' | 'radiobox' | 'json';
  enum?: string[];
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  enumTitles?: string[];
  step?: number | 'any';
}

/**
 * @public
 */
export type BooleanSchema = CommonSchema & {
  type: 'boolean';
  format?: 'checkbox' | 'select' | 'button';
}

/**
 * @public
 */
export type NullSchema = CommonSchema & {
  type: 'null';
}

export var defaultLocale = {
  button: {
    collapse: 'Collapse',
    expand: 'Expand',
    add: 'Add',
    delete: 'Delete'
  },
  error: {
    minLength: 'Value must be at least {0} characters long.',
    maxLength: 'Value must be at most {0} characters long.',
    pattern: "Value doesn't match the pattern {0}.",
    minimum: 'Value must be >= {0}.',
    maximum: 'Value must be <= {0}.',
    largerThan: 'Value must be > {0}.',
    smallerThan: 'Value must be < {0}.',
    minItems: 'The length of the array must be >= {0}.',
    uniqueItems: 'The item in {0} and {1} must not be same.',
    multipleOf: 'Value must be multiple value of {0}.',
    minProperties: 'Properties count must be >= {0}.',
    maxProperties: 'Properties count must be <= {0}.',
  },
  info: {
    notExists: 'Not defined',
    true: 'True',
    false: 'False',
    search: 'Search'
  }
};

export type Locale = typeof defaultLocale

export function getLocale(locale: undefined | null | Locale): Locale {
  return locale || defaultLocale
}

export const largeTheme = {
  card: 'w-full',
  row: 'block pb-1',
  rowContainer: '',
  img: 'p-1 mt-1 w-full bg-white rounded-md border shadow-sm dark:bg-gray-900 border-gray-200 dark:border-gray-700',
  errorRow: 'border-red-500',
  button: 'px-2 py-2 cursor-pointer text-md leading-4 font-medium rounded-md text-sky-700 bg-sky-100 enabled:hover:bg-sky-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-sky-500 dark:bg-sky-600 enabled:dark:hover:bg-sky-500 dark:text-black disabled:opacity-50',
  input: 'bg-gray-50 border mx-1 text-sm dark:text-gray-300 rounded-md block w-full p-2 dark:bg-gray-800 dark:placeholder-gray-400 disabled:shadow-none',
  errorInput: 'bg-red-50 mx-1 border border-red-500 text-red-900 placeholder-red-700 text-sm rounded-md focus:ring-red-500 dark:bg-gray-800 focus:border-red-500 block w-full p-2 dark:text-red-500 dark:placeholder-red-500 dark:border-red-500',
  textarea: 'bg-gray-50 mx-1 border text-sm dark:text-gray-300 rounded-md block w-full p-2 dark:bg-gray-800 dark:placeholder-gray-400 dark:focus:ring-blue-500',
  codeEditor: 'bg-gray-50 p-2 border text-sm dark:border-gray-700 dark:text-gray-300 rounded-md block w-full dark:bg-gray-800 dark:placeholder-gray-400 dark:focus:ring-blue-500',
  errorTextarea: 'bg-gray-50 mx-1 border border-red-300 text-gray-900 text-sm rounded-md focus:ring-red-500 focus:border-red-500 block w-full p-2 dark:bg-gray-800 dark:border-red-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-red-500 dark:focus:border-red-500',
  errorCodeEditor: 'bg-gray-50 border p-1 border-red-300 text-gray-900 text-sm rounded-md focus:ring-red-500 focus:border-red-500 block w-full dark:bg-gray-800 dark:border-red-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-red-500 dark:focus:border-red-500',
  checkbox: '',
  checkboxInput: 'w-4 h-4 bg-gray-100 rounded-md',
  radioboxLabel: 'text-sm text-gray-900 dark:text-gray-300',
  radiobox: 'inline-block pr-2',
  radioboxInput: 'text-blue-600 bg-gray-100 border-gray-200 focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-800 dark:border-gray-600',
  buttonGroup: '',
  title: 'block text-sm font-medium text-gray-900 dark:text-gray-300 text-left flex mb-1',
  label: 'block text-sm font-medium text-gray-900 dark:text-gray-300 text-left whitespace-nowrap px-1',
  description: 'block mb-1 px-1 text-sm font-thin font-light text-gray-900 dark:text-gray-300',
  errorDescription: 'block mb-1 px-1 text-sm font-light text-red-900 dark:text-red-400',
  select: 'bg-gray-50 border mx-1 text-sm rounded-md focus:ring-blue-500 block w-full p-2 dark:bg-gray-800 dark:placeholder-gray-400 dark:focus:ring-blue-500 dark:focus:border-blue-500',
  selectError: 'bg-gray-50 border mx-1 text-sm rounded-md focus:ring-blue-500 block w-full p-2 dark:bg-gray-800 dark:placeholder-gray-400 dark:focus:ring-blue-500 dark:focus:border-blue-500',
  expression: 'inline-block border-indigo-500 dark:border-indigo-900 text-sm focus:outline-none text-indigo-500 dark:text-indigo-500',
  staticText: 'border-gray-200 dark:focus:ring-blue-500 focus:ring-blue-500 focus:border-blue-500 dark:border-gray-600 dark:focus:border-blue-500 text-gray-900 dark:text-white'
}

export const smallTheme = {
  card: 'w-full',
  row: 'block pb-1',
  rowContainer: '',
  img: 'p-1 mt-1 w-full bg-white rounded-md border shadow-sm dark:bg-gray-900 border-gray-200 dark:border-gray-700',
  errorRow: 'border-red-500',
  button: 'px-2 py-1.5 cursor-pointer text-xs leading-4 font-medium rounded-md text-sky-700 bg-sky-100 enabled:hover:bg-sky-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-sky-500 dark:bg-sky-600 enabled:dark:hover:bg-sky-500 dark:text-white disabled:opacity-50',
  input: 'bg-gray-50 border text-xs dark:text-gray-300 rounded-md block w-full p-1.5 dark:bg-gray-800 dark:placeholder-gray-400 disabled:shadow-none dark:border-gray-600 border-gray-200 focus:outline-none focus:ring-1 focus:ring-sky-500 focus:border-sky-500',
  errorInput: 'bg-red-50 border border-red-500 text-red-900 placeholder-red-700 text-xs rounded-md focus:ring-red-500 dark:bg-gray-800 focus:border-red-500 block w-full p-1.5 dark:text-red-500 dark:placeholder-red-500 dark:border-red-500 focus:outline-none',
  textarea: 'bg-gray-50 border text-xs dark:text-gray-300 rounded-md block w-full p-1.5 dark:bg-gray-800 dark:placeholder-gray-400 dark:focus:ring-blue-500 dark:border-gray-600 border-gray-200 focus:outline-none focus:ring-1 focus:ring-sky-500 focus:border-sky-500',
  codeEditor: 'bg-gray-50 p-2 border text-xs dark:border-gray-700 dark:text-gray-300 rounded-md block w-full dark:bg-gray-800 dark:placeholder-gray-400 dark:focus:ring-blue-500 focus:outline-none',
  errorTextarea: 'bg-gray-50 border p-1.5 border-red-300 text-gray-900 text-xs rounded-md focus:ring-red-500 focus:border-red-500 block w-full dark:bg-gray-800 dark:border-red-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-red-500 dark:focus:border-red-500 focus:outline-none',
  errorCodeEditor: 'bg-gray-50 border p-1 border-red-300 text-gray-900 text-xs rounded-md focus:ring-red-500 focus:border-red-500 block w-full dark:bg-gray-800 dark:border-red-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-red-500 dark:focus:border-red-500 focus:outline-none',
  checkbox: '',
  checkboxInput: 'w-4 h-4 bg-gray-100 rounded-md focus:outline-none',
  radioboxLabel: 'text-xs text-gray-900 dark:text-gray-300',
  radiobox: 'inline-block pr-2',
  radioboxInput: 'text-blue-600 bg-gray-100 border-gray-200 focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-800 dark:border-gray-600 focus:outline-none',
  buttonGroup: '',
  title: 'block text-xs font-medium text-gray-900 dark:text-gray-300 text-left flex mb-0.5',
  label: 'block text-xs font-medium text-gray-900 dark:text-gray-300 text-left whitespace-nowrap',
  description: 'block mb-1 text-xs font-thin font-light text-gray-500 dark:text-gray-400',
  errorDescription: 'block mb-1 text-xs font-light text-red-900 dark:text-red-400',
  select: 'appearance-none bg-gray-50 border text-xs rounded-md block w-full p-1.5 pr-8 dark:bg-gray-800 dark:placeholder-gray-400 dark:border-gray-600 border-gray-200 dark:text-gray-300 focus:outline-none focus:ring-1 focus:ring-sky-500 focus:border-sky-500 bg-no-repeat bg-[length:16px_16px] bg-[position:right_0.5rem_center] bg-[url("data:image/svg+xml,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' fill=\'none\' viewBox=\'0 0 24 24\' stroke=\'%236b7280\'%3E%3Cpath stroke-linecap=\'round\' stroke-linejoin=\'round\' stroke-width=\'2\' d=\'M19 9l-7 7-7-7\'/%3E%3C/svg%3E")]',
  selectError: 'appearance-none bg-gray-50 border text-xs rounded-md block w-full p-1.5 pr-8 dark:bg-gray-800 dark:placeholder-gray-400 dark:border-gray-600 border-gray-200 focus:outline-none focus:ring-1 focus:ring-red-500 focus:border-red-500 bg-no-repeat bg-[length:16px_16px] bg-[position:right_0.5rem_center] bg-[url("data:image/svg+xml,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' fill=\'none\' viewBox=\'0 0 24 24\' stroke=\'%236b7280\'%3E%3Cpath stroke-linecap=\'round\' stroke-linejoin=\'round\' stroke-width=\'2\' d=\'M19 9l-7 7-7-7\'/%3E%3C/svg%3E")]',
  expression: 'inline-block border-indigo-500 dark:border-indigo-900 text-xs focus:outline-none text-indigo-500 dark:text-indigo-500',
  staticText: 'border-gray-200 dark:focus:ring-blue-500 focus:ring-blue-500 focus:border-blue-500 dark:border-gray-600 dark:focus:border-blue-500 text-gray-900 dark:text-white'
}


export type Theme = typeof smallTheme

export function getTheme(name: string | undefined | Theme): Theme {
  if (typeof name === 'object') {
    return name
  }
  if (name === 'lg' || name === 'large') {
    return largeTheme as Theme
  }
  return smallTheme
}

export function recordInvalidPropertiesOfObject(invalidProperties: string[], isValid: boolean, property: string) {
  const index = invalidProperties.indexOf(property)
  if (isValid) {
    if (index !== -1) {
      invalidProperties.splice(index, 1)
    }
  } else {
    if (index === -1) {
      invalidProperties.push(property)
    }
  }
}

export function getOptions(schema: NumberSchema | StringSchema | ArraySchema) {
  const enumTitles: string[] = schema.enumTitles || []
  return (schema.enum as (number | string)[]).map((e, i) => ({
    value: e,
    label: typeof enumTitles[i] === 'string' ? enumTitles[i] : e
  }))
}

export function recordInvalidIndexesOfArray(invalidIndexes: number[], isValid: boolean, i: number) {
  const index = invalidIndexes.indexOf(i)
  if (isValid) {
    if (index !== -1) {
      invalidIndexes.splice(index, 1)
    }
  } else {
    if (index === -1) {
      invalidIndexes.push(i)
    }
  }
}

export function isSame(value1: ValueType, value2: ValueType) {
  if (typeof value1 === 'string'
    || typeof value1 === 'number'
    || typeof value1 === 'boolean'
    || value1 === null
    || value1 === undefined) {
    return value1 === value2
  }
  if (typeof value2 === 'string'
    || typeof value2 === 'number'
    || typeof value2 === 'boolean'
    || value2 === null
    || value2 === undefined) {
    return false
  }
  if (Array.isArray(value1)) {
    if (Array.isArray(value2) && (value1 as ValueType[]).length === (value2 as ValueType[]).length) {
      for (let i = 0; i < (value1 as ValueType[]).length; i++) {
        if (!isSame((value1 as ValueType[])[i], (value2 as ValueType[])[i])) {
          return false
        }
      }
      return true
    }
    return false
  }
  if (Array.isArray(value2)
    || Object.keys((value1 as { [name: string]: ValueType })).length !== Object.keys((value2 as { [name: string]: ValueType })).length) {
    return false
  }
  for (const key in value1) {
    if (value1.hasOwnProperty(key) && !isSame((value1 as { [name: string]: ValueType })[key], (value2 as { [name: string]: ValueType })[key])) {
      return false
    }
  }
  return true
}

export function getDefaultValue(required: boolean | undefined = undefined, schema: Schema, initialValue: ValueType | undefined): ValueType | undefined {
  if (initialValue !== undefined) {

    switch (schema.type) {
      case 'object':
        if (isObject(initialValue)) {
          // @ts-ignore
          if (initialValue.hasOwnProperty('value') && isObject(initialValue.value)) {
            // @ts-ignore
            return initialValue.value
          }
          return initialValue || undefined
        }
        break
      case 'array':
        if (isObject(initialValue) && initialValue.hasOwnProperty('value')) {
          // @ts-ignore
          return initialValue.value
        }
        if (Array.isArray(initialValue)) {
          return initialValue
        }
        break
      case 'number':
      case 'integer':
        if (isObject(initialValue)) {
          // @ts-ignore
          return initialValue.value || null
        }
        if (typeof initialValue === 'number') {
          return initialValue
        }
        break
      case 'boolean':
        if (isObject(initialValue)) {
          // @ts-ignore
          return initialValue.value || null
        }
        if (typeof initialValue === 'boolean') {
          return initialValue
        }
        break
      case 'string':
        if (isObject(initialValue)) {
          // @ts-ignore
          return initialValue.value || null
        }
        if (typeof initialValue === 'string') {
          return initialValue
        }
        break
      case undefined:
        if (isObject(initialValue)) {
          if (initialValue.hasOwnProperty('value')) {
            // @ts-ignore
            return initialValue.value
          }
        }
        return initialValue
      case 'null':
      default:
        if (initialValue === null) {
          return initialValue
        }
    }
  }

  if (!required) {
    return undefined
  }

  if (schema.default !== undefined) {
    switch (schema.type) {
      case 'object':
        if (isObject(schema.default)) {
          return schema.default
        }
        break
      case 'array':
        if (Array.isArray(schema.default)) {
          return schema.default
        }
        break
      case 'number':
      case 'integer':
        if (typeof schema.default === 'number') {
          return schema.default
        }
        break
      case 'boolean':
        if (typeof schema.default === 'boolean') {
          return schema.default
        }
        break
      case 'string':
        if (typeof schema.default === 'string') {
          return schema.default
        }
        break
      case 'null':
      default:
        if (schema.default === null) {
          return schema.default
        }
    }
  }

  switch (schema.type) {
    case 'object':
      const value: any = {}
      for (const property in schema.properties) {
        const propertySchema = schema.properties[property]
        value[propertySchema.propertyName || property] = undefined
      }
      return value
    case 'array':
      return []
    case 'number':
    case 'integer':
      if (schema.enum !== undefined && schema.enum.length > 0) {
        return schema.enum[0]
      } else {
        return 0
      }
    case 'boolean':
      return false
    case 'string':
      if (schema.enum !== undefined && schema.enum.length > 0) {
        return schema.enum[0]
      } else {
        return ''
      }
    case 'null':
    default:
      return null
  }
}

export function getErrorMessageOfArray(value: any[] | undefined, schema: ArraySchema, locale: Locale) {
  if (value !== undefined) {
    if (schema.minItems !== undefined && value.length < schema.minItems) {
      return locale.error.minItems.replace('{0}', String(schema.minItems))
    }
    if (schema.uniqueItems) {
      for (let i = 1; i < value.length; i++) {
        for (let j = 0; j < i; j++) {
          if (isSame(value[i], value[j])) {
            return locale.error.uniqueItems.replace('{0}', String(j)).replace('{1}', String(i))
          }
        }
      }
    }
  }
  return ''
}

export function getErrorMessageOfString(value: string | undefined, schema: StringSchema, required: boolean, locale: Locale) {
  if (value !== undefined && value !== null) {
    if (schema.minLength !== undefined
      && value.length < schema.minLength) {
      return locale.error.minLength.replace('{0}', String(schema.minLength))
    }
    if (schema.maxLength !== undefined
      && value.length > schema.maxLength) {
      return locale.error.maxLength.replace('{0}', String(schema.maxLength))
    }
    if (schema.pattern !== undefined
      && !new RegExp(schema.pattern).test(value)) {
      return locale.error.pattern.replace('{0}', String(schema.pattern))
    }
  }
  if (value === undefined && required) {
    return 'Field is required';
  }
  return ''
}

export function getErrorMessageOfNumber(value: number | undefined, schema: NumberSchema, locale: Locale) {
  if (value !== undefined) {
    if (schema.minimum !== undefined) {
      if (schema.exclusiveMinimum != undefined) {
        if (value <= schema.exclusiveMinimum) {
          return locale.error.largerThan.replace('{0}', String(schema.exclusiveMinimum))
        }
      } else {
        if (value < schema.minimum) {
          return locale.error.minimum.replace('{0}', String(schema.minimum))
        }
      }
    }
    if (schema.maximum !== undefined) {
      if (schema.exclusiveMaximum != undefined) {
        if (value >= schema.exclusiveMaximum) {
          return locale.error.smallerThan.replace('{0}', String(schema.exclusiveMaximum))
        }
      } else {
        if (value > schema.maximum) {
          return locale.error.maximum.replace('{0}', String(schema.maximum))
        }
      }
    }
    if (schema.multipleOf && schema.multipleOf > 0 && !isInteger(value / schema.multipleOf)) {
      return locale.error.multipleOf.replace('{0}', String(schema.multipleOf))
    }
  }
  return ''
}

export function getErrorMessageOfObject(value: { [name: string]: ValueType } | undefined, schema: ObjectSchema, locale: Locale) {
  if (value !== undefined) {
    let length = 0
    for (const key in value) {
      if (value.hasOwnProperty(key) && value[key] !== undefined) {
        length++
      }
    }
    if (schema.minProperties !== undefined
      && length < schema.minProperties) {
      return locale.error.minProperties.replace('{0}', String(schema.minProperties))
    }
    if (schema.maxProperties !== undefined
      && length > schema.maxProperties) {
      return locale.error.maxProperties.replace('{0}', String(schema.maxProperties))
    }
  }
  return ''
}

export function isRequired(
  required: string[] | undefined,
  value: { [name: string]: ValueType } | undefined,
  schema: ObjectSchema,
  property: string) {
  /**
   * return true: required
   * return undefined: optional
   * return false: hidden
   */
  if (required && required.some(r => r === property)) {
    return true
  }
  if (value && schema && schema.properties && schema.properties[property]) {
    const requiredWhen = schema.properties[property].requiredWhen
    if (requiredWhen) {
      const [left, operator] = requiredWhen
      const right = requiredWhen[2]
      if (schema.properties[left]) {
        let valueLeft = value[left]
        // @ts-ignore
        valueLeft = isObject(valueLeft) && valueLeft.hasOwnProperty("value") ? valueLeft.value : valueLeft

        if (operator === '===' || operator === 'equal') {
          return valueLeft === right
        }
        if (operator === 'in') {
          return Array.isArray(right) && right.indexOf(valueLeft) !== -1
        }
        if (operator === 'isUndefined') {
          return valueLeft === undefined
        }
      }
    }

    const optionalWhen = schema.properties[property].optionalWhen
    if (optionalWhen) {
      const [left, operator] = optionalWhen
      const right = optionalWhen[2]
      if (schema.properties[left]) {
        let valueLeft = value[left]
        // @ts-ignore
        valueLeft = isObject(valueLeft) && valueLeft.hasOwnProperty("value") ? valueLeft.value : valueLeft

        if (operator === '===' || operator === 'equal') {
          return valueLeft === right ? undefined : false
        }
        if (operator === 'in') {
          return Array.isArray(right) && right.indexOf(valueLeft) !== -1 ? undefined : false
        }
        if (operator === 'isUndefined') {
          return valueLeft === undefined ? undefined : false
        }
      }
    }
  }
  return undefined
}

export function findTitle(value: { [name: string]: ValueType } | undefined, properties: { property: string; schema: Schema }[]) {
  if (value) {
    for (const { property, schema } of properties || []) {
      const title = value[property]
      if ((schema.type === 'number'
        || schema.type === 'integer'
        || schema.type === 'string')
        && schema.enum && schema.enumTitles) {
        const index = (schema.enum as (string | number)[]).indexOf(title as string | number)
        if (index !== -1 && index < schema.enumTitles.length) {
          const enumTitle = schema.enumTitles[index]
          if (typeof enumTitle === 'string' && enumTitle.length > 0) {
            if (enumTitle.length > 33) {
              return enumTitle.substring(0, 30) + '...'
            }
            return enumTitle
          }
        }
      }
      if (typeof title === 'string' && title.length > 0) {
        if (title.length > 33) {
          return title.substring(0, 30) + '...'
        }
        return title
      }
    }
  }
  return undefined
}

function findTitleFromSchema(value: { [name: string]: ValueType } | undefined, schema: ObjectSchema) {
  if (value) {
    for (const property in schema.properties) {
      if (schema.properties.hasOwnProperty(property)) {
        const title = value[property]
        if (typeof title === 'string' && title.length > 0) {
          if (title.length > 33) {
            return title.substring(0, 30) + '...'
          }
          return title
        }
      }
    }
  }
  return undefined
}

/**
 * @public
 */
export function getTitle(...titles: any[]) {
  for (const title of titles) {
    if (title === undefined || title == null || title == "") {
      continue
    }
    return String(title)
  }
  return ''
}

export function compare(a: { property: string; schema: Schema }, b: { property: string; schema: Schema }) {
  //@ts-ignore
  return (a.schema.propertyOrder || 0) - (b.schema.propertyOrder || 0)
}

/**
 * @public
 */
export function filterObject({ property, schema }: { property: string; schema: Schema }, filterValue: string): boolean {
  return filterValue === ''
    || property.indexOf(filterValue) !== -1
    || (!!schema.title && schema.title.indexOf(filterValue) !== -1)
    || (!!schema.description && schema.description.indexOf(filterValue) !== -1)
}

/**
 * @public
 */
export function filterArray(value: ValueType, index: number, schema: Schema, filterValue: string): boolean {
  const result = filterValue === ''
    || String(index).indexOf(filterValue) !== -1
    || (schema.type === 'string' && (value as string).indexOf(filterValue) !== -1)
    || ((schema.type === 'number' || schema.type === 'integer') && String(value as number).indexOf(filterValue) !== -1)
  if (result) {
    return true
  }
  if (schema.type === 'object') {
    const title = getTitle(findTitleFromSchema(value as { [name: string]: ValueType }, schema), schema.title)
    return title.indexOf(filterValue) !== -1
  }
  return false
}


const imageExtensions = ['.png', '.jpg', '.bmp', '.gif']

/**
 * @public
 */
export function isImageUrl(value?: string) {
  if (!value || value.length <= 'https://'.length) {
    return false
  }
  if (value.substr(0, 'http://'.length) !== 'http://'
    && value.substr(0, 'https://'.length) !== 'https://') {
    return false
  }
  const extensionName = value.substr(value.length - 4, 4)
  return imageExtensions.indexOf(extensionName) !== -1
}

export function getNumberStep(schema: NumberSchema) {
  if (schema.step !== undefined) {
    return schema.step
  }
  return schema.type === 'number' ? 'any' : undefined
}
