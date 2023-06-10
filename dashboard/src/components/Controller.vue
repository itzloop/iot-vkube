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
import type { Device } from '@/api/types'
import { readonly } from 'vue'
const controllerName = router.currentRoute.value.params['name'] as string
const search = ref('')
const dialog = ref(false)
const dialogDelete = ref(false)
const editedIndex = ref(-1)
const editedDevice = ref<Device>({
  name: '',
  readiness: false
})
const devices = ref<Map<string, Device>>(new Map())
const alert = ref({
  enable: false,
  title: '',
  msg: ''
})

const headers = ref([
  { title: 'Name', align: 'center', key: 'name' },
  { title: 'Readiness', align: 'center', key: 'readiness' },
  { title: 'Actions', align: 'center', key: 'actions', sortable: false }
])

const deviceArray = computed(() => {
  const deviceArray = new Array<Device>()
  for (let c of devices.value) {
    deviceArray.push(c[1])
  }

  return deviceArray
})

const formTitle = computed(() => {
  editedIndex.value === -1 ? 'New Device' : 'Edit Device'
})

function getDevices() {
  API.DevicesApi.list(controllerName)
    .then((res) => {
      for (let c of res) {
        devices.value.set(c.name, c)
      }
    })
    .catch((err) => {
      alert.value.msg = err
      alert.value.title = 'error'
      alert.value.enable = true
      devices.value.clear()
    })
}

function addOrEditDevice() {
  if (editedIndex.value !== -1) {
    // edit
    API.DevicesApi.update(controllerName, editedDevice.value)
      .then((res) => {
        devices.value.set(editedDevice.value.name, editedDevice.value)
        close()
      })
      .catch((err) => {
        alert.value.msg = err
        alert.value.title = 'error'
        alert.value.enable = true
        close()
      })
    return
  }
  // add

  API.DevicesApi.create(controllerName, editedDevice.value)
    .then((res) => {
      devices.value.set(editedDevice.value.name, editedDevice.value)
      close()
    })
    .catch((err) => {
      alert.value.msg = err
      alert.value.title = 'error'
      alert.value.enable = true
      close()
    })
}

function updateDevice(device: Device) {
  editedDevice.value = Object.assign({}, device)
  editedIndex.value = deviceArray.value.indexOf(device)
  dialog.value = true
  console.log('updateDevice', device)
}

function deleteDevice(device: Device) {
  editedDevice.value = Object.assign({}, device)
  editedIndex.value = deviceArray.value.indexOf(device)
  dialogDelete.value = true
}

function close() {
  dialog.value = false
  nextTick(() => {
    editedIndex.value = -1
    editedDevice.value = {
      name: '',
      readiness: false
    }
  })
}

function closeDelete() {
  dialogDelete.value = false
  nextTick(() => {
    editedIndex.value = -1
    editedDevice.value = {
      name: '',
      readiness: false
    }
  })
}

function deleteDeviceConfirm() {
  const device = deviceArray.value[editedIndex.value]

  if (!device) {
    throw new Error(`no device with index ${editedIndex.value}`)
  }

  API.DevicesApi.delete(controllerName, device)
    .then((res) => {
      devices.value.delete(device.name)
      closeDelete()
    })
    .catch((err) => {
      alert.value.msg = err
      alert.value.title = 'error'
      alert.value.enable = true
      closeDelete()
    })
}

onMounted(() => {
  getDevices()
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
    :items="deviceArray"
    :search="search"
    class="elevation-1"
    density="compact"
    device-key="read"
  >
    <template v-slot:top>
      <v-toolbar flat>
        <v-toolbar-title>Devices</v-toolbar-title>
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
          <v-card>
            <v-card-title>
              <span class="text-h5">{{ formTitle }}</span>
            </v-card-title>

            <v-card-text>
              <v-container>
                <v-row>
                  <v-col cols="12" sm="6" md="4">
                    <v-text-field v-model="editedDevice.name" label="Name"></v-text-field>
                  </v-col>
                  <v-col cols="12" sm="6" md="4">
                    <v-checkbox v-model="editedDevice.readiness" label="Readiness"></v-checkbox>
                  </v-col>
                </v-row>
              </v-container>
            </v-card-text>

            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn color="teal-darken-4" variant="text" @click="close"> Cancel </v-btn>
              <v-btn color="teal-darken-4" variant="text" @click="addOrEditDevice">
                Save
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
        <v-dialog v-model="dialogDelete" max-width="500px">
          <v-card>
            <v-card-title class="text-h5">Are you sure you want to delete this device?</v-card-title>
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn color="teal-darken-4" variant="text" @click="closeDelete">Cancel</v-btn>
              <v-btn color="teal-darken-4" variant="text" @click="deleteDeviceConfirm">OK</v-btn>
              <v-spacer></v-spacer>
            </v-card-actions>
          </v-card>
        </v-dialog>
      </v-toolbar>
    </template>
    <template v-slot:item.readiness="{ item }">
      <v-chip :color="item.columns.readiness ? 'green' : 'red'">
        {{ item.columns.readiness ? 'READY' : 'NOT-READY' }}
      </v-chip>
    </template>

    <template v-slot:item.actions="{ item }">
      <v-icon size="small" class="me-2" @click="updateDevice(item.raw)"> mdi-pencil </v-icon>
      <v-icon size="small" class="me-2" @click="deleteDevice(item.raw)"> mdi-delete </v-icon>
    </template>
  </v-data-table-virtual>
</template>
