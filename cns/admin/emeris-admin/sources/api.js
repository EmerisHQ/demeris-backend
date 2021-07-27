import axios from 'axios'
import config from '../nuxt.config'

export default axios.create({
  baseURL: config.axios.apiUrl
});