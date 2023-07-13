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
import type { Controller } from '@/api/types'
import { readonly } from 'vue'
const search = ref('')
const dialog = ref(false)
const dialogDelete = ref(false)
const editedIndex = ref(-1)
const editedController = ref<Controller>({
  name: '',
  host: '',
  readiness: false
})
const controllers = ref<Map<string, Controller>>(new Map())
const alert = ref({
  enable: false,
  title: '',
  msg: ''
})

const headers = ref([
  { title: 'Name', align: 'center', key: 'name' },
  { title: 'Host', align: 'center', key: 'host' },
  { title: 'Readiness', align: 'center', key: 'readiness' },
  { title: 'Actions', align: 'center', key: 'actions', sortable: false }
])

const controllerArray = computed(() => {
  const controllerArray = new Array<Controller>()
  for (let c of controllers.value) {
    controllerArray.push(c[1])
  }

  return controllerArray
})

const formTitle = computed(() => {
  editedIndex.value === -1 ? 'New Item' : 'Edit Item'
})

function getControllers() {
  API.ControllersApi.list()
    .then((res) => {
      for (let c of res) {
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

function addOrEditController() {
  if (editedIndex.value !== -1) {
    // get old name
    const oldName = controllerArray.value[editedIndex.value].name

    // edit
    API.ControllersApi.update(oldName, editedController.value)
      .then((res) => {
        if (oldName !== editedController.value.name) {
          controllers.value.delete(oldName)
        }

        controllers.value.set(editedController.value.name, editedController.value)
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

  API.ControllersApi.create(editedController.value)
    .then((res) => {
      controllers.value.set(editedController.value.name, editedController.value)
      close()
    })
    .catch((err) => {
      alert.value.msg = err
      alert.value.title = 'error'
      alert.value.enable = true
      close()
    })
}

function updateController(controller: Controller) {
  editedController.value = Object.assign({}, controller)
  editedIndex.value = controllerArray.value.indexOf(controller)
  dialog.value = true
}

function deleteController(controller: Controller) {
  editedController.value = Object.assign({}, controller)
  editedIndex.value = controllerArray.value.indexOf(controller)
  dialogDelete.value = true
}

function close() {
  dialog.value = false
  nextTick(() => {
    editedIndex.value = -1
    editedController.value = {
      name: '',
      host: '',
      readiness: false
    }
  })
}

function closeDelete() {
  dialogDelete.value = false
  nextTick(() => {
    editedIndex.value = -1
    editedController.value = {
      name: '',
      host: '',
      readiness: false
    }
  })
}

function deleteItemConfirm() {
  const controller = controllerArray.value[editedIndex.value]

  if (!controller) {
    throw new Error(`no controller with index ${editedIndex.value}`)
  }

  API.ControllersApi.delete(controller)
    .then((res) => {
      controllers.value.delete(controller.name)
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
    :title="alert.title"
  >
    {{ alert.msg }}
  </v-alert>

  <v-data-table-virtual
    :headers="headers"
    :items="controllerArray"
    :search="search"
    class="elevation-1"
    density="compact"
    item-key="read"
  >
    <template v-slot:top>
      <v-toolbar flat>
        <v-toolbar-title>Controllers</v-toolbar-title>
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
                    <v-text-field v-model="editedController.name" label="Name"></v-text-field>
                  </v-col>
                  <v-col cols="12" sm="6" md="4">
                    <v-text-field v-model="editedController.host" label="Host"></v-text-field>
                  </v-col>
                  <v-col cols="12" sm="6" md="4">
                    <v-checkbox v-model="editedController.readiness" label="Readiness"></v-checkbox>
                  </v-col>
                </v-row>
              </v-container>
            </v-card-text>

            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn color="teal-darken-4" variant="text" @click="close"> Cancel </v-btn>
              <v-btn color="teal-darken-4" variant="text" @click="addOrEditController">
                Save
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
        <v-dialog v-model="dialogDelete" max-width="500px">
          <v-card>
            <v-card-title class="text-h5">Are you sure you want to delete this item?</v-card-title>
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn color="teal-darken-4" variant="text" @click="closeDelete">Cancel</v-btn>
              <v-btn color="teal-darken-4" variant="text" @click="deleteItemConfirm">OK</v-btn>
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
      <v-icon size="small" class="me-2" @click="updateController(item.raw)"> mdi-pencil </v-icon>
      <v-icon size="small" class="me-2" @click="deleteController(item.raw)"> mdi-delete </v-icon>
      <v-icon size="small" @click="router.push(`/controllers/${item.raw.name}`)"> mdi-eye </v-icon>
    </template>
  </v-data-table-virtual>
</template>
