@echo off
ECHO ========================================================
ECHO Open Match + Agones - Full Setup
ECHO ========================================================

minikube status >nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    ECHO Minikube is not running. Starting Minikube...
    minikube start --cpus=4 --memory=8192 --ports 7000-7100:7000-7100/udp
) ELSE (
    ECHO Minikube is already running.
)

helm repo add agones https://agones.dev/chart/stable >nul 2>&1
helm repo add open-match https://open-match.dev/chart/stable >nul 2>&1
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts >nul 2>&1
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts >nul 2>&1
helm repo update

helm upgrade --install prometheus prometheus-community/kube-prometheus-stack ^
  --namespace monitoring ^
  --create-namespace ^
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false ^
  --set grafana.service.type=LoadBalancer ^
  --set grafana.service.port=3000 ^
  --wait

helm upgrade --install jaeger jaegertracing/jaeger ^
    --namespace monitoring ^
    --set provisionDataStore.cassandra=false ^
    --set provisionDataStore.elasticsearch=false ^
    --set storage.type=memory ^
    --set allInOne.enabled=true ^
    --set agent.enabled=false ^
    --set collector.enabled=false ^
    --set query.enabled=false ^
    --wait
kubectl patch svc jaeger -n monitoring -p "{\"spec\": {\"type\": \"LoadBalancer\"}}"


helm upgrade --install agones agones/agones ^
    --namespace agones-system ^
    --create-namespace ^
    --set gameservers.minPort=7000,gameservers.maxPort=7100 ^
    --set agones.metrics.prometheus.enabled=true ^    
    --set agones.controller.autoscaler.syncPeriod=3s ^
    --wait

helm upgrade --install open-match --create-namespace --namespace open-match open-match/open-match ^
    --set open-match-customize.enabled=true ^
    --set open-match-customize.evaluator.enabled=true ^
    --set open-match-override.enabled=true ^
    --set redis.metrics.enabled=true ^
    --set global.telemetry.prometheus.enabled=true ^
    --set global.telemetry.jaeger.enabled=true ^
    --set global.telemetry.jaeger.collectorEndpoint="http://jaeger-collector.monitoring.svc.cluster.local:14268/api/traces" ^
    --set redis.image.tag=latest ^
    --set redis.metrics.image.tag=latest ^
    --wait

kubectl get secret --namespace monitoring prometheus-grafana -o jsonpath="{.data.admin-password}" | ForEach-Object { [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($_)) }

ECHO.
ECHO ========================================================
ECHO Setup Complete!
ECHO ========================================================
pause
