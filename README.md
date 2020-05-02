# Wallclock Controller

I have included multiple ways of building and deploying the controller (see below). Once everything is deployed, you can view the timezones with:

```bash
kubectl get timezones.projectjudge.k8s.io
```

and associated wallclocks with:

```bash
watch -n1 kubectl get wallclocks.projectjudge.k8s.io -A -o=custom-columns=NAME:metadata.name,TIME:status.time
```

## Option 1. Docker build and kubectl deploy

```bash
make
```

Or

```bash
bash __build__/scripts/simple_deploy.sh
```

## Option 2. Skaffold

This will deploy the controller into a local kubernetes cluster (default context)

### Prerequisites

- skaffold
- helmenv
- helm

### Running

```bash
make deploy
```

## Option 3. Local build

If you don't have any of the prerequsites installed, you can run the following to deploying the wallclock controller

```bash
kubectl apply -f __build__/k8s/wallclock-operator/crds/crd.yaml
go run . -kubeconfig=$HOME/.kube/config

# In a new terminal window
kubectl apply -f __build__/k8s/wallclocks/templates/timezones.yaml
```

## Option 4. Local deploy of pre-compiled code (macOS Catalina 10.15.4)

If you don't have any of the prerequsites installed, you can run the following to deploying the wallclock controller

```bash
kubectl apply -f __build__/k8s/wallclock-operator/crds/crd.yaml
./wallclock-controller -kubeconfig=$HOME/.kube/config

# In a new terminal window
kubectl apply -f __build__/k8s/wallclocks/templates/timezones.yaml
```

## Future work

- Handling cleanups
- Performance optimisation
- Multiple controller support
