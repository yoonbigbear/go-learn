@echo off
echo ========================================================
echo ??? Open Match + Agones FULL RESET & SETUP
echo ========================================================

REM ---------------------------------------------------------
REM 1. ?? ?? (Open Match & Agones)
REM ---------------------------------------------------------
echo ?? Cleaning up Open Match...
kubectl delete namespace open-match --ignore-not-found=true

echo ?? Cleaning up Agones...
kubectl delete namespace agones-system --ignore-not-found=true

echo ? Waiting for clean up (15s)...
timeout /t 15 /nobreak >nul

REM ---------------------------------------------------------
REM 2. [...](asc_slot://start-slot-1)Open Match ??
REM ---------------------------------------------------------
echo ?? Installing Open Match Core...
kubectl create namespace open-match
kubectl apply --namespace open-match -f https://open-match.dev/install/v1.8.0/yaml/01-open-match-core.yaml

echo ?? Installing Default Evaluator...
kubectl apply --namespace open-match -f https://open-match.dev/install/v1.8.0/yaml/06-open-match-override-configmap.yaml -f https://open-match.dev/install/v1.8.0/yaml/07-open-match-default-evaluator.yaml

REM ---------------------------------------------------------
REM 3. Redis ?? & ?? ??
REM ---------------------------------------------------------
echo ?? Patching Redis (Bitnami + No Password)...
timeout /t 5 /nobreak >nul
REM ???? Bitnami? ???, ???? ?? ??? ??(ALLOW_EMPTY_PASSWORD)
kubectl set image statefulset/open-match-redis-master -n open-match redis=docker.io/bitnami/redis:latest metrics=docker.io/bitnami/redis-exporter:latest
kubectl set env statefulset/open-match-redis-master -n open-match ALLOW_EMPTY_PASSWORD=yes

echo ?? Rebooting Open Match Pods...
kubectl delete pod --all -n open-match

REM ---------------------------------------------------------
REM 4. [...](asc_slot://start-slot-3)Agones ??
REM ---------------------------------------------------------
echo ?? Installing Agones...
kubectl create namespace agones-system
kubectl apply --server-side -f https://raw.githubusercontent.com/googleforgames/agones/release-1.38.0/install/yaml/install.yaml

echo ? Waiting for Agones Controller...
kubectl wait --for=condition=available --timeout=120s deployment/agones-controller -n agones-system

echo ========================================================
echo ? ALL SYSTEMS GO!
echo ========================================================
pause
