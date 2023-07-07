package stats

import (
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/client-go/listers/core/v1"
	"net/http"
	"sort"
)

type StatsHandler struct {
	podLister  v1.PodLister
	nodeLister v1.NodeLister
	selector   labels.Selector
}

func NewStatsHandler(podLister v1.PodLister, nodeLister v1.NodeLister, selector labels.Selector, engine *gin.Engine) *StatsHandler {
	sh := &StatsHandler{podLister: podLister, nodeLister: nodeLister, selector: selector}
	sh.AddHandlers(engine)
	return sh
}

func (h *StatsHandler) AddHandlers(r *gin.Engine) {
	r.GET("/stats/nodes", h.ListNodes)
	r.GET("/stats/nodes/:node_name/pods", h.ListPods)
}

func (h *StatsHandler) ListNodes(c *gin.Context) {
	nodes, err := h.nodeLister.List(labels.Everything())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nodeResults := []struct {
		Name            string `json:"name"`
		Cpu             string `json:"cpu"`
		Memory          string `json:"memory"`
		AllocatablePods string `json:"allocatablePods"`
		MaxPods         string `json:"maxPods"`
		Readiness       bool   `json:"readiness"`
	}{}

	for _, node := range nodes {
		var readiness bool
		if len(node.Status.Conditions) > 0 {
			cond := node.Status.Conditions[len(node.Status.Conditions)-1]
			if cond.Type == corev1.NodeReady {
				readiness = cond.Status == corev1.ConditionTrue
			}
		}

		nodeResults = append(nodeResults, struct {
			Name            string `json:"name"`
			Cpu             string `json:"cpu"`
			Memory          string `json:"memory"`
			AllocatablePods string `json:"allocatablePods"`
			MaxPods         string `json:"maxPods"`
			Readiness       bool   `json:"readiness"`
		}{
			Name:            node.Name,
			Cpu:             node.Status.Capacity.Cpu().String(),
			Memory:          node.Status.Capacity.Memory().String(),
			AllocatablePods: node.Status.Allocatable.Pods().String(),
			MaxPods:         node.Status.Capacity.Pods().String(),
			Readiness:       readiness,
		})
	}

	c.JSON(http.StatusOK, nodeResults)
}

func (h *StatsHandler) ListPods(c *gin.Context) {
	nodeName := c.Param("node_name")

	pods, err := h.podLister.List(labels.Everything())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	podResults := []struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Readiness bool   `json:"readiness"`
	}{}

	for _, pod := range pods {
		if pod.Spec.NodeName == nodeName {
			var readiness bool
			if len(pod.Status.Conditions) > 0 {
				cond := pod.Status.Conditions[len(pod.Status.Conditions)-1]
				if cond.Type == corev1.PodReady {
					readiness = cond.Status == corev1.ConditionTrue
				}
			}

			podResults = append(podResults, struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
				Readiness bool   `json:"readiness"`
			}{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				Readiness: readiness,
			})
		}

	}

	sort.Slice(podResults, func(i, j int) bool {
		return podResults[i].Name < podResults[j].Name
	})

	c.JSON(http.StatusOK, podResults)
}
