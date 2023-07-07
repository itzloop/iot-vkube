<script lang="ts" setup>
import { computed } from '@vue/reactivity'
import { nextTick, onUnmounted } from 'vue'
import type { DeepReadonly } from 'vue'
import { watch } from 'vue'
import { reactive } from 'vue'
import { ref } from 'vue'

import axios from 'axios'
import { onMounted } from 'vue'
import router from '@/router'
import API from '@/api/api'
import type { Pod } from '@/api/types'
import { readonly } from 'vue'
const nodeName = router.currentRoute.value.params['name'] as string
const search = ref('')
const dialog = ref(false)
const dialogDelete = ref(false)
const editedIndex = ref(-1)
const pods = ref<Map<string, Pod>>(new Map())
const alert = ref({
  enable: false,
  title: '',
  msg: ''
})

const headers = ref([
  { title: 'Name', align: 'center', key: 'name' },
  { title: 'Namespace', align: 'center', key: 'namespace' },
  { title: 'Readiness', align: 'center', key: 'readiness' },
])

const podArray = computed(() => {
  const podArray = new Array<Pod>()
  for (let c of pods.value) {
    podArray.push(c[1])
  }

  return podArray
})

const formTitle = computed(() => {
  editedIndex.value === -1 ? 'New Pod' : 'Edit Pod'
})

function getPods() {
  API.PodsApi.list(nodeName)
    .then((res) => {
      for (let c of res) {
        pods.value.set(c.name, c)
      }
    })
    .catch((err) => {
      alert.value.msg = err
      alert.value.title = 'error'
      alert.value.enable = true
      pods.value.clear()
    })
}


const intr = setInterval(getPods, 1000)

onMounted(() => {
  getPods()
})

onUnmounted(() => {
    clearInterval(intr)
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
    :title="alert.title"
  >
    {{ alert.msg }}
  </v-alert>

  <v-data-table-virtual
    :headers="headers"
    :items="podArray"
    :search="search"
    class="elevation-1"
    density="compact"
    pod-key="read"
  >
    <template v-slot:top>
      <v-toolbar flat>
        <v-toolbar-title>Pods</v-toolbar-title>
        <v-divider class="mx-4" inset vertical></v-divider>
        <v-spacer></v-spacer>
        <v-dialog v-model="dialog" max-width="500px">
          <template v-slot:activator="{ props }">
            <v-btn color="teal-darken-4" dark class="mb-2" v-bind="props"> Add </v-btn>
            <v-card color="grey-lighten-3" min-width="200">
              <v-card-text>
                <v-text-field
                  v-model="search"
                  append-inner-icon="mdi-magnify"
                  density="compact"
                  label="Search"
                  single-line
                  hide-details
                ></v-text-field>
              </v-card-text>
            </v-card>
          </template>
        </v-dialog>
      </v-toolbar>
    </template>
    <template v-slot:item.readiness="{ item }">
      <v-chip :color="item.columns.readiness ? 'green' : 'red'">
        {{ item.columns.readiness ? 'READY' : 'NOT-READY' }}
      </v-chip>
    </template>
  </v-data-table-virtual>
</template>
