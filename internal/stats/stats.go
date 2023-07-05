package stats

import (
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/client-go/listers/core/v1"
	"net/http"
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

	// TODO extract needed info
	vkubeNodes := []*corev1.Node{}
	for _, node := range nodes {
		for _, taint := range node.Spec.Taints {
			if taint.Key == "itzloop.dev/virtual-kubelet" {
				vkubeNodes = append(vkubeNodes, node)
			}
		}
	}
	c.JSON(http.StatusOK, vkubeNodes)
}

func (h *StatsHandler) ListPods(c *gin.Context) {
	nodeName := c.Param("node_name")

	pods, err := h.podLister.List(labels.Everything())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO extract needed info
	nodePods := []*corev1.Pod{}
	for _, pod := range pods {
		if pod.Spec.NodeName == nodeName {
			nodePods = append(nodePods, pod)
		}
	}

	c.JSON(http.StatusOK, pods)
}
