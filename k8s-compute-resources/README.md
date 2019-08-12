How do I design this program?

The following metrics will be calculated
Capacity vs. Allocation: https://godoc.org/k8s.io/api/core/v1#NodeStatus
Reservation vs. Utilization: https://kubernetes.io/docs/concepts/policy/resource-quotas/

Cluster Capacity
Cluster Allocatation    (node definitions)
Cluster Utilization     (metrics-server)
Cluster Reservations    (resourcequota)

Node Capacity
Node Allocation         (node definitions)
Node Utilization        (metrics-server)    
Node Reservations       (pod quotas)     


Node Subscription       (% stuff)
Cluster Subscription       (% stuff)

Reserved = (Capacity - Allocation)
Given Allocatation and not Capacity is used for scheduling, is Capacity a useful
metric?
