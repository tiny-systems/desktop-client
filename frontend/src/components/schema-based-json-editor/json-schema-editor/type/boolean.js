const value = {
    description: null,
    format:undefined
}
const attr = {
    description: {
      name: 'Description',
      type: 'string'
    },
    format: {
      name:'Format',
      type:'array',
      enums:['button', 'select']
    }
}
const wrapper = {value, attr}
export default wrapper
