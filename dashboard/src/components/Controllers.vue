<script lang="ts" setup>
import { computed } from '@vue/reactivity'
import { nextTick } from 'vue'
import { watch } from 'vue'
import { reactive } from 'vue'
import { ref } from 'vue'

import axios from 'axios'
import { onMounted } from 'vue'
import router from '@/router'

type Controller = {
  name: string
  readiness: boolean
}

const controllers = ref<Map<string, Controller>>(new Map())

const alert = ref({
  enable: false,
  title: '',
  msg: ''
})

function getControllers() {
  axios('http://localhost:5000/controllers', {
    method: 'get'
  })
    .then((res) => {
      const cs: Array<Controller> = res.data
      for (let c of cs) {
        controllers.value.set(c.name, c)
      }
    })
    .catch((err) => {
      alert.value.msg = err
      alert.value.title = 'error'
      alert.value.enable = true
      controllers.value.clear()
    })
}

onMounted(() => {
  getControllers()
})
</script>

<template>
  <v-alert
    v-model="alert.enable"
    border="start"
    variant="tonal"
    closable
    close-label="Close Alert"
    type="error"
    title="alert.title"
  >
    {{ alert.msg }}
  </v-alert>
  <v-table v-if="controllers.size > 0" title="Controller">
    <thead>
      <tr>
        <th class="text-left">Name</th>
        <th class="text-left">Readiness</th>
      </tr>
    </thead>
    <tbody>
      <tr @click="router.push(`/controllers/${item[1].name}`)" v-for="item in controllers" :key="item[1].name">
        <td>{{ item[1].name }}</td>
        <td>{{ item[1].readiness }}</td>
      </tr>
    </tbody>
  </v-table>
  <h1 v-else>No Controllers</h1>
  <VBtn color="teal-darken-4">Add</VBtn
  ><VBtn @click="getControllers" color="teal-darken-4">Refresh</VBtn>
</template>
