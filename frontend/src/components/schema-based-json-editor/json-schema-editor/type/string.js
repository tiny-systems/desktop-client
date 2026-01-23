const value = {
    description: null,
    maxLength: null,
    minLength: null,
    pattern: null,
    format:undefined,
    enum:undefined,
    readonly: false,
}
const attr = {
    description: {
      name: 'Description',
      type: 'string',
    },
    maxLength:{
        name:'Max length',
        type:'integer'
    },
    minLength:{
        name:'Min length',
        type:'integer'
    },
    pattern: {
        name: 'Pattern',
        type:'string'
    },
    format: {
        name:'Format',
        type:'array',
        enums:['date', 'date-time', 'textarea', 'email','hostname','ipv4','ipv6','uri', 'checkbox', 'code', 'base64', 'json', 'radiobox']
    },
    enum:{
        name:'Enum',
        type:'array'
    },
    readonly: {
      name: 'Readonly',
      type: 'boolean'
    }
}
const wrapper = {value, attr}
export default wrapper
