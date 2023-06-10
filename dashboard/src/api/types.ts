interface Controller {
  name: string
  host: string
  readiness: boolean
  // devices: []Device ?
}

interface Device {
  name: string
  readiness: boolean
}

interface Pod {
  apiVersion: string
  kind: string
  metadata: {
    annotations: {
      controllerAddress: string
      controllerName: string
    }
    creationTimestamp: string
    name: string
    namespace: string
    resourceVersion: string
    uid: string
  }
  spec: {
    containers: [
      {
        image: string
        imagePullPolicy: string
        name: string
        resources: {}
        terminationMessagePath: string
        terminationMessagePolicy: string
        volumeMounts: [
          {
            mountPath: string
            name: string
            readOnly: true
          }
        ]
      }
    ]
    dnsPolicy: string
    enableServiceLinks: true
    nodeSelector: {
      'kubernetes.io/os': string
      'kubernetes.io/role': string
      type: string
    }
    preemptionPolicy: string
    priority: number
    restartPolicy: string
    schedulerName: string
    securityContext: {}
    serviceAccount: string
    serviceAccountName: string
    terminationGracePeriodSeconds: number
    tolerations: [
      {
        key: string
        operator: string
      },
      {
        effect: string
        key: string
        operator: string
        tolerationSeconds: number
      },
      {
        effect: string
        key: string
        operator: string
        tolerationSeconds: number
      }
    ]
    volumes: [
      {
        name: string
        projected: {
          defaultMode: number
          sources: [
            {
              serviceAccountToken: {
                expirationSeconds: number
                path: string
              }
            },
            {
              configMap: {
                items: [
                  {
                    key: string
                    path: string
                  }
                ]
                name: string
              }
            },
            {
              downwardAPI: {
                items: [
                  {
                    fieldRef: {
                      apiVersion: string
                      fieldPath: string
                    }
                    path: string
                  }
                ]
              }
            }
          ]
        }
      }
    ]
  }
  status: {
    conditions: [
      {
        lastProbeTime: string | undefined
        lastTransitionTime: string | undefined
        message: string
        reason: string
        status: string
        type: string
      }
    ]
    containerStatuses: [
      {
        image: string
        imageID: string
        lastState: {}
        name: string
        ready: false
        restartCount: number
        started: true
        state: {
          running: {
            startedAt: string
          }
        }
      }
    ]
    message: string
    phase: string
    qosClass: string
  }
}

interface Node {
  apiVersion: string
  kind: string
  metadata: {
    annotations: {
      'node.alpha.kubernetes.io/ttl': string
      'virtual-kubelet.io/last-applied-node-status': string
      'virtual-kubelet.io/last-applied-object-meta': string
    }
    creationTimestamp: string
    labels: {
      'alpha.service-controller.kubernetes.io/exclude-balancer': string
      'beta.kubernetes.io/os': string
      'kubernetes.io/hostname': string
      'kubernetes.io/role': string
      type: string
    }
    name: string
    resourceVersion: string
    uid: string
  }
  spec: {
    podCIDR: string
    podCIDRs: string[]
    taints: [
      {
        effect: string
        key: string
        timeAdded?: string
        value: string
      },
    ]
  }
  status: {
    allocatable: {
      cpu: string
      memory: string
      pods: string
    }
    capacity: {
      cpu: string
      memory: string
      pods: string
    }
    conditions: [
      {
        lastHeartbeatTime: string
        lastTransitionTime: string
        message: string
        reason: string
        status: string
        type: string
      },
    ]
    daemonEndpoints: {
      kubeletEndpoint: {
        Port: number
      }
    }
    nodeInfo: {
      architecture: string
      bootID: string
      containerRuntimeVersion: string
      kernelVersion: string
      kubeProxyVersion: string
      kubeletVersion: string
      machineID: string
      operatingSystem: string
      osImage: string
      systemUUID: string
    }
  }
}

export type { Controller, Device, Pod, Node }
