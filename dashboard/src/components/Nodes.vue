<script lang="ts" setup>
import { computed } from '@vue/reactivity'
import { nextTick } from 'vue'
import type { DeepReadonly } from 'vue'
import { watch } from 'vue'
import { reactive } from 'vue'
import { ref } from 'vue'

import axios from 'axios'
import { onMounted } from 'vue'
import router from '@/router'
import API from '@/api/api'
import type { Node } from '@/api/types'
import { onUnmounted } from 'vue'
const search = ref('')
const dialog = ref(false)
const nodes = ref<Map<string, Node>>(new Map())
const alert = ref({
  enable: false,
  title: '',
  msg: ''
})

const headers = ref([
  { title: 'Name', align: 'center', key: 'name' },
  { title: 'Cpu', align: 'center', key: 'cpu' },
  { title: 'Memory', align: 'center', key: 'memory' },
  { title: 'Allocatable Pods', align: 'center', key: 'allocatablePods' },
  { title: 'Max Pods', align: 'center', key: 'maxPods' },
  { title: 'Readiness', align: 'center', key: 'readiness' },
  { title: 'Actions', align: 'center', key: 'actions', sortable: false }
])

const nodeArray = computed(() => {
  const nodeArray = new Array<Node>()
  for (let c of nodes.value) {
    nodeArray.push(c[1])
  }

  return nodeArray
})


function getNodes() {
  API.NodesApi.list()
    .then((res) => {
      for (let c of res) {
        nodes.value.set(c.name, c)
      }
    })
    .catch((err) => {
      alert.value.msg = err
      alert.value.title = 'error'
      alert.value.enable = true
      nodes.value.clear()
    })
}

const intr = setInterval(getNodes, 1000)

onMounted(() => {
  getNodes()
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
    :items="nodeArray"
    :search="search"
    class="elevation-1"
    density="compact"
    item-key="read"
  >
    <template v-slot:top>
      <v-toolbar flat>
        <v-toolbar-title>Nodes</v-toolbar-title>
        <v-divider class="mx-4" inset vertical></v-divider>
        <v-spacer></v-spacer>
        <v-dialog v-model="dialog" max-width="500px">
          <template v-slot:activator="{ props }">
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

    <template v-slot:item.actions="{ item }">
      <v-icon size="small" @click="router.push(`/nodes/${item.raw.name}`)"> mdi-eye </v-icon>
    </template>
  </v-data-table-virtual>
</template>
