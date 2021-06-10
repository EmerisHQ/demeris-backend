import axios from 'axios'

export default axios.create({
  baseURL: process.env.CNS_URL || "http://localhost:9999"
})