# vigilant

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go](https://img.shields.io/badge/Go-1.11-blue.svg)](#)
[![Version](https://images.microbadger.com/badges/version/bincyber/vigilant.svg)](https://microbadger.com/images/bincyber/vigilant)
[![Layers](https://images.microbadger.com/badges/image/bincyber/vigilant.svg)](https://microbadger.com/images/bincyber/vigilant)
[![CircleCI](https://circleci.com/gh/bincyber/vigilant.svg?style=svg)](https://circleci.com/gh/bincyber/vigilant)


_vigilant_ is a Kubernetes security controller.

## What It Does

_vigilant_ ensures the following for every Namespace in your Kubernetes cluster:

1. the Namespace has the label `name`

```
$ kubectl get namespaces --show-labels

NAME          STATUS   AGE     LABELS
default       Active   2m42s   name=default
kube-public   Active   2m42s   name=kube-public
kube-system   Active   2m42s   name=kube-system
```

This faciliates the use of the `namespaceSelector` in [NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) objects.

For example, this NetworkPolicy can be applied without having to manually add the label `name=web-app` to the `web-app` Namespace:

```
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-web-app
  namespace: pgsql
spec:
  policyTypes:
  - Ingress
  podSelector:
    matchLabels:
      app: postgres-10
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: web-app
```


2. the Namespace has a default NetworkPolicy that denies all ingress and egress traffic:

```
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: example
spec:
  policyTypes:
  - Ingress
  - Egress
  podSelector: {}
  ingress: []
  egress:
  - ports:
    - port: 53
      protocol: TCP
    - port: 53
      protocol: UDP
```

This policy will apply to all Pods in the Namespace and only permit outbound DNS traffic.


## How It Works

_vigilant_ is a [DecoratorController](https://metacontroller.app/api/decoratorcontroller/).

It is registered with the [metacontroller](https://github.com/GoogleCloudPlatform/metacontroller) and watches for the creation of Namespace objects. When a new namespace is created, the metacontroller sends a POST request to _vigilant_ at its `/sync` endpoint. _vigilant_ returns the `name` label and NetworkPolicy to add to the namespace which is done by the metacontroller.

```
$ kubectl -n metacontroller logs metacontroller-0 --tail=12

I0124 21:47:31.633272       1 controller.go:423] DecoratorController knsc: sync Namespace /kube-system
I0124 21:47:31.636532       1 controller.go:423] DecoratorController knsc: sync Namespace /default
I0124 21:47:31.638269       1 controller.go:423] DecoratorController knsc: sync Namespace /kube-public
I0124 21:47:31.638274       1 controller.go:508] DecoratorController knsc: updating Namespace /kube-system
I0124 21:47:31.642925       1 controller.go:508] DecoratorController knsc: updating Namespace /default
I0124 21:47:31.644319       1 controller.go:508] DecoratorController knsc: updating Namespace /kube-public
I0124 21:47:31.646605       1 manage_children.go:246] Namespace kube-system: creating NetworkPolicy kube-system/default-deny-all
I0124 21:47:31.646648       1 manage_children.go:246] Namespace default: creating NetworkPolicy default/default-deny-all
I0124 21:47:31.647014       1 manage_children.go:246] Namespace kube-public: creating NetworkPolicy kube-public/default-deny-all
I0124 21:47:31.653321       1 controller.go:423] DecoratorController knsc: sync Namespace /metacontroller
I0124 21:47:31.655149       1 controller.go:508] DecoratorController knsc: updating Namespace /metacontroller
I0124 21:47:31.658274       1 manage_children.go:246] Namespace metacontroller: creating NetworkPolicy metacontroller/default-deny-all
```

## Prerequisites

_vigilant_ requires the [metacontroller](https://github.com/GoogleCloudPlatform/metacontroller) add-on running in your Kubernetes cluster.


## Usage

Deploy _vigilant_:
```
$ kubectl apply -f https://raw.githubusercontent.com/bincyber/vigilant/master/manifests/deployment.yaml
```

Register the DecoratorController with the metacontroller:
```
$ kubectl apply -f https://github.com/bincyber/vigilant/blob/master/manifests/decoratorcontroller.yaml
```

Verify that namespaces have had a `name` label added to them:
```
$ kubectl get namespaces --show-labels
```

Verify that a NetworkPolicy has been added to each namespace:
```
$ kubectl get networkpolicy --all-namespaces
```
